package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/p1nant0m/xdp-tracing/service"
	"github.com/p1nant0m/xdp-tracing/service/strategy"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

type RPCType uint32

const (
	SERVER_NAME = "server.grpc.io"
)

const (
	InstallStrategy RPCType = iota
	RevokeStrategy
)

const (
	KEEP_ALIVE_TIMEOUT = 65
)

type Retry struct {
	RPCParams   *sendRPCParams
	RPCType     RPCType
	Reason      string
	RetryPeriod time.Duration
}

type PolicyController interface {
	Read() string
	Append(string)
	Generate(context.Context) // Generate() will trace new comming policy and trigger [InstallStrategyRPC] Or [RevokeStrategyRPC]
	ToByte() []byte
}

type remoteHostCache struct {
	Storage map[string]string
	mu      sync.Mutex
}

type sendRPCParams struct {
	Ctx        context.Context
	IPAddr     string
	NodeID     string
	Policy     []byte
	RetryTimes int
}

var (
	remoteHost remoteHostCache                  = remoteHostCache{Storage: make(map[string]string)} // Take local records for cluster's nodes
	connCache  map[string]*grpc.ClientConn      = make(map[string]*grpc.ClientConn)                 // Reuse the long-alive connection
	creds      credentials.TransportCredentials                                                     // Credentials used for TLS connection to the gRPC server
	recycleCh  chan *Retry                      = make(chan *Retry, 10)
)

type PolicyControFromRest struct {
	mu              sync.Mutex
	BootstrapServer string
	Policy          []string
}

func (rContro *PolicyControFromRest) Read() string {
	rContro.mu.Lock()
	defer rContro.mu.Unlock()
	return strings.Join(rContro.Policy, " ")
}

func (rContro *PolicyControFromRest) Append(policy string) {
	rContro.mu.Lock()
	defer rContro.mu.Unlock()
	rContro.Policy = append(rContro.Policy, policy)
}

func (rContro *PolicyControFromRest) Generate(ctx context.Context) {
	// This should using Read() and Append() to operate Policy safely
	/** TODO: Request RestAPI for the newest judgement about the Application Data Flow
	Make Policy based on the prediction whether an action was malicious or not **/
}

func (rContro *PolicyControFromRest) ToByte() []byte {
	// This should using Read() and Append() to operate Policy safely
	return []byte(rContro.Read())
}

// testPolicyController uses for testing gRPC configuration and whether eBPF agent
// can apply policy to eBPF kernel program
type testPolicyContro struct {
	mu     sync.Mutex // using for protecting the operation in Policy
	Policy []string
}

func (tContro *testPolicyContro) Read() string {
	tContro.mu.Lock()
	defer tContro.mu.Unlock()
	return strings.Join(tContro.Policy, " ")
}

func (tContro *testPolicyContro) Append(policy string) {
	tContro.mu.Lock()
	defer tContro.mu.Unlock()
	tContro.Policy = append(tContro.Policy, policy)
}

func (tContro *testPolicyContro) Generate(ctx context.Context) {
	// This should using Read() and Append() to operate Policy safely
	newRules := "172.17.0.11"
	tContro.Append(newRules)
	go SendRPCToPeers(ctx, RevokeStrategy, []byte(newRules))
}

func (tContro *testPolicyContro) ToByte() []byte {
	// This should using Read() and Append() to operate Policy safely
	return []byte(tContro.Read())
}

// SendRPCToPeers will call registed RPC method based on input "rpcTpye" to every node in the cluster
func SendRPCToPeers(ctx context.Context, rpcType RPCType, policy []byte) {
	var goFunc func(*sendRPCParams, chan<- *Retry)

	// Choosing different Handle Function for Sending RPC
	switch rpcType {
	case InstallStrategy:
		goFunc = sendInstallStrategyRPC
	case RevokeStrategy:
		goFunc = sendRevokeStrategyRPC
	}

	remoteHost.mu.Lock()
	for nodeID, IPAddr := range remoteHost.Storage {
		go goFunc(&sendRPCParams{Ctx: ctx, IPAddr: IPAddr, NodeID: nodeID, Policy: policy}, recycleCh)
	}
	remoteHost.mu.Unlock()
}

