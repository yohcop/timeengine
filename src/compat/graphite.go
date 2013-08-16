package compat

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"

	"dashboard"
	"timeseries"
	"users"

	"appengine"
	"net/http"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	re := regexp.MustCompile("/dashboard/load/(.*)")
	match := re.FindStringSubmatch(r.URL.Path)
	name := match[1]

	c := appengine.NewContext(r)
	dash := dashboard.GetDashboard(c, name)
	if dash == nil {
		http.Error(w, "Dashboard not found", http.StatusBadRequest)
		return
	}

	data := make([]*dashboard.Graph, 0)
	err = json.Unmarshal(dash.G, &data)
	if err != nil {
		http.Error(w, "Dashboard not found", http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{
   "state":{
      "name":"` + name + `",
      "graphs":[
         `))
	for i, g := range data {
		w.Write([]byte(`
         [
            "` + g.Name + `",
            {
               "target":["` + strings.Join(g.Targets, `","`) + `"]
            }
         ]`))
		if i < len(data)-1 {
			w.Write([]byte(","))
		}
	}
	w.Write([]byte(`
      ]
    }
  }
  `))
}

func Render(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	from, err := strconv.ParseInt(r.FormValue("from"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().Unix()
	until, err := strconv.ParseInt(r.FormValue("until"), 10, 64)
	if err != nil || until > now {
		until = now
	}

	targets := r.Form["target"]
	jsonp := r.FormValue("jsonp")

	if len(targets) == 0 || until < from {
		return
	}

	req := timeseries.GetReq{}

	// We only understand time spans in seconds. (second arg. to summarize)
	// example: summarize(foo.bar, "15s", "avg")
	re := regexp.MustCompile(
		"summarize\\(\\W*([\\w\\*.-]*)\\W*,\\W*\"([\\d]+)s\"\\W*,.*\\)")
	for _, t := range targets {
		match := re.FindStringSubmatch(t)
		r := 1
		if len(match) == 3 {
			r, err = strconv.Atoi(match[2])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = match[1]
		}
		s := &timeseries.SerieDef{
			R:  r,
			T:  from,
			To: until,
			M:  t,
		}
		req.Serie = append(req.Serie, s)
	}

	c := appengine.NewContext(r)
	resp, err := timeseries.GetData(c, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s, _ := json.MarshalIndent(resp.Series, "", "  ")
	if len(jsonp) > 0 {
		w.Write([]byte(jsonp))
		w.Write([]byte("("))
	}
	w.Write(s)
	if len(jsonp) > 0 {
		w.Write([]byte(");"))
	}
}
