package timeseries

import (
	"log"

	"appengine"
)

var _ = log.Println

func GetData(c appengine.Context, req *GetReq) (*GetResp, error) {
	resp := &GetResp{}
	for _, serie := range req.Serie {
		res := selectTimeSlice(serie.R)
		pts, err := getPoints(c, res, serie.T, serie.To, serie.M)
		if err != nil {
			return nil, err
		}
		maxPoints := int(serie.To-serie.T) / int(res)
		s := &SerieData{
			Target:     serie.M,
			Datapoints: make([]*DataPoint, 0, maxPoints),
		}
		//log.Printf("Got %d points @%d, prepared for %d: %d\n",
		//    len(pts), int(res), maxPoints, serie.To - serie.T)
		for _, p := range pts {
			t := float64(p.t)
			v := &p.V
			s.Datapoints = append(
				s.Datapoints, &DataPoint{v, &t})
		}
		resp.Series = append(resp.Series, s)
	}
	return resp, nil
}

func getPoints(c appengine.Context, r TimeSlice, from, to int64, m string) ([]*P, error) {
	if int(r) == 1 {
		return getRawPoints(c, from, to, m)
	}
	return genAggregate(c, r, from, to, m, Avg)
}

func addMissing(from, to int64, res TimeSlice, s *SerieData) {
	r := int64(res)
	for at := from + r; at < to; at += r {
		a := float64(at)
		s.Datapoints = append(s.Datapoints, &DataPoint{nil, &a})
	}
}