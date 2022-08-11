package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Options struct {
	Hosts                  []string
	Username               string
	Password               string
	WriteConcern           string
	ConnectTimeout         time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	ServerSelectionTimeout uint64
}

func NewMongoDBClient(opts *Options) (*mongo.Client, error) {
	// connection URI format to connect mongodb Server
	dsn := fmt.Sprintf("mongodb://%v:%v@%v/?minPoolSize=%v&maxPoolSize=%v&w=%v&connectTimeoutMS=%d&serverSelectionTimeoutMS=%v",
		opts.Username, opts.Password, strings.Join(opts.Hosts, ","),
		opts.MinPoolSize, opts.MaxPoolSize, opts.WriteConcern, opts.ConnectTimeout, opts.ServerSelectionTimeout)

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return nil, err
	}

	return client, nil
}
