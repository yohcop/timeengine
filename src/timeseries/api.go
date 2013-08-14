package timeseries

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

  "users"
  "namespace"

	"appengine"
)

// Put ===================================================

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

func PutDataPoints(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
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
  if !namespace.VerifyNamespace(c, req.Ns, req.NsSecret) {
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
				m: namespace.MetricName(req.Ns, v.M),
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

// Get =========================================================

type SerieDef struct {
	R  int
	T  int64
	To int64
	M  string
}

type GetReq struct {
	Serie []*SerieDef
}

type DataPoint []*float64

type SerieData struct {
	Target     string       `json:"target"`
	Datapoints []*DataPoint `json:"datapoints"`
}

type GetResp struct {
	Series []*SerieData
}

func GetDataPoints(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		return
	}
	req := GetReq{}
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
	resp, err := GetData(c, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