func slowlyRetry(ctx context.Context) {
	for retryEvent := range recycleCh {
		go func(retryEvent *Retry) {
			time.Sleep(time.Second * 3) // Wait for network recover or key node:.* expire
			remoteHost.mu.Lock()
			if _, exists := remoteHost.Storage[retryEvent.RPCParams.NodeID]; !exists {
				// Retry will be invalid since the remote host is disconnected
				remoteHost.mu.Unlock()
				return
			}
			remoteHost.mu.Unlock()
			switch retryEvent.RPCType {
			case InstallStrategy:
				logrus.Warnf("[Policy Controller] Retrying InstallStrategyRPC %v:%v",
					retryEvent.RPCParams.NodeID, retryEvent.RPCParams.IPAddr)
				sendInstallStrategyRPC(retryEvent.RPCParams, recycleCh)
			case RevokeStrategy:
				logrus.Warnf("[Policy Controller] Retrying RevokeStrategyRPC %v:%v",
					retryEvent.RPCParams.NodeID, retryEvent.RPCParams.IPAddr)
				sendRevokeStrategyRPC(retryEvent.RPCParams, recycleCh)
			}
		}(retryEvent)
	}
}

// makeTLSConfiguration return crendentials for TLS Connection based on client key and client's
// certificate, it will also load CA certificate to verify the server certificate during
// TLS handshake
func makeTLSConfiguration(credsPath string) credentials.TransportCredentials {
	credsRootPath := credsPath
	certificate, err := tls.LoadX509KeyPair(credsRootPath+"client.crt", credsRootPath+"client.key")
	if err != nil {
		logrus.Fatal("[grpc Client] error occurs when loading x509 key-pair err=", err.Error())
	}

	certPool := x509.NewCertPool()
	ca, err := ioutil.ReadFile(credsRootPath + "ca.crt")
	if err != nil {
		logrus.Fatal("[grpc Client] error occurs when ReadingFile ca.crt", "err=", err.Error())
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		log.Fatal("failed to append ca certs")
	}

	creds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{certificate},
		ServerName:   SERVER_NAME,
		RootCAs:      certPool,
	})

	return creds
}

func main() {
	if len(os.Args) < 3 {
		logrus.Fatal("./policy-controller [required <path of config.yml>]" +
			" [required <path of credentials>]")
	}

	err := service.ReadAndParseConfig(os.Args[1])
	if err != nil {
		logrus.Fatal("[Policy Controller] error occurs when ReadAndParseConfig", "err=", err.Error())
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := make(chan os.Signal, 1)
	signal.Notify(watcher, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-watcher
		// OS Signal Catched, exit the program gracefully
		cancel()
	}()

	// Make TLS Configuration for gRPC Client
	creds = makeTLSConfiguration(os.Args[2])
	var testGen PolicyController = &testPolicyContro{}

	nodeWatcher(ctx, testGen) // this goroutine trace the modification of cluster nodes, and sync the cluster policy
	go testGen.Generate(ctx)  // this goroutine used for receiving new policy instrcution
	go slowlyRetry(ctx)       // this goroutine will retired the failure RPCs until it is success or remote host be removed

	<-ctx.Done()
}

// makeClient makes a new client on new connection or reused the established connection
func makeClient(ipAddr string) (strategy.StrategyClient, error) {
	var (
		conn   *grpc.ClientConn
		err    error
		exists bool
	)

	// Firstly Check whether there is a long-alive connection for the client
	if conn, exists = connCache[ipAddr]; exists && conn.GetState() == connectivity.Connecting {
		// Cache Connection can be reused
		goto out
	} else {
		conn, err = grpc.Dial(ipAddr, grpc.WithTransportCredentials(creds),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{Timeout: KEEP_ALIVE_TIMEOUT}))
		if err != nil {
			// Slowly Retry making connection between client and server
			return nil, fmt.Errorf("did not connect:%v", err.Error())
		}
		// Closing the old Connection if it exists
		if _, exists = connCache[ipAddr]; exists {
			connCache[ipAddr].Close()
		}
		// old connection will be GC
		connCache[ipAddr] = conn
	}

out:
	c := strategy.NewStrategyClient(conn)
	return c, nil
}

