package mongodb_test

import (
	"context"
	"testing"

	"github.com/p1nant0m/xdp-tracing/internal/pump/store/mongodb"
	v1 "github.com/p1nant0m/xdp-tracing/pkg/api/v1"
	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
	"github.com/p1nant0m/xdp-tracing/pkg/options"
)

func TestMongoDBStorage(t *testing.T) {
	storeIns, err := mongodb.GetMongoDBFactoryOr(options.NewMongoDBOptions())
	if err != nil {
		t.Fatal("fail to initialize the mongodb instance.", "err=", err)
	}
	err = storeIns.Session().SavingSession(context.Background(), &v1.Session{InstanceID: "233"},
		metav1.SavingSessionOptions{MongoDBGenericOptions: &metav1.MongoDBGenericOptions{Database: "xdp-tracing", Collection: "session"}})
	if err != nil {
		t.Fatal("fail to get id=123", "err=", err)
	}

	packets, err := storeIns.Session().GetSpecificSession(context.Background(), "233",
		metav1.GetSpecificSessionOptions{MongoDBGenericOptions: &metav1.MongoDBGenericOptions{Database: "xdp-tracing", Collection: "session"}})
	if err != nil {
		t.Fatal("fail to get id=123", "err=", err, "packets", packets)
	}

	t.Log(packets[0])
}
