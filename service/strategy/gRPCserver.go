package strategy

import (
	"context"
	"strings"
)

var localPolicyCache map[string]struct{} = make(map[string]struct{})

const (
	REVOKE  = "revoke"
	INSTALL = "install"
)

type PolicyOp struct {
	Type string
	Rule string
}

type Server struct {
	UnimplementedStrategyServer
	LocalStrategyCh chan *PolicyOp
}

func (s *Server) InstallStrategy(ctx context.Context,
	in *UpdateStrategy) (*UpdateStrategyReply, error) {
	rulesList := strings.Split(string(in.Blockoutrules), " ")
	for _, rule := range rulesList {
		if _, exists := localPolicyCache[rule]; !exists {
			localPolicyCache[rule] = struct{}{}
			s.LocalStrategyCh <- &PolicyOp{Type: INSTALL, Rule: rule}
		}
	}
	return &UpdateStrategyReply{Status: "OK"}, nil
}

func (s *Server) RevokeStrategy(ctx context.Context,
	in *UpdateStrategy) (*UpdateStrategyReply, error) {
	rulesList := strings.Split(string(in.Blockoutrules), " ")
	for _, rule := range rulesList {
		if _, exists := localPolicyCache[rule]; !exists {
			delete(localPolicyCache, rule)
			s.LocalStrategyCh <- &PolicyOp{Type: REVOKE, Rule: rule}
		}
	}
	return &UpdateStrategyReply{Status: "OK"}, nil
}

func (s *Server) GetLocalStrategyCh() chan *PolicyOp {
	return s.LocalStrategyCh
}
