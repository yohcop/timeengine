package timeengine

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"appengine"
	"net/http"
)

func dashboard(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}

	w.Write([]byte(`
  {
   "state":{
      "name":"ycoppel1",
      "graphs":[
         [
            "target=foo.bar.1",
            {
               "target":[
                  "foo.bar.1"
               ]
            }
         ],
         [
            "target=foo.bar.2",
            {
               "target":[
                  "foo.bar.2"
               ]
            }
         ],
         [
            "target=foo.bar.1&target=foo.bar.2",
            {
               "target":[
                  "foo.bar.1",
                  "foo.bar.2"
               ]
            }
         ]
      ]
    }
  }
  `))
}

func render(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
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

	req := GetReq{}

  // We only understand time spans in seconds. (second arg. to summarize)
  // example: summarize(foo.bar, "15s", "avg")
	re := regexp.MustCompile(
		"summarize\\(\\W*([\\w.-]*)\\W*,\\W*\"([\\d]+)s\"\\W*,.*\\)")
	for _, t := range targets {
		match := re.FindStringSubmatch(t)
		w.Write([]byte(fmt.Sprintf("// %q\n", match)))
		r := 1
		if len(match) == 3 {
			r, err = strconv.Atoi(match[2])
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			t = match[1]
		}
		s := &SerieDef{
			R:  r,
			T:  from,
			To: until,
			M:  t,
		}
		req.Serie = append(req.Serie, s)
	}

	c := appengine.NewContext(r)
	resp, err := getData(c, &req)
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
