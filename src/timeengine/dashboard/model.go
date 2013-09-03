package dashboard

import (
	"encoding/json"
)

type Dashboard struct {
	// Definition. Currently just a string.
	// It actually is a DashConfig object serialized with JSON.
	G []byte

	// Name.
	// Immutable, but not stored anywhere anyways.
	name string
}

type DashConfig struct {
	Description string   `json:"description"`
	Graphs      []Graph  `json:"graphs"`
	ACL         []string `json:"acl,omitempty"`
}

type Expression struct {
	Label string `json:"label"`
	Expr  string `json:"expr"`
}

type Graph struct {
	Name        string                 `json:"name"`
	Targets     []string               `json:"targets"`
	Expressions []Expression           `json:"expressions,omitempty"`
	Resolution  int64                  `json:"resolution,omitempty"`
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
