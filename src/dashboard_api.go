package timeengine

import (
	"encoding/json"
	"net/http"

	"appengine"
	"appengine/datastore"
)

func newDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
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
	if getDashboard(c, d) != nil {
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
	Name   string
}

type DashboardListResp struct {
	Dashboards []*DashboardResp
}

func listDashboards(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
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
			Name:   keys[i].StringID(),
		})
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

func saveDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
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
	dashboard := getDashboard(c, d)
  if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

  dashboard.G, _ = json.Marshal(data)

	key := DashboardKey(c, d)
	if _, err := datastore.Put(c, key, dashboard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
