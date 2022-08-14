package v1

import (
	"context"

	v1 "github.com/p1nant0m/xdp-tracing/pkg/api/v1"
	"github.com/p1nant0m/xdp-tracing/service/rest/store"
)

type PolicySrv interface {
	Create(context.Context, v1.Policy) error
	Delete(context.Context, string) error
	List(context.Context) ([]string, error)
}

type policyService struct {
	store store.Factory
}

func newPolicy(srv *service) *policyService {
	return &policyService{store: srv.store}
}

func (p *policyService) Create(ctx context.Context, policy v1.Policy) error {
	if err := p.store.Policy().Create(policy.Policy); err != nil {
		return err
	}

	return nil
}

func (p *policyService) Delete(ctx context.Context, policy string) error {
	if err := p.store.Policy().Delete(string(policy)); err != nil {
		return err
	}

	return nil
}

func (p *policyService) List(ctx context.Context) ([]string, error) {
	policies, err := p.store.Policy().List()
	if err != nil {
		return nil, err
	}
	return policies, nil
}
