package bridge

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func (b *Bridge) fetchFromAS(method, gatewayID, gatewayKey string, response interface{}) error {
	keyParts := strings.SplitN(gatewayKey, ".", 2)
	if len(keyParts) != 2 {
		return errors.New("Invalid access key")
	}
	server, ok := b.accountServers[keyParts[0]]
	if !ok {
		return fmt.Errorf("Account server %s not found", keyParts[0])
	}
	req, err := http.NewRequest("GET", server+fmt.Sprintf(method, gatewayID), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Key "+gatewayKey)
	res, err := b.http.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()
	if res.StatusCode != 200 {
		return fmt.Errorf("Account Server returned %d", res.StatusCode)
	}
	return json.NewDecoder(res.Body).Decode(response)
}

type Gateway struct {
	FrequencyPlan   string `json:"frequency_plan"`
	LocationPublic  bool   `json:"location_public"`
	StatusPublic    bool   `json:"status_public"`
	OwnerPublic     bool   `json:"owner_public"`
	AntennaLocation *struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
		Altitude  int     `json:"altitude"`
	} `json:"antenna_location,omitempty"`
	Collaborators []struct {
		Username string `json:"username"`
	} `json:"collaborators"`
	Attributes map[string]string `json:"attributes"`
	Owner      struct {
		Username string `json:"username"`
	} `json:"owner"`
	Token struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   uint64 `json:"expires_in"`
	} `json:"token"`
}

func (b *Bridge) fetchGateway(gatewayID, gatewayKey string) (*Gateway, error) {
	var gateway Gateway
	err := b.fetchFromAS("/api/v2/gateways/%s", gatewayID, gatewayKey, &gateway)
	if err != nil {
		return nil, err
	}
	return &gateway, nil
}
