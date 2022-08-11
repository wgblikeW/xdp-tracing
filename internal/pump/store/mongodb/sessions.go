package mongodb

import (
	"context"

	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
	"go.mongodb.org/mongo-driver/mongo"
)

type sessions struct {
	client *mongo.Client
}

func newSessions(ds *datastore) *sessions {
	return &sessions{ds.client}
}

func (s *sessions) GetAllSessions(ctx context.Context, opts metav1.GetAllSessionsOptions) error {
	return nil
}
