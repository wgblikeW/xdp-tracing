package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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
	"google.golang.org/grpc/credentials"
)

type PolicyGenerator interface {
	Read() string
	Append(string)
	Generate(context.Context) // Generate() will trace new comming policy and trigger [InstallStrategyRPC] Or [RevokeStrategyRPC]
	ToByte() []byte
}

type remoteHostCache struct {
	Storage map[string]string
	mu      sync.Mutex
}

var (
	remoteHost remoteHostCache = remoteHostCache{Storage: make(map[string]string)}
	creds      credentials.TransportCredentials
)

type testPolicyGen struct {
	Policy []string
}

func (tGen *testPolicyGen) Read() string {
	return strings.Join(tGen.Policy, " ")
}

func (tGen *testPolicyGen) Append(policy string) {
	tGen.Policy = append(tGen.Policy, policy)
}

func (tGen *testPolicyGen) Generate(ctx context.Context) {
	newRules := "172.17.0.1"
	tGen.Append(newRules)
	for nodeID, IPAddr := range remoteHost.Storage {
		go sendInstallStrategyRPC(ctx, IPAddr, nodeID, tGen) //TODO: Implement Failure Dection and retry latter
	}
}

func (tGen *testPolicyGen) ToByte() []byte {
	return []byte(tGen.Read())
}

func makeTLSConfiguration() credentials.TransportCredentials {
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

	var testGen PolicyGenerator = &testPolicyGen{}
	go testGen.Generate(ctx)
	go nodeWatcher(ctx)
}

func makeClient(ipAddr string) strategy.StrategyClient {
	conn, err := grpc.Dial(ipAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	conn.GetState() // TODO: Keep-alive for long connection, using timeout to close(), when reuse the connection, check it state before using
	defer conn.Close()

	c := strategy.NewStrategyClient(conn)
	return c
}

func sendInstallStrategyRPC(ctx context.Context, ipAddr string, nodeID string, policyGen PolicyGenerator) {
	c := makeClient(ipAddr)

	// Contact the server and print out its response.
	ctxT, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()

	r, err := c.InstallStrategy(ctxT, &strategy.UpdateStrategy{Blockoutrules: policyGen.ToByte()})
	if err != nil {
		logrus.Fatalf("[Policy Controller] error occurs when sending RPC to %v err=%v", ipAddr, err)
	}
	logrus.Infof("[Policy Controller] response from %v status=%v", ipAddr, r.Status)
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

				case clientv3.EventTypePut:
					// New Node Server Instance Connected to clusters
					remoteHost.Storage[string(event.Kv.Key)] = string(event.Kv.Value)
					remoteHost.mu.Unlock()

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

func notify(ctx context.Context) {

}
