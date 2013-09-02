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
	Graphs      []Graph
	Description string
	ACL         []string
}

type Graph struct {
	Name              string
	Targets           []string
	Resolution        int64
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
