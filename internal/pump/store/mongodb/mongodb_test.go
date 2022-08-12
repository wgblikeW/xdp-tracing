package mongodb_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
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

	id := uuid.New().String()
	err = storeIns.Session().SavingSession(context.Background(), &v1.Session{InstanceID: id, TCPIPIdentifier: metav1.TCPIPIdentifier{SrcIP: "192.168.176.128"}},
		metav1.SavingSessionOptions{MongoDBGenericOptions: &metav1.MongoDBGenericOptions{Database: "xdp-tracing", Collection: "session"}})
	if err != nil {
		t.Fatal("fail to get id=123", "err=", err)
	}

	packets, err := storeIns.Session().GetSpecificSession(context.Background(), id,
		metav1.GetSpecificSessionOptions{MongoDBGenericOptions: &metav1.MongoDBGenericOptions{Database: "xdp-tracing", Collection: "session"}})
	if err != nil {
		t.Fatal("fail to get id=123", "err=", err, "packets", packets)
	}

	if packets[0].SrcIP != "192.168.176.128" {
		t.Errorf("Expected srcIP 192.168.176.128, got %v", packets[0].SrcIP)
	}

	count, err := storeIns.Session().DeleteSession(context.Background(), id, metav1.DeleteSessionOptions{MongoDBGenericOptions: &metav1.MongoDBGenericOptions{Database: "xdp-tracing", Collection: "session"}})
	if count != 1 || err != nil {
		t.Errorf("Expected count 1 no error,got %v %v", count, err)
	}

	t.Log(packets[0])
}
