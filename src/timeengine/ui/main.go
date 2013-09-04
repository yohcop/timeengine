package ui

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"

	"timeengine/dashboard"
	"timeengine/users"

	"appengine"
)

var rootTmpl = template.Must(template.ParseGlob("timeengine/ui/*.html"))

type rootTmplData struct {
	User  *users.User
	Login string
}

type dashboardTmplData struct {
	Tpl    rootTmplData
	Name   string
	Graphs string
}

func Index(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "index", &rootTmplData{
		User:  user,
		Login: users.LogoutURL(appengine.NewContext(r)),
	})
}

func Dashboards(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "dashboards", &rootTmplData{
		User:  user,
		Login: users.LogoutURL(appengine.NewContext(r)),
	})
}

func DashboardEditor(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	d, err := dashboard.ValidDashboard(r.FormValue("dashboard"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	// Check if the dashboard already exists.
	dash := dashboard.GetDashFromDatastore(c, d)
	if dash == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}
	// Try go get a better formatted version.
	obj := make(map[string]interface{})
	err = json.Unmarshal(dash.G, &obj)
	if err != nil {
		http.Error(w, "Error parsing dashboard config: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jscfg, _ := json.MarshalIndent(obj, "", "  ")

	rootTmpl.ExecuteTemplate(w, "dashboard-editor", &dashboardTmplData{
		Tpl: rootTmplData{
			User:  user,
			Login: users.LogoutURL(appengine.NewContext(r)),
		},
		Name:   d,
		Graphs: string(jscfg),
	})
}

func Namespaces(w http.ResponseWriter, r *http.Request) {
	log.Println("foo")
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "namespaces", &rootTmplData{
		User:  user,
		Login: users.LogoutURL(appengine.NewContext(r)),
	})
}
