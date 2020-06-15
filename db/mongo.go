package db

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// IMongoDB ...
type IMongoDB interface {
	GetClient() *mongo.Client
	ConnectDB(connectionString string, database string, collection string) error
	Insert(document interface{}) error
	Find(id string) (bson.M, error)
	GetAll() ([]bson.M, error)
	Delete(id string) error
}

type mongoDB struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func (db *mongoDB) GetClient() *mongo.Client {
	return db.client
}

func (db *mongoDB) ConnectDB(connectionStr string, databaseStr string, collectionStr string) error {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(connectionStr))
	if err != nil {
		return err
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		return err
	}
	collection := client.Database(databaseStr).Collection(collectionStr)
	db.client = client
	db.collection = collection
	return nil
}

func (db *mongoDB) Insert(document interface{}) error {
	_, err := db.collection.InsertOne(context.TODO(), document)
	return err
}

func (db *mongoDB) Find(id string) (result bson.M, erro error) {
	erro = db.collection.FindOne(context.TODO(), bson.M{"id": id}).Decode(&result)
	if erro == mongo.ErrNoDocuments {
		erro = nil
	}
	return
}

func (db *mongoDB) GetAll() (results []bson.M, erro error) {
	cursor, erro := db.collection.Find(context.TODO(), bson.D{}, options.Find())
	if erro != nil {
	}
	erro = cursor.All(context.TODO(), &results)
	return
}

func (db *mongoDB) Delete(id string) error {
	_, err := db.collection.DeleteOne(context.TODO(), bson.M{"id": id})
	return err
}

var (
	// MongoDB ...
	MongoDB IMongoDB = &mongoDB{}
)
