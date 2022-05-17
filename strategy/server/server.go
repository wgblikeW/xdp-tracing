package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/p1nant0m/xdp-tracing/strategy"
	"google.golang.org/grpc"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

type server struct {
	strategy.UnimplementedStrategyServer
}

func (s *server) InstallStrategy(ctx context.Context,
	in *strategy.UpdateStrategy) (*strategy.UpdateStrategyReply, error) {
	fmt.Printf("Blockout Rules:%v", string(in.Blockoutrules))
	return &strategy.UpdateStrategyReply{Status: "OK"}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	strategy.RegisterStrategyServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
