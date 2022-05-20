package strategy

import (
	"context"
)

type Server struct {
	UnimplementedStrategyServer
	LocalStrategyCh chan string
}

func (s *Server) InstallStrategy(ctx context.Context,
	in *UpdateStrategy) (*UpdateStrategyReply, error) {
	s.LocalStrategyCh <- string(in.Blockoutrules)
	return &UpdateStrategyReply{Status: "OK"}, nil
}

func (s *Server) GetLocalStrategyCh() chan string {
	return s.LocalStrategyCh
}
