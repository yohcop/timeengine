package timeengine

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"appengine"
)

type Vals struct {
	M string
	V float64
}

type Points struct {
	R        int
	T        int64
	Vs       []*Vals
}

type PutReq struct {
	Ns       string
	NsSecret string
	Pts []*Points
}

func put(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
	if user == nil || err != nil {
		return
	}
	req := PutReq{}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = json.Unmarshal(body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := appengine.NewContext(r)
  if !VerifyNamespace(c, req.Ns, req.NsSecret) {
		http.Error(w, "Missing or unknown namespace/secret",
        http.StatusUnauthorized)
		return
  }

	ps := make([]*P, 0)
	for _, p := range req.Pts {
		for _, v := range p.Vs {
			p := &P{
				V: v.V,
				t: p.T,
				r: p.R,
				m: MetricName(req.Ns, v.M),
			}
			ps = append(ps, p)
		}
	}

	err = putRawPoints(c, ps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
