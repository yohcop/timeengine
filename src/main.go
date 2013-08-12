package timeengine

import (
	"html/template"
	"net/http"

	"appengine"
)

var rootTmpl = template.Must(template.New("index.html").
	ParseFiles("index.html"))

func init() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/put", put)
	http.HandleFunc("/get", get)
	http.HandleFunc("/namespace/new/", newNs)
	http.HandleFunc("/namespace/list/", listNs)

	// Backward compatible with graphite:
	// get a dashboard, and a tiny subset of the json renderer.
	http.HandleFunc("/dashboard/load/", dashboard)
	http.HandleFunc("/render/", render)
}

type rootTmplData struct {
	User  *User
	Login string
}

func handler(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	rootTmpl.Execute(w, &rootTmplData{
		User:  user,
		Login: LogoutURL(appengine.NewContext(r)),
	})
}
