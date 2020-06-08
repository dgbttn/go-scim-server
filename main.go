package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgbttn/go-scim-server/db"
	"github.com/dgbttn/go-scim-server/optional"
	"github.com/dgbttn/go-scim-server/schema"
	"github.com/dgbttn/go-scim-server/scim"

	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/viper"
)

func makeData() map[string]UserData {
	data := make(map[string]UserData)
	// Generate enough test data to test pagination
	for i := 1; i < 21; i++ {
		data[fmt.Sprintf("000%d", i)] = UserData{
			resourceAttributes: scim.ResourceAttributes{
				"userName":   fmt.Sprintf("test%d", i),
				"externalId": fmt.Sprintf("external%d", i),
			},
			meta: map[string]string{
				"created":      fmt.Sprintf("2020-01-%02dT15:04:05+07:00", i),
				"lastModified": fmt.Sprintf("2020-02-%02dT16:05:04+07:00", i),
				"version":      fmt.Sprintf("v%d", i),
			},
		}
	}
	return data
}

func initServer() {
	userData := makeData()
	userResourceHandler := UserResourceHandler{data: userData}

	server := scim.Server{
		Config: scim.ServiceProviderConfig{},
		ResourceTypes: []scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				Handler:     userResourceHandler,
			},
		},
	}

	log.Fatal(http.ListenAndServe(":8080", server))
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
