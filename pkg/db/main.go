package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
)

type Address struct {
	Street string
	City   string
	State  string
}
type Student struct {
	FirstName string  `bson:"first_name,omitempty"`
	LastName  string  `bson:"last_name,omitempty"`
	Address   Address `bson:"inline"`
	Age       int
}

// Connection URI
const uri = "mongodb://root:root@192.168.176.128:27017/?maxPoolSize=20&w=majority"

func main() {
	// Create a new client and connect to the server

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	// Ping the primary
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		panic(err)
	}
	fmt.Println("Successfully connected and pinged.")

	// coll := client.Database("school").Collection("students")
	// address1 := Address{"1 Lakewood Way", "Elwood City", "PA"}
	// student1 := Student{FirstName: "Arthur", Address: address1, Age: 8}
	// _, err = coll.InsertOne(context.TODO(), student1)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	coll := client.Database("school").Collection("students")
	filter := bson.M{"age": 8}
	var result bson.M
	err = coll.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(result)
}
