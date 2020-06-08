package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/dgbttn/go-scim-server/errors"
	"github.com/dgbttn/go-scim-server/scim"
)

// UserData ...
type UserData struct {
	resourceAttributes scim.ResourceAttributes
	meta               map[string]string
}

// UserResourceHandler ...
type UserResourceHandler struct {
	data map[string]UserData
}

// Create ...
func (h UserResourceHandler) Create(r *http.Request, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// create unique identifier
	rand.Seed(time.Now().UnixNano())
	id := fmt.Sprintf("%04d", rand.Intn(9999))

	// store resource
	h.data[id] = UserData{
		resourceAttributes: attributes,
	}

	now := time.Now()

	// return stored resource
	return scim.Resource{
		ID:         id,
		Attributes: attributes,
		Meta: scim.Meta{
			Created:      &now,
			LastModified: &now,
			Version:      fmt.Sprintf("v%s", id),
		},
	}, nil
}

// Get ...
func (h UserResourceHandler) Get(r *http.Request, id string) (scim.Resource, error) {
	// check if resource exists
	data, ok := h.data[id]
	if !ok {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", data.meta["created"]), time.UTC)
	lastModified, _ := time.Parse(time.RFC3339, fmt.Sprintf("%v", data.meta["lastModified"]))

	// return resource with given identifier
	return scim.Resource{
		ID:         id,
		Attributes: data.resourceAttributes,
		Meta: scim.Meta{
			Created:      &created,
			LastModified: &lastModified,
			Version:      fmt.Sprintf("%v", data.meta["version"]),
		},
	}, nil
}

// GetAll ...
func (h UserResourceHandler) GetAll(r *http.Request, params scim.ListRequestParams) (scim.Page, error) {
	resources := make([]scim.Resource, 0)
	i := 1

	for k, v := range h.data {
		if i > (params.StartIndex + params.Count - 1) {
			break
		}

		if i >= params.StartIndex {
			resources = append(resources, scim.Resource{
				ID:         k,
				Attributes: v.resourceAttributes,
			})
		}
		i++
	}

	return scim.Page{
		TotalResults: len(h.data),
		Resources:    resources,
	}, nil
}

// Replace ...
func (h UserResourceHandler) Replace(r *http.Request, id string, attributes scim.ResourceAttributes) (scim.Resource, error) {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return scim.Resource{}, errors.ScimErrorResourceNotFound(id)
	}

	// replace (all) attributes
	h.data[id] = UserData{
		resourceAttributes: attributes,
	}

	// return resource with replaced attributes
	return scim.Resource{
		ID:         id,
		Attributes: attributes,
	}, nil
}

// Delete ...
func (h UserResourceHandler) Delete(r *http.Request, id string) error {
	// check if resource exists
	_, ok := h.data[id]
	if !ok {
		return errors.ScimErrorResourceNotFound(id)
	}

	// delete resource
	delete(h.data, id)

	return nil
}

// Patch ...
func (h UserResourceHandler) Patch(r *http.Request, id string, req scim.PatchRequest) (scim.Resource, error) {
	for _, op := range req.Operations {
		switch op.Op {
		case scim.PatchOperationAdd:
			if op.Path != "" {
				h.data[id].resourceAttributes[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					if arr, ok := h.data[id].resourceAttributes[k].([]interface{}); ok {
						arr = append(arr, v)
						h.data[id].resourceAttributes[k] = arr
					} else {
						h.data[id].resourceAttributes[k] = v
					}
				}
			}
		case scim.PatchOperationReplace:
			if op.Path != "" {
				h.data[id].resourceAttributes[op.Path] = op.Value
			} else {
				valueMap := op.Value.(map[string]interface{})
				for k, v := range valueMap {
					h.data[id].resourceAttributes[k] = v
				}
			}
		case scim.PatchOperationRemove:
			h.data[id].resourceAttributes[op.Path] = nil
		}
	}

	created, _ := time.ParseInLocation(time.RFC3339, fmt.Sprintf("%v", h.data[id].meta["created"]), time.UTC)
	now := time.Now()

	// return resource with replaced attributes
	return scim.Resource{
		ID:         id,
		Attributes: h.data[id].resourceAttributes,
		Meta: scim.Meta{
			Created:      &created,
			LastModified: &now,
			Version:      fmt.Sprintf("%s.patch", h.data[id].meta["version"]),
		},
	}, nil
}