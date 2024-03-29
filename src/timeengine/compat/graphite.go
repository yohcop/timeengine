package compat

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"time"

	"timeengine/ae/impl"
	"timeengine/timeseries"
	"timeengine/users"

	"appengine"
	"net/http"
)

var _ = log.Println

func Render(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := users.IsAuthorized(r); !ok {
		return
	}

	from, err := strconv.ParseInt(r.FormValue("from"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	now := time.Now().UnixNano() / 1000
	until, err := strconv.ParseInt(r.FormValue("until"), 10, 64)
	if err != nil || until > now {
		until = now
	}

	targets := r.Form["target"]
	jsonp := r.FormValue("jsonp")

	if len(targets) == 0 || until < from {
		log.Println("no targets or bad dates", from, until, len(targets))
		return
	}

	req := timeseries.GetReq{}

	// We only understand time spans in seconds. (second arg. to summarize)
	// example: summarize(foo.bar, "15s", "avg")
	re := regexp.MustCompile(
		"summarize\\(" +
			"\\W*([\\w\\*.-]+)\\W*," + // Metric name
			"\\W*\"(\\d+)s\"\\W*," + // Summary size (e.g. "60s")
			"\\W*\"(\\w+)\"\\W*\\)") // Summary function name (e.g. "avg")
	for _, t := range targets {
		match := re.FindStringSubmatch(t)
		r := int64(1)
		summary := "avg"
		if len(match) == 4 {
			r, err = strconv.ParseInt(match[2], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = match[1]
			summary = match[3]
		}
		s := &timeseries.SerieDef{
			R:  r * 1000000,
			T:  from,
			To: until,
			M:  t,
			S:  summary,
		}
		req.Serie = append(req.Serie, s)
	}

	c := appengine.NewContext(r)
	resp, err := timeseries.GetData(&impl.Appengine{c}, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s, _ := json.Marshal(resp.Series)
	if len(jsonp) > 0 {
		w.Write([]byte(jsonp))
		w.Write([]byte("("))
	}
	w.Write(s)
	if len(jsonp) > 0 {
		w.Write([]byte(");"))
	}
}
