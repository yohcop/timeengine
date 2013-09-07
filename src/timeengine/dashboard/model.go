package dashboard

import (
	"encoding/json"
)

type Dashboard struct {
	// Definition. Currently just a string.
	// It actually is a DashConfig object serialized with JSON.
	// The reason for it to stay as is, is we can actually change
	// the definition of DashConfig, without breaking the current
	// schema. Users can still retrieve the data, but can't save it.
	// They get a chance to edit the json object, and save it in the
	// new format.
	G []byte

	// Name.
	// Immutable, but not stored anywhere anyways.
	name string
}

// Variable name -> default.
type Preselection map[string]string

type DashConfig struct {
	Description string `json:"description"`
	// metric name -> variable name.
	Targets map[string]string `json:"targets"`
	Graphs  []Graph           `json:"graphs"`
	ACL     []string          `json:"acl,omitempty"`
	// preselection name -> config.
	// The "default" preselection can be used.
	Preselection map[string]Preselection `json:"preselection,omitempty"`
}

type Graph struct {
	Name        string                 `json:"name"`
	Expressions map[string]string      `json:"expressions"`
	DygraphOpts map[string]interface{} `json:"dygraphOpts,omitempty"`
}

func (d *Dashboard) Cfg() (*DashConfig, error) {
	cfg := &DashConfig{}
	err := json.Unmarshal(d.G, cfg)
	return cfg, err
}

func (d *Dashboard) IsAcled(email string) (bool, error) {
	cfg, err := d.Cfg()
	if err != nil {
		// If there is an error reading from the DB, we ignore.
		// This is useful in when updates break the schema.
		// If this never happen, we can do:
		// return false, err
	}
	// If there is no one in the ACL list, this is considered public.
	if len(cfg.ACL) == 0 {
		return true, nil
	}
	for _, authorizedEmail := range cfg.ACL {
		if authorizedEmail == email {
			return true, nil
		}
	}
	return false, nil
}
