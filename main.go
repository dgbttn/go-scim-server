package main

import (
	"log"
	"net/http"

	"github.com/dgbttn/go-scim-server/db"
	"github.com/dgbttn/go-scim-server/optional"
	"github.com/dgbttn/go-scim-server/schema"
	"github.com/dgbttn/go-scim-server/scim"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

func initServer() {
	UserProvisioner := scim.ProvisioningClient{
		BaseURI: "http://localhost:8081/scim/v2/Users",
		Params: map[string]string{
			"client_id": "02a8434b7c1e4758bf91638978cdb9c6",
		},
	}
	UserResourceType = scim.ResourceType{
		ID:          optional.NewString("User"),
		Name:        "User",
		Endpoint:    "/Users",
		Description: optional.NewString("User Account"),
		Schema:      schema.CoreUserSchema(),
		Handler:     userResourceHandler,
		Provisioner: &UserProvisioner,
	}

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
