package timeseries

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"timeengine/ae/impl"
	"timeengine/namespace"
	"timeengine/timeseries/points"
	"timeengine/users"

	"appengine"
	"appengine/delay"
)

var _ = log.Println

// Put ===================================================

type Points struct {
	T int64
	M string
	V float64
}

type PutReq struct {
	Ns       string
	NsSecret string
	Pts      []*Points
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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	c := appengine.NewContext(r)
	if !namespace.VerifyNamespace(c, req.Ns, req.NsSecret) {
		http.Error(w, "Missing or unknown namespace/secret",
			http.StatusUnauthorized)

		c.Errorf("Missing or unknown namespace/secret: %s, %s",
			req.Ns, req.NsSecret)
		return
	}

	delayInputProcess.Call(c, &req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ProcessInput(c appengine.Context, req *PutReq) error {
	ps := make([]*points.P, 0)
	for _, p := range req.Pts {
		p := points.NewP(p.V, p.T, namespace.MetricName(req.Ns, p.M))
		ps = append(ps, p)
	}

	return points.PutRawPoints(&impl.Appengine{c}, ps)
}

// Warining: if the string is changed (name of the delay function),
// pending operations will break when the new version is pushed to
// appengine.
var delayInputProcess = delay.Func("ProcessInput", ProcessInput)

// Get =========================================================

type SerieDef struct {
	R  int64
	T  int64
	To int64
	M  string
	S  string
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
	resp, err := GetData(&impl.Appengine{c}, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}
