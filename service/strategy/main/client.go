package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/p1nant0m/xdp-tracing/service/strategy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	addr = flag.String("addr", "localhost:50002", "the address to connect to")
)

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := strategy.NewStrategyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.InstallStrategy(ctx, &strategy.UpdateStrategy{Blockoutrules: []byte("192.168.176.133")})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	fmt.Println(r.Status)
}
