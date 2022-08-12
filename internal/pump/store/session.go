package store

import (
	"context"

	v1 "github.com/p1nant0m/xdp-tracing/pkg/api/v1"
	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
)

// SessionStore is an interface that defines the methods that will be
// used in Service layer, and it operates in storage layer.
type SessionStore interface {
	SavingSession(context.Context, *v1.Session, metav1.SavingSessionOptions) error
	GetSpecificSession(context.Context, string, metav1.GetSpecificSessionOptions) ([]*v1.Session, error)
	DeleteSession(context.Context, string, metav1.DeleteSessionOptions) (int64, error)
}
