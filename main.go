package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/dgbttn/go-scim-server/optional"
	"github.com/dgbttn/go-scim-server/schema"
	"github.com/dgbttn/go-scim-server/scim"
)

func makeData() map[string]Data {
	data := make(map[string]Data)
	// Generate enough test data to test pagination
	for i := 1; i < 21; i++ {
		data[fmt.Sprintf("000%d", i)] = Data{
			resourceAttributes: scim.ResourceAttributes{
				"userName": fmt.Sprintf("test%d", i),
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

func newServer() scim.Server {
	data := makeData()
	resourceHandler := ResourceHandler{data: data}

	return scim.Server{
		Config: scim.ServiceProviderConfig{},
		ResourceTypes: []scim.ResourceType{
			{
				ID:          optional.NewString("User"),
				Name:        "User",
				Endpoint:    "/Users",
				Description: optional.NewString("User Account"),
				Schema:      schema.CoreUserSchema(),
				Handler:     resourceHandler,
			},
		},
	}
}

func main() {
	server := newServer()
	log.Fatal(http.ListenAndServe(":8080", server))
}
