package scim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
)

// ProvisioningClient ...
type ProvisioningClient struct {
	BaseURI string
	Params  map[string]string
}

// URL ...
func (p *ProvisioningClient) URL(identifier string) string {
	URL := p.BaseURI
	if len(identifier) > 0 {
		URL += "/" + identifier
	}
	params := url.Values{}
	for k, v := range p.Params {
		params.Add(k, v)
	}
	URL += "?" + params.Encode()
	return URL
}

// Post ...
func (p *ProvisioningClient) Post(body io.Reader) (id string, extraAttributes map[string]string, erro error) {
	req, erro := http.NewRequest(http.MethodPost, p.URL(""), body)
	if erro != nil {
		return
	}

	req.Header.Add("Content-Type", "application/json")
	resp, erro := http.DefaultClient.Do(req)
	if erro != nil {
		return
	}
	defer resp.Body.Close()

	returnedData, erro := ioutil.ReadAll(resp.Body)
	if erro != nil {
		return
	}
	if resp.StatusCode != 201 {
		erro = fmt.Errorf(string(returnedData))
		return
	}

	fmt.Println("Response from POST provisioning: ", string(returnedData))

	returnedAttrs := make(map[string]interface{})
	if erro = json.Unmarshal(returnedData, &returnedAttrs); erro != nil {
		return
	}

	externalID, ok := returnedAttrs["externalId"].(string)
	if !ok {
		erro = fmt.Errorf("Not found externalId as string")
		return
	}
	id, ok = returnedAttrs["id"].(string)
	if !ok {
		erro = fmt.Errorf("Not found id as string")
		return
	}
	extraAttributes = make(map[string]string)
	extraAttributes["externalId"] = externalID
	return
}

// Patch ...
func (p *ProvisioningClient) Patch(id string, body io.Reader) (erro error) {
	req, erro := http.NewRequest(http.MethodPatch, p.URL(id), body)
	if erro != nil {
		return
	}
	req.Header.Add("Content-Type", "application/json")
	resp, erro := http.DefaultClient.Do(req)
	if erro != nil {
		return
	}
	defer resp.Body.Close()
	returnedData, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response from PATCH provisioning: ", string(returnedData))
	return
}

// Delete ...
func (p *ProvisioningClient) Delete(id string) (erro error) {
	req, erro := http.NewRequest(http.MethodDelete, p.URL(id), bytes.NewBuffer([]byte{}))
	if erro != nil {
		return
	}
	resp, erro := http.DefaultClient.Do(req)
	if erro != nil {
		return
	}
	defer resp.Body.Close()
	returnedData, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("Response from DELETE provisioning: ", string(returnedData))
	return
}
