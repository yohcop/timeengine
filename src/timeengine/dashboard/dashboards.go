package dashboard

import (
	"encoding/json"
	"errors"
	"log"
	"regexp"
	"strings"

	"appengine"
	"appengine/datastore"
)

var _ = log.Println

func ValidDashboard(name string) (string, error) {
	name = strings.ToLower(name)
	if regexp.MustCompile("[a-z0-9-_.]+").Match([]byte(name)) {
		return name, nil
	}
	return "", errors.New(
		"Invalid dashboard. Should match /[a-z0-9-_.]+/")
}

func DashboardKey(c appengine.Context, name string) *datastore.Key {
	return datastore.NewKey(c, "Dash", name, 0, nil)
}

func GetDashFromDatastore(c appengine.Context, name string) *Dashboard {
	ts := &Dashboard{}
	if err := datastore.Get(c, DashboardKey(c, name), ts); err != nil {
		return nil
	}
	return ts
}

func ValidateRawConfig(rawData []byte) (*DashConfig, error) {
	asInterface := make(map[string]interface{})
	if err := json.Unmarshal(rawData, &asInterface); err != nil {
		return nil, err
	}
	asInterfaceTxt, _ := json.Marshal(asInterface)

	// Prepare the new config: parse it from JSON.
	data := DashConfig{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		return nil, err
	}
	fromCfg, _ := json.Marshal(data)

	if len(fromCfg) != len(asInterfaceTxt) {
		return nil, errors.New("You have unknown fields in your json, " +
			"or missing required fields.")
	}
	return &data, nil
}
