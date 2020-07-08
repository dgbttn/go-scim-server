package main

import (
	"log"
	"net/http"

	"github.com/dgbttn/go-scim-server/db"
	"github.com/dgbttn/go-scim-server/scim"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

func initServer() {
	server := scim.Server{
		Config: scim.ServiceProviderConfig{},
		ResourceTypes: []scim.ResourceType{
			UserResourceType,
		},
	}

	log.Fatal(http.ListenAndServe(":8082", server))
}

func connectMongoDB() {
	connectionStr := viper.GetString("MONGODB_CONNECTION")
	databaseStr := viper.GetString("DATABASE")
	collectionStr := viper.GetString("COLLECTION")
	if err := db.MongoDB.ConnectDB(connectionStr, databaseStr, collectionStr); err != nil {
		panic(err)
	}
}

func main() {
	viper.AutomaticEnv()
	connectMongoDB()
	initServer()
}
