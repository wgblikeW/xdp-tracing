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
	InstallStrategy RPCType = iota
	RevokeStrategy
)

const (
	KEEP_ALIVE_TIMEOUT = 65
)

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
	Ctx       context.Context
	IPAddr    string
	NodeID    string
	policyGen PolicyController
}

var (
	remoteHost remoteHostCache                  = remoteHostCache{Storage: make(map[string]string)} // Take local records for cluster's nodes
	connCache  map[string]*grpc.ClientConn      = make(map[string]*grpc.ClientConn)                 // Reuse the long-alive connection
	creds      credentials.TransportCredentials                                                     // Credentials used for TLS connection to the gRPC server
	slowQueue  chan string                      = make(chan string, 10)
)

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
	tContro.mu.Lock()
	defer tContro.mu.Unlock()
	newRules := "172.17.0.1"
	tContro.Append(newRules)
	SendRPCToPeers(ctx, InstallStrategy, tContro)
}

func (tContro *testPolicyContro) ToByte() []byte {
	tContro.mu.Lock()
	defer tContro.mu.Unlock()
	return []byte(tContro.Read())
}

func SendRPCToPeers(ctx context.Context, rpcType RPCType, policyGen PolicyController) {
	var goFunc func(params *sendRPCParams)

	// Choosing different Handle Function for Sending RPC
	switch rpcType {
	case InstallStrategy:
		goFunc = sendInstallStrategyRPC
	case RevokeStrategy:
		goFunc = sendRevokeStrategyRPC
	}

	remoteHost.mu.Lock()
	for nodeID, IPAddr := range remoteHost.Storage {
		go goFunc(&sendRPCParams{Ctx: ctx, IPAddr: IPAddr, NodeID: nodeID, policyGen: policyGen})
	}
	remoteHost.mu.Unlock()
}

// makeTLSConfiguration return crendentials for TLS Connection based on client key and client's
// certificate, it will also load CA certificate to verify the server certificate during
// TLS handshake
func makeTLSConfiguration() credentials.TransportCredentials {
	// TODO: Move "credsRootPath" reading from Config file
	credsRootPath := "../x509/"
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
		ServerName:   "server.grpc.io",
		RootCAs:      certPool,
	})

	return creds
}

func main() {
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
	creds = makeTLSConfiguration()

	var testGen PolicyController = &testPolicyContro{}
	go testGen.Generate(ctx)
	go nodeWatcher(ctx)
	go retryConn()
}

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
			slowQueue <- ipAddr
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

func retryConn() {
	for ipAddr := range slowQueue {
		time.Sleep(time.Second * 5)
		makeClient(ipAddr)
	}
}

func sendInstallStrategyRPC(params *sendRPCParams) {
	c, err := makeClient(params.IPAddr)
	if err != nil {
		logrus.Warnf("[Policy Controller] error occurs when making Client err=", err.Error())
		return
	}
	// Contact the server and print out its response.
	ctxT, cancel := context.WithTimeout(params.Ctx, time.Second*1)
	defer cancel()

	r, err := c.InstallStrategy(ctxT, &strategy.UpdateStrategy{Blockoutrules: params.policyGen.ToByte()})
	if err != nil {
		logrus.Warnf("[Policy Controller] error occurs when sending RPC to %v err=%v", params.IPAddr, err)
		// Slowly Retry sending InstallStrategyRPC to remote server
		go func() {
			var sleepTime time.Duration = 1
			for {
				sleepTime = sleepTime + 1 // Slightly increase the sleepTime as every failure tryings
				time.Sleep(time.Second * sleepTime)

				if _, exists := remoteHost.Storage[params.NodeID]; exists {
					r, err := c.InstallStrategy(ctxT, &strategy.UpdateStrategy{Blockoutrules: params.policyGen.ToByte()})
					if err != nil {
						logrus.Warnf("[Policy Controller] error occurs when retrying sending RPC to %v err=%v", params.IPAddr, err)
						continue
					}
					logrus.Infof("[Policy Controller] response from %v Status:%v", params.NodeID, r.Status)
					return
				} else {
					return
				}
			}
		}()
	}
	logrus.Infof("[Policy Controller] response from %v status=%v", params.IPAddr, r.Status)
}

func sendRevokeStrategyRPC(params *sendRPCParams) {

}

func nodeWatcher(ctx context.Context) {
	var etcdService *service.EtcdService = service.NewEtcdService(ctx)
	if err := etcdService.Conn(); err != nil {
		logrus.Fatalf("[etcd Service] failed to start etcd componet err=%v", err.Error())
	}

	watchCh := etcdService.Client.Watch(ctx, "node", clientv3.WithPrefix())

	go func() {
		for watch := range watchCh {
			for _, event := range watch.Events {
				remoteHost.mu.Lock()
				switch event.Type {

				case clientv3.EventTypeDelete:
					// Node Server Instance Disconnnected from clusters
					delete(remoteHost.Storage, string(event.Kv.Key)) // remove disconnected server from localcache
					remoteHost.mu.Unlock()
					logrus.Warn("[Policy Controller] remote server disconnected %v", string(event.Kv.Key))
				case clientv3.EventTypePut:
					// New Node Server Instance Connected to clusters
					remoteHost.Storage[string(event.Kv.Key)] = string(event.Kv.Value)
					// TODO: Sending InstallStrategyRPC to new Node
					remoteHost.mu.Unlock()
					logrus.Warn("[Policy Controller] remote server connected %v", string(event.Kv.Key))

				default:
					remoteHost.mu.Unlock()
				}
			}

			select {
			case <-etcdService.StopCh:
				return
			default:
			}
		}
	}()
}
