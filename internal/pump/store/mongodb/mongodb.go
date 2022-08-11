package mongodb

import (
	"fmt"
	"sync"

	"github.com/p1nant0m/xdp-tracing/internal/pump/store"
	"github.com/p1nant0m/xdp-tracing/pkg/db"
	"github.com/p1nant0m/xdp-tracing/pkg/options"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	mogodbFactory store.Factory
	once          sync.Once
)

type datastore struct {
	client *mongo.Client
}

func (ds *datastore) Session() store.SessionStore {
	return newSession(ds)
}

func (ds *datastore) Sessions() store.SessionsStore {
	return newSessions(ds)
}

func GetMongoDBFactoryOr(opts *options.MongoDBOptions) (store.Factory, error) {
	if opts == nil && mogodbFactory == nil {
		return nil, fmt.Errorf("failed to get mongodb store facotry")
	}

	var err error
	var dbClient *mongo.Client
	once.Do(func() {
		options := &db.Options{
			Hosts:                  opts.Hosts,
			Username:               opts.Username,
			Password:               opts.Password,
			WriteConcern:           opts.WriteConcern,
			ConnectTimeout:         opts.ConnectTimeout,
			MaxPoolSize:            opts.MaxPoolSize,
			MinPoolSize:            opts.MinPoolSize,
			ServerSelectionTimeout: opts.ServerSelectionTimeout,
		}
		dbClient, err = db.NewMongoDBClient(options)

		mogodbFactory = &datastore{dbClient}
	})

	if mogodbFactory == nil || err != nil {
		return nil, fmt.Errorf("failed to get mysql store fatory, mogodbFactory: %+v, error: %w", mogodbFactory, err)
	}

	return mogodbFactory, nil
}
