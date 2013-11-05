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

var _ = log.Println

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

type pushTmplData struct {
	Tpl       rootTmplData
	Namespace string
	NsSecret  string
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
	jscfg := make([]byte, 0)
	if err = json.Unmarshal(dash.G, &obj); err == nil {
		jscfg, _ = json.MarshalIndent(obj, "", "  ")
	} else {
		jscfg = dash.G
	}

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
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "namespaces", &rootTmplData{
		User:  user,
		Login: users.LogoutURL(appengine.NewContext(r)),
	})
}

func PushPage(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	if len(r.FormValue("ns")) == 0 {
		rootTmpl.ExecuteTemplate(w, "push-select", &pushTmplData{
			Tpl: rootTmplData{
				User:  user,
				Login: users.LogoutURL(appengine.NewContext(r)),
			},
		})
	} else {
		rootTmpl.ExecuteTemplate(w, "push", &pushTmplData{
			Tpl: rootTmplData{
				User:  user,
				Login: users.LogoutURL(appengine.NewContext(r)),
			},
			Namespace: r.FormValue("ns"),
			NsSecret:  r.FormValue("nssecret"),
		})
	}
}
