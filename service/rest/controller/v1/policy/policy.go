package policy

import (
	v1 "github.com/p1nant0m/xdp-tracing/service/rest/service/v1"
	"github.com/p1nant0m/xdp-tracing/service/rest/store"
)

type PolicyController struct {
	srv v1.Service
}

func NewPolicyController(store store.Factory) *PolicyController {
	return &PolicyController{srv: v1.NewService(store)}
}
