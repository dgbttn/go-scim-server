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
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	clientID        = viper.GetString("CLIENT_ID")
	provisionerURL  = viper.GetString("PROVISIONING_CLIENT_URL")
	userProvisioner = scim.ProvisioningClient{
		BaseURI: provisionerURL,
		Params:  map[string]string{"client_id": clientID},
	}
	userResourceHandler = UserResourceHandler{}
	// UserResourceType ...
	UserResourceType = scim.ResourceType{
		ID:          optional.NewString("User"),
		Name:        "User",
		Endpoint:    "/Users",
		Description: optional.NewString("User Account"),
		Schema:      schema.CoreUserSchema(),
		Handler:     userResourceHandler,
		Provisioner: &userProvisioner,
	}
)

// UserResourceHandler ...
type UserResourceHandler struct{}

// Create ...
func (h UserResourceHandler) Create(r *http.Request, userInfo scim.ResourceAttributes) (scim.Resource, error) {
	id := uuid.New().String()
	now := time.Now()
	_, externalID, attributes, _ := h.extractUserData(userInfo)
	resource := scim.Resource{
		ID:         id,
		ExternalID: externalID,
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
	delete(user, "_id")
	return h.userDataToResource(user), nil
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
		delete(user, "_id")
		resources = append(resources, h.userDataToResource(user))
	}

	return scim.Page{
		TotalResults: len(data),
		Resources:    resources,
	}, nil
}

// Replace ...
func (h UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	user, err := db.MongoDB.Find(id)
	if err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	if len(user) == 0 {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}
	db.MongoDB.Delete(id)

	created := h.userDataToResource(user).Meta.Created
	lastModified := time.Now()

	_, externalID, attrs, meta := h.extractUserData(attributes)
	newUser := scim.Resource{
		ID:         id,
		ExternalID: externalID,
		Attributes: attrs,
		Meta: scim.Meta{
			ResourceType: meta.ResourceType,
			Created:      created,
			LastModified: &lastModified,
			Location:     meta.Location,
		},
	}
	// store resource
	if err := db.MongoDB.Insert(newUser.Map(UserResourceType)); err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	return newUser, nil
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
	user, err := db.MongoDB.Find(id)
	if err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	if len(user) == 0 {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	for _, op := range req.Operations {
		switch op.Op {
		case scim.PatchOperationAdd:
			if op.Path != "" {
				user[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					if arr, ok := v.([]interface{}); ok {
						list, _ := json.Marshal(user[k])
						oldArr := []interface{}{}
						json.Unmarshal(list, &oldArr)
						user[k] = append(oldArr, arr...)
					} else {
						user[k] = v
					}
				}
			}
		case scim.PatchOperationReplace:
			if op.Path != "" {
				user[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					user[k] = v
				}
			}
		case scim.PatchOperationRemove:
			user[op.Path] = nil
		}
	}

	db.MongoDB.Delete(id)

	meta := h.userDataToResource(user).Meta
	lastModified := time.Now()

	_, externalID, attrs, _ := h.extractUserData(user)

	// return resource with replaced attributes
	newUser := scim.Resource{
		ID:         id,
		ExternalID: externalID,
		Attributes: attrs,
		Meta: scim.Meta{
			ResourceType: meta.ResourceType,
			Created:      meta.Created,
			LastModified: &lastModified,
			Location:     meta.Location,
		},
	}
	// store resource
	if err := db.MongoDB.Insert(newUser.Map(UserResourceType)); err != nil {
		return scim.Resource{}, errors.ScimErrorInternal
	}
	return newUser, nil
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

func (h UserResourceHandler) getSlice(v interface{}) (s []interface{}, ok bool) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &s)
	if err != nil {
		return
	}
	return s, true
}

func (h UserResourceHandler) getMap(v interface{}) (m map[string]interface{}, ok bool) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return
	}
	return m, true
}

func (h UserResourceHandler) cleanMap(userData map[string]interface{}) map[string]interface{} {
	for k, v := range userData {
		if v == nil {
			delete(userData, k)
			continue
		}
		if attr, ok := h.getMap(v); ok {
			userData[k] = h.cleanMap(attr)
			continue
		}
		if attr, ok := h.getSlice(v); ok {
			for i := 0; i < len(attr); i++ {
				if element, ok := h.getMap(attr[i]); ok {
					attr[i] = h.cleanMap(element)
				}
			}
			userData[k] = attr
		}
	}
	return userData
}

func (h UserResourceHandler) extractUserData(userData map[string]interface{}) (id string, externalID optional.String, attributes map[string]interface{}, meta scim.Meta) {
	// id
	if idAttr, ok := userData["id"]; ok {
		id, _ = idAttr.(string)
	}

	// externalID
	if eIDAttr, ok := userData["externalId"]; ok {
		if eIDStr, ok := eIDAttr.(string); ok {
			externalID = optional.NewString(eIDStr)
		}
	}

	// other attributes
	attributes = h.cleanMap(userData)

	// meta
	if m, ok := userData["meta"]; ok {
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

func (h UserResourceHandler) userDataToResource(userData map[string]interface{}) scim.Resource {
	id, externalID, attributes, meta := h.extractUserData(userData)
	return scim.Resource{
		ID:         id,
		ExternalID: externalID,
		Attributes: attributes,
		Meta:       meta,
	}
}
