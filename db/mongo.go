package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IMongoDB ...
type IMongoDB interface {
	GetClient() *mongo.Client
	ConnectDB(connectionString string, database string, collection string) error
}

type mongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (db *mongoDB) GetClient() *mongo.Client {
	return db.client
}

func (db *mongoDB) ConnectDB(connectionStr string, databaseStr string, collectionStr string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelFunc()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionStr))
	if err != nil {
		return err
	}
	collection := client.Database(databaseStr).Collection(collectionStr)
	db.client = client
	db.collection = collection
	return nil
}

// MongoDB ...
var MongoDB IMongoDB = &mongoDB{}
