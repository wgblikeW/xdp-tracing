package strategy

import (
	"context"
	"fmt"
)

type Server struct {
	UnimplementedStrategyServer
}

func (s *Server) InstallStrategy(ctx context.Context,
	in *UpdateStrategy) (*UpdateStrategyReply, error) {
	fmt.Printf("Blockout Rules:%v", string(in.Blockoutrules))
	return &UpdateStrategyReply{Status: "OK"}, nil
}
