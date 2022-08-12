package v1

import "github.com/p1nant0m/xdp-tracing/service/rest/store"

type Service interface {
	Policy() PolicySrv
}

type service struct {
	store store.Factory
}

func NewService(store store.Factory) Service {
	return &service{
		store: store,
	}
}

func (s *service) Policy() PolicySrv {
	return newPolicy(s)
}
