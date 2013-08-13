package timeengine

import (
	"html/template"
	"log"
	"net/http"

	"appengine"
)

var rootTmpl = template.Must(template.ParseGlob("*.html"))

func init() {
	// Us pages.
	http.HandleFunc("/", handler)
	http.HandleFunc("/dashboards", dashboards)
	http.HandleFunc("/dashboard/edit", dashboardEditor)
	http.HandleFunc("/namespaces", namespaces)

	// Api stuff. Doesn't render a UI, but handles ajax calls.
	http.HandleFunc("/api/put", put)
	http.HandleFunc("/api/get", get)

	http.HandleFunc("/api/namespace/new/", newNs)
	http.HandleFunc("/api/namespace/list/", listNs)

	http.HandleFunc("/api/dashboard/new/", newDashboard)
	http.HandleFunc("/api/dashboard/list/", listDashboards)
	http.HandleFunc("/api/dashboard/save/", saveDashboard)

	// Backward compatible with graphite:
	// get a dashboard, and a tiny subset of the json renderer.
	http.HandleFunc("/dashboard/load/", dashboard)
	http.HandleFunc("/render/", render)
}

type rootTmplData struct {
	User  *User
	Login string
}

type dashboardTmplData struct {
	rootTmplData
	Name   string
	Graphs string
}

func handler(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "index", &rootTmplData{
		User:  user,
		Login: LogoutURL(appengine.NewContext(r)),
	})
}

func dashboards(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "dashboards", &rootTmplData{
		User:  user,
		Login: LogoutURL(appengine.NewContext(r)),
	})
}

func dashboardEditor(w http.ResponseWriter, r *http.Request) {
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
	dashboard := getDashboard(c, d)
	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

	rootTmpl.ExecuteTemplate(w, "dashboard-editor", &dashboardTmplData{
		rootTmplData: rootTmplData{
			User:  user,
			Login: LogoutURL(appengine.NewContext(r)),
		},
		Name:   d,
		Graphs: string(dashboard.G),
	})
}

func namespaces(w http.ResponseWriter, r *http.Request) {
	log.Println("foo")
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.ExecuteTemplate(w, "namespaces", &rootTmplData{
		User:  user,
		Login: LogoutURL(appengine.NewContext(r)),
	})
}