// sendInstallStrategyRPC sends InstallStrategy to gRPC Server Endpoint, it takes remote address,
// nodeID and policy cotroller as input. No matter what happen (network partition, network failures or
// remote server crash) it should ensure the every call will be treaded properly
// (in the end it will go to remote server or remote server leave the cluster)
func sendInstallStrategyRPC(params *sendRPCParams, recycle chan<- *Retry) {
	logrus.Infof("[Policy Controller] trying sending InstallStrategyRPC to %v", params.IPAddr)
	c, err := makeClient(params.IPAddr)
	if err != nil {
		logrus.Warnf("[Policy Controller] error occurs when making Client err=", err.Error())
		recycle <- &Retry{RPCParams: params, RPCType: InstallStrategy, Reason: err.Error()}
		return
	}
	// Contact the server and print out its response.
	ctxT, cancel := context.WithTimeout(params.Ctx, time.Second*1)
	defer cancel()

	r, err := c.InstallStrategy(ctxT, &strategy.UpdateStrategy{Blockoutrules: params.Policy})
	if err != nil {
		recycle <- &Retry{RPCParams: params, RPCType: InstallStrategy, Reason: err.Error()}
		logrus.Warnf("[Policy Controller] error occurs when sending InstallStrategyRPC to %v err=%v", params.IPAddr, err)
		return
	}
	logrus.Infof("[Policy Controller] response from %v status=%v", params.IPAddr, r.Status)
}

func sendRevokeStrategyRPC(params *sendRPCParams, recycle chan<- *Retry) {
	logrus.Infof("[Policy Controller] trying sending RevokeStrategyRPC to %v", params.IPAddr)
	c, err := makeClient(params.IPAddr)
	if err != nil {
		logrus.Warnf("[Policy Controller] error occurs when making Client err=", err.Error())
		recycle <- &Retry{RPCParams: params, RPCType: RevokeStrategy, Reason: err.Error()}
		return
	}
	// Contact the server and print out its response.
	ctxT, cancel := context.WithTimeout(params.Ctx, time.Second*1)
	defer cancel()

	r, err := c.RevokeStrategy(ctxT, &strategy.UpdateStrategy{Blockoutrules: params.Policy})
	if err != nil {
		recycle <- &Retry{RPCParams: params, RPCType: RevokeStrategy, Reason: err.Error()}
		logrus.Warnf("[Policy Controller] error occurs when sending RevokeStrategyRPC to %v err=%v", params.IPAddr, err)
		return
	}
	logrus.Infof("[Policy Controller] response from %v status=%v", params.IPAddr, r.Status)
}

// nodeWatcher will connect to Etcd Server which used for Service Discovery and setup a wather in
// observating the changes of cluster's node. If there is a new node join in the cluster, it should be
// applied the strategy that any other nodes has, (syn process) and if a node leave the cluster, we should
// keep things go right
func nodeWatcher(ctx context.Context, policyGen PolicyController) {
	var etcdService *service.EtcdService = service.NewEtcdService(ctx)
	if err := etcdService.Conn(); err != nil {
		logrus.Fatalf("[etcd Service] failed to start etcd componet err=%v", err.Error())
	}

	// Retrieve cluster's node that has already registed
	kvs, err := etcdService.Client.Get(ctx, "node", clientv3.WithPrefix())
	if err != nil {
		logrus.Fatal("[etcd Service] error occurs when trying to get key with prefix [node]")
	}
	for _, kv := range kvs.Kvs {
		logrus.Info("[Policy Controller] Cluster already has node before Policy Controller Start", "node=", string(kv.Key))
		remoteHost.Storage[string(kv.Key)] = string(kv.Value)
	}
	watchCh := etcdService.Client.Watch(ctx, "node", clientv3.WithPrefix())

	go func() {
		logrus.Info("[Policy Controller] Successfully Start Etcd Service and Setup Wather on key with prefix [node]")
		for watch := range watchCh {
			for _, event := range watch.Events {
				remoteHost.mu.Lock()
				switch event.Type {

				case clientv3.EventTypeDelete:
					// Node Server Instance Disconnnected from clusters
					delete(remoteHost.Storage, string(event.Kv.Key)) // remove disconnected server from localcache
					remoteHost.mu.Unlock()
					logrus.Warnf("[Policy Controller] remote server disconnected %v", string(event.Kv.Key))

				case clientv3.EventTypePut:
					// New Node Server Instance Connected to clusters
					remoteHost.Storage[string(event.Kv.Key)] = string(event.Kv.Value)
					logrus.Warnf("[Policy Controller] remote server connected %v", string(event.Kv.Key))
					// Send Single RPC to new node to sync policy across the cluster
					sendInstallStrategyRPC(&sendRPCParams{
						NodeID: string(event.Kv.Key),
						IPAddr: string(event.Kv.Value),
						Ctx:    ctx,
						Policy: policyGen.ToByte()}, recycleCh)

					remoteHost.mu.Unlock()
				default:
					logrus.Warnf("[Policy Controller] not intention type received %v", event.Type.String())
					remoteHost.mu.Unlock()
				}
			}

			select {
			case <-etcdService.StopCh:
				return
			case <-ctx.Done():
				return
			default:
			}
		}
	}()
}
