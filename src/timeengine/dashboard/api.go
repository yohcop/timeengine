package dashboard

import (
	"encoding/json"
	"log"
	"net/http"

	"timeengine/users"

	"appengine"
	"appengine/datastore"
)

var _ = log.Println

func NewDashboard(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := users.IsAuthorized(r); !ok {
		return
	}

	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	// Check if the dashboard already exists.
	if GetDashFromDatastore(c, d) != nil {
		http.Error(w, "Dashboard exists", http.StatusBadRequest)
		return
	}

	rawData := []byte(r.FormValue("data"))
	data := make([]byte, 0)
	if len(rawData) > 0 {
		if cfg, err := ValidateRawConfig(rawData); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			data, _ = json.Marshal(cfg)
		}
	} else {
		data, _ = json.Marshal(&DashConfig{
			Targets: map[string]string{
				"var1": "my.namespace*my.target.0",
				"var2": "${namespace}*my.target.${x}",
			},
			Graphs: []Graph{
				{
					Name: "First graph",
					Expressions: map[string]string{
						"Timeseries 1": "var1 + var2",
					},
				},
			},
		})
	}

	key := DashboardKey(c, d)
	dashboard := &Dashboard{G: data}
	if _, err := datastore.Put(c, key, dashboard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Dashboard list =============================================

type DashboardResp struct {
	Name        string
	Description string
	Presets     map[string]Preset
}

type DashboardListResp struct {
	Dashboards []*DashboardResp
}

func ListDashboards(w http.ResponseWriter, r *http.Request) {
	ok, user, _ := users.IsAuthorized(r)
	if !ok {
		return
	}

	c := appengine.NewContext(r)
	q := datastore.NewQuery("Dash").Order("__key__")
	dashs := make([]*Dashboard, 0)
	keys, err := q.GetAll(c, &dashs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := &DashboardListResp{}
	for i, dash := range dashs {
		if authorized, err := dash.IsAcled(user.Email); err != nil || !authorized {
			continue
		}
		dashresp := &DashboardResp{
			Name: keys[i].StringID(),
		}
		if cfg, err := dash.Cfg(); err == nil {
			dashresp.Description = cfg.Description
			dashresp.Presets = cfg.Presets
		}
		resp.Dashboards = append(resp.Dashboards, dashresp)
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

// GetDashboard ==================================================

func GetDashboard(w http.ResponseWriter, r *http.Request) {
	ok, user, _ := users.IsAuthorized(r)
	if !ok {
		return
	}

	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	// Check if the dashboard already exists.
	dashboard := GetDashFromDatastore(c, d)
	if dashboard == nil {
		http.Error(w, "Dashboard doesn't exists", http.StatusBadRequest)
		return
	}

	obj := make(map[string]interface{})
	err = json.Unmarshal(dashboard.G, &obj)
	if err != nil {
		http.Error(w, "Error parsing dashboard config", http.StatusInternalServerError)
		return
	}

	if authorized, err := dashboard.IsAcled(user.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !authorized {
		http.Error(w, "Not authorized to edit this dashboard.",
			http.StatusUnauthorized)
		return
	}

	extraCfg, _ := json.MarshalIndent(obj, "", "  ")
	w.Write([]byte(extraCfg))
}

// Save dashboard ================================================

func SaveDashboard(w http.ResponseWriter, r *http.Request) {
	ok, user, _ := users.IsAuthorized(r)
	if !ok {
		return
	}

	// Check if the name is valid, and normalize the name.
	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the dashboard exists.
	c := appengine.NewContext(r)
	dashboard := GetDashFromDatastore(c, d)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

	// Check if the user can edit this dashboard.
	if authorized, err := dashboard.IsAcled(user.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !authorized {
		http.Error(w, "Not authorized to edit this dashboard.",
			http.StatusUnauthorized)
		return
	}

	// Make sure the data is valid json.
	rawData := []byte(r.FormValue("data"))
	if data, err := ValidateRawConfig(rawData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		dashboard.G, _ = json.Marshal(data)
	}

	// Check if the user can still edit the dashboard after it is
	// saved...
	if authorized, err := dashboard.IsAcled(user.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !authorized {
		http.Error(w, "You can't remove yourself from the ACL list.",
			http.StatusUnauthorized)
		return
	}

	// Ok.. write to datastore then!
	key := DashboardKey(c, d)
	if _, err := datastore.Put(c, key, dashboard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Delete dashboard =======================================

func DeleteDashboard(w http.ResponseWriter, r *http.Request) {
	ok, user, _ := users.IsAuthorized(r)
	if !ok {
		return
	}

	// Check if the name is valid, and normalize the name.
	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Check if the dashboard exists.
	c := appengine.NewContext(r)
	dashboard := GetDashFromDatastore(c, d)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

	// Check if the user can edit this dashboard.
	if authorized, err := dashboard.IsAcled(user.Email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !authorized {
		http.Error(w, "Not authorized to edit this dashboard.",
			http.StatusUnauthorized)
		return
	}

	key := DashboardKey(c, d)
	if err := datastore.Delete(c, key); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
