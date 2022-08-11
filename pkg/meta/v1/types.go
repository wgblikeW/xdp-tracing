package v1

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
)

// ObjectMeta is metadata that all persisted resources must have, which includes all objects
// ObjectMeta is also used by mongo.
type ObjectMeta struct {
	ID         string
	CreatedAt  time.Time
	InstanceID string
}

type GetAllSessionsOptions struct {
}

type TCPIPIdentifier struct {
	SrcIP   string `bson:"srcip"`
	DstIP   string `bson:"dstip"`
	SrcPort int32  `bson:"src_port"`
	DstPort int32  `bson:"dst_port"`
}

type SavingSessionOptions struct {
}

type GetSpecificSessionOptions struct {
	*MongoDBGenericOptions
	Database   string
	Collection string
}

type MongoDBGenericOptions struct {
	DBoptions     []*options.DatabaseOptions
	CollecOptions []*options.CollectionOptions
	FindOptions   []*options.FindOptions
}
