package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/p1nant0m/xdp-tracing/service/strategy"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	addr = flag.String("addr", "192.168.176.128:50003", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
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

	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := strategy.NewStrategyClient(conn)
	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	r, err := c.InstallStrategy(ctx, &strategy.UpdateStrategy{Blockoutrules: []byte("172.17.0.4")})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(r.Status)
}
