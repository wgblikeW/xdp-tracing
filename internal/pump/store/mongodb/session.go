package mongodb

import (
	"context"

	v1 "github.com/p1nant0m/xdp-tracing/pkg/api/v1"
	metav1 "github.com/p1nant0m/xdp-tracing/pkg/meta/v1"
	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

const (
	DeleteNothing int64 = 0
)

var filterFunc = func(id string) bson.M {
	return bson.M{"identifier": id}
}

type session struct {
	client *mongo.Client
}

func newSession(ds *datastore) *session {
	return &session{ds.client}
}

func (s *session) GetSpecificSession(ctx context.Context, id string, opts metav1.GetSpecificSessionOptions) ([]*v1.Session, error) {
	coll := s.client.Database(opts.Database, opts.DBoptions...).Collection(opts.Collection, opts.CollecOptions...)
	filter := filterFunc(id)

	cur, err := coll.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var results []bson.M
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	packets := []*v1.Session{}

	for _, result := range results {
		session := v1.Session{}
		bsonBytes, _ := bson.Marshal(result)
		bson.Unmarshal(bsonBytes, &session)
		packets = append(packets, &session)
	}

	return packets, nil
}

func (s *session) SavingSession(ctx context.Context, packet *v1.Session, opts metav1.SavingSessionOptions) error {
	coll := s.client.Database(opts.Database, opts.DBoptions...).Collection(opts.Collection, opts.CollecOptions...)

	_, err := coll.InsertOne(ctx, packet)
	if err != nil {
		return err
	}

	return nil
}

func (s *session) DeleteSession(ctx context.Context, id string, opts metav1.DeleteSessionOptions) (int64, error) {
	coll := s.client.Database(opts.Database, opts.DBoptions...).Collection(opts.Collection, opts.CollecOptions...)
	filter := filterFunc(id)

	result, err := coll.DeleteMany(ctx, filter, opts.DeleteOptions...)
	if err != nil {
		return DeleteNothing, err
	}

	return result.DeletedCount, nil
}
