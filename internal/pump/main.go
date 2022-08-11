package main

import (
	"context"

	"github.com/p1nant0m/xdp-tracing/internal/pump/store/mongodb"
	v1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
	"github.com/p1nant0m/xdp-tracing/pkg/options"
	"github.com/sirupsen/logrus"
)

func main() {
	storeIns, err := mongodb.GetMongoDBFactoryOr(options.NewMongoDBOptions())
	if err != nil {
		logrus.Fatal("fail to initialize the mongodb instance.", "err=", err)
	}

	packets, err := storeIns.Session().GetSpecificSession(context.Background(), "123",
		v1.GetSpecificSessionOptions{Database: "xdp-tracing", Collection: "session", MongoDBGenericOptions: &v1.MongoDBGenericOptions{}})
	if err != nil {
		logrus.Fatal("fail to get id=123", "err=", err, "packets", packets)
	}
	logrus.Infof("%v", packets[0])
}
