package strategy

import (
	"context"
	"strings"
)

var localPolicyCache map[string]int = make(map[string]int)

type Server struct {
	UnimplementedStrategyServer
	LocalStrategyCh chan string
}

func (s *Server) InstallStrategy(ctx context.Context,
	in *UpdateStrategy) (*UpdateStrategyReply, error) {
	rulesList := strings.Split(string(in.Blockoutrules), " ")
	for _, rule := range rulesList {
		if _, exists := localPolicyCache[rule]; !exists {
			localPolicyCache[rule] = 0
			s.LocalStrategyCh <- string(rule)
		}
	}
	return &UpdateStrategyReply{Status: "OK"}, nil
}

func (s *Server) GetLocalStrategyCh() chan string {
	return s.LocalStrategyCh
}
