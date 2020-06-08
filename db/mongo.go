package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// IMongoDB ...
type IMongoDB interface {
	GetClient() *mongo.Client
	ConnectDB(connectionString string, database string, collection string) error
	Insert(document map[string]interface{}) error
}

type mongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (db *mongoDB) GetClient() *mongo.Client {
	return db.client
}

func (db *mongoDB) ConnectDB(connectionStr string, databaseStr string, collectionStr string) error {
	ctx := contextWithTimeout(10)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionStr))
	if err != nil {
		return err
	}
	if err = client.Ping(contextWithTimeout(2), readpref.Primary()); err != nil {
		return err
	}
	collection := client.Database(databaseStr).Collection(collectionStr)
	db.client = client
	db.collection = collection
	return nil
}

func (db *mongoDB) Insert(document map[string]interface{}) error {
	_, err := db.collection.InsertOne(contextWithTimeout(5), document)
	return err
}

// MongoDB ...
var MongoDB IMongoDB = &mongoDB{}

func contextWithTimeout(sec time.Duration) context.Context {
	ctx, _ := context.WithTimeout(context.Background(), sec*time.Second)
	return ctx
}
