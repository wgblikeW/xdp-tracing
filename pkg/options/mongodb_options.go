package options

import "time"

// MongoDBOptions defines options for mysql database.
type MongoDBOptions struct {
	Hosts                  []string
	Username               string
	Password               string
	WriteConcern           string
	ConnectTimeout         time.Duration
	MaxPoolSize            uint64
	MinPoolSize            uint64
	ServerSelectionTimeout uint64
}

// NewMongoDBOptionscreate a `zero` value instance.
func NewMongoDBOptions() *MongoDBOptions {
	return &MongoDBOptions{
		Hosts:                  []string{"127.0.0.1:27017"},
		Username:               "root",
		Password:               "root",
		WriteConcern:           "majority",
		ConnectTimeout:         300000,
		MaxPoolSize:            100,
		MinPoolSize:            0,
		ServerSelectionTimeout: 30000,
	}
}
