package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dgbttn/go-scim-server/db"
	"github.com/dgbttn/go-scim-server/errors"
	"github.com/dgbttn/go-scim-server/optional"
	"github.com/dgbttn/go-scim-server/schema"
	"github.com/dgbttn/go-scim-server/scim"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	userResourceHandler = UserResourceHandler{}
	// UserResourceType ...
	UserResourceType = scim.ResourceType{
		ID:          optional.NewString("User"),
		Name:        "User",
		Endpoint:    "/Users",
		Description: optional.NewString("User Account"),
		Schema:      schema.CoreUserSchema(),
		Handler:     userResourceHandler,
	}
)

// UserResourceHandler ...
type UserResourceHandler struct{}

// Create ...
func (h UserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	id := uuid.New().String()
	now := time.Now()
	resource := scim.Resource{
		ID:         id,
		ExternalID: h.externalID(attributes),
		Attributes: attributes,
		Meta: scim.Meta{
			ResourceType: UserResourceType.Name,
			Created:      &now,
			LastModified: &now,
			Location:     fmt.Sprintf("%s/%s", UserResourceType.Endpoint[1:], url.PathEscape(id)),
		},
	}
	// store resource
	if err := db.MongoDB.Insert(resource.Map(UserResourceType)); err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	// return stored resource
	return resource, nil
}

// Get ...
func (h UserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	// check if resource exists
	user, err := db.MongoDB.Find(id)
	if err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	if len(user) == 0 {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	return h.attributesToResource(user), nil
}

// GetAll ...
func (h UserResourceHandler) GetAll(r *http.Request, params *scim.ListRequestParams) (scim.Page, error) {
	resources := make([]scim.Resource, 0)

	data, err := db.MongoDB.GetAll()
	if err != nil {
		return scim.Page{}, errors.ScimErrorInternal
	}

	if len(data) == 0 {
		return scim.Page{
			TotalResults: 0,
			Resources:    []scim.Resource{},
		}, nil
	}

	var from, to int
	if params.StartIndex > 0 {
		from = params.StartIndex - 1
	}
	if from+params.Count >= len(data) {
		params.Count = len(data) - from
	}
	to = from + params.Count
	for _, user := range data[from:to] {
		resources = append(resources, h.attributesToResource(user))
	}

	return scim.Page{
		TotalResults: len(data),
		Resources:    resources,
	}, nil
}

// Replace ...
func (h UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// // check if resource exists
	// _, ok := h.data[id]
	// if !ok {
	// 	return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	// }

	// // replace (all) attributes
	// h.data[id] = UserData{
	// 	resourceAttributes: attributes,
	// }

	// // return resource with replaced attributes
	// return scim.Resource{
	// 	ID:         id,
	// 	ExternalID: h.externalID(attributes),
	// 	Attributes: attributes,
	// }, nil
	return scim.Resource{}, nil
}

// Delete ...
func (h UserResourceHandler) Delete(r *http.Request, id string) error {
	// check if resource exists
	user, err := db.MongoDB.Find(id)
	if err != nil {
		return errors.ScimErrorInternal
	}
	if len(user) == 0 {
		return errors.ScimErrorResourceNotFound(id)
	}

	// delete resource
	db.MongoDB.Delete(id)
	return nil
}

// Patch ...
func (h UserResourceHandler) Patch(r *http.Request, id string, req scim.PatchRequest) (scim.Resource, error) {
	// for _, op := range req.Operations {
	// 	switch op.Op {
	// 	case scim.PatchOperationAdd:
	// 		if op.Path != "" {
	// 			h.data[id].resourceAttributes[op.Path] = op.Value
	// 		} else {
	// 			valueMap := op.Value.(map[string]interface{})
	// 			for k, v := range valueMap {
	// 				if arr, ok := h.data[id].resourceAttributes[k].([]interface{}); ok {
	// 					arr = append(arr, v)
	// 					h.data[id].resourceAttributes[k] = arr
	// 				} else {
	// 					h.data[id].resourceAttributes[k] = v
	// 				}
	// 			}
	// 		}
	// 	case scim.PatchOperationReplace:
	// 		if op.Path != "" {
	// 			h.data[id].resourceAttributes[op.Path] = op.Value
	// 		} else {
	// 			valueMap := op.Value.(map[string]interface{})
	// 			for k, v := range valueMap {
	// 				h.data[id].resourceAttributes[k] = v
	// 			}
	// 		}
	// 	case scim.PatchOperationRemove:
	// 		h.data[id].resourceAttributes[op.Path] = nil
	// 	}
	// }

	// created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", h.data[id].meta["created"]), time.UTC)
	// now := time.Now()

	// // return resource with replaced attributes
	// return scim.Resource{
	// 	ID:         id,
	// 	ExternalID: h.externalID(h.data[id].resourceAttributes),
	// 	Attributes: h.data[id].resourceAttributes,
	// 	Meta: scim.Meta{
	// 		Created:      &created,
	// 		LastModified: &now,
	// 		Version:      fmt.Sprintf("%s.patch", h.data[id].meta["version"]),
	// 	},
	// }, nil
	return scim.Resource{}, nil
}

func (h UserResourceHandler) externalID(attributes map[string]interface{}) optional.String {
	if eID, ok := attributes["externalId"]; ok {
		externalID, ok := eID.(string)
		if !ok {
			return optional.String{}
		}
		return optional.NewString(externalID)
	}

	return optional.String{}
}

func (h UserResourceHandler) id(attributes map[string]interface{}) string {
	if id, ok := attributes["id"]; ok {
		idStr, ok := id.(string)
		if !ok {
			return ""
		}
		return idStr
	}
	return ""
}

func (h UserResourceHandler) meta(attributes map[string]interface{}) (meta scim.Meta) {
	if m, ok := attributes["meta"]; ok {
		b, err := bson.MarshalExtJSON(m, true, true)
		if err != nil {
			return
		}
		var metaAttr map[string]string
		err = json.Unmarshal(b, &metaAttr)
		if err != nil {
			return
		}
		resourceType, ok := metaAttr["resourcetype"]
		if !ok {
			return
		}
		location, ok := metaAttr["location"]
		if !ok {
			return
		}
		meta.ResourceType = resourceType
		meta.Location = location
		if created, ok := metaAttr["created"]; ok {
			createdTime, _ := time.Parse(time.RFC3339, created)
			meta.Created = &createdTime
		}
		if lastModified, ok := metaAttr["lastmodified"]; ok {
			lastModifiedTime, _ := time.Parse(time.RFC3339, lastModified)
			meta.LastModified = &lastModifiedTime
		}
		if version, ok := metaAttr["version"]; ok {
			meta.Version = version
		}
	}
	return
}

func (h UserResourceHandler) attributesToResource(attributes map[string]interface{}) scim.Resource {
	return scim.Resource{
		ID:         h.id(attributes),
		ExternalID: h.externalID(attributes),
		Attributes: attributes,
		Meta:       h.meta(attributes),
	}
}
