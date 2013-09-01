package dashboard

import (
	"encoding/json"
	"net/http"

	"timeengine/users"

	"appengine"
	"appengine/datastore"
)

func NewDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	// Check if the dashboard already exists.
	if GetDashboard(c, d) != nil {
		http.Error(w, "Dashboard exists", http.StatusBadRequest)
		return
	}

	key := DashboardKey(c, d)
	dashboard := &Dashboard{}
	if _, err := datastore.Put(c, key, dashboard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type DashboardResp struct {
	Name string
}

type DashboardListResp struct {
	Dashboards []*DashboardResp
}

func ListDashboards(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
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
	for i := range dashs {
		resp.Dashboards = append(resp.Dashboards, &DashboardResp{
			Name: keys[i].StringID(),
		})
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

func SaveDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	d, err := ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := make([]*Graph, 0)
	err = json.Unmarshal([]byte(r.FormValue("data")), &data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	dashboard := GetDashboard(c, d)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

	dashboard.G, _ = json.MarshalIndent(data, "", "\t")

	key := DashboardKey(c, d)
	if _, err := datastore.Put(c, key, dashboard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
