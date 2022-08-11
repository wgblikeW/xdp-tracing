package store

import (
	"context"

	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
)

// SessionStore is an interface that defines the methods that will be
// used in Service layer, and it operates in storage layer.
type SessionsStore interface {
	GetAllSessions(context.Context, metav1.GetAllSessionsOptions) error
}
