package timeengine

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"appengine"
	"net/http"
)

var _ = log.Println

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

func get(w http.ResponseWriter, r *http.Request) {
	user, err := AuthUser(w, r)
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
	resp, err := getData(c, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

func getData(c appengine.Context, req *GetReq) (*GetResp, error) {
	resp := &GetResp{}
	for _, serie := range req.Serie {
    res := selectTimeSlice(serie.R)
		pts, err := getPoints(c, res, serie.T, serie.To, serie.M)
		if err != nil {
			return nil, err
		}
		s := &SerieData{
			Target: serie.M,
			Datapoints: make([]*DataPoint, 0,
				int(serie.To-serie.T)/int(res)),
		}
		last := serie.T
		for _, p := range pts {
			addMissing(last, p.t, res, s)
			t := float64(p.t)
			v := &p.V
			s.Datapoints = append(
				s.Datapoints, &DataPoint{v, &t})
			last = p.t
		}
		addMissing(last, serie.To+int64(res), res, s)
		resp.Series = append(resp.Series, s)
	}
	return resp, nil
}

func getPoints(c appengine.Context, r TimeSlice, t, to int64, m string) ([]*P, error) {
	if int(r) == 1 {
		return getRawPoints(c, r, t, to, m)
	}
	return genAggregate(c, r, t, to, m, Avg)
}

func addMissing(from, to int64, res TimeSlice, s *SerieData) {
	r := int64(res)
	for at := from + r; at < to; at += r {
    log.Printf("%d From %d-%d, %d\n", at, from, to, res);
		a := float64(at)
		s.Datapoints = append(s.Datapoints, &DataPoint{nil, &a})
	}
}
