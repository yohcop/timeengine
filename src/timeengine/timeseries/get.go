package timeseries

import (
  "log"

	"timeengine/ae"
	"timeengine/timeseries/points"
)

var _ = log.Println

func GetData(c ae.Context, req *GetReq) (*GetResp, error) {
	resp := &GetResp{}
	for _, serie := range req.Serie {
    log.Println(serie.S)

    summaryFn := points.GetSummarySelector(serie.S)
		res := points.SelectFrameSize(serie.R)
		pts, err := points.GetPoints(c, serie.M,
        points.NewSpan(res, serie.T, serie.To))
		if err != nil {
			return nil, err
		}
		maxPoints := int(serie.To-serie.T)
    if res > 1 {
      maxPoints /= int(res)
    }
		s := &SerieData{
			Target:     serie.M + "@" + serie.S,
			Datapoints: make([]*DataPoint, 0, maxPoints),
		}
		//log.Printf("Got %d points @%d, prepared for %d: %d\n",
		//    len(pts), int(res), maxPoints, serie.To - serie.T)
		for _, p := range pts {
			t := float64(p.GetTs())
			v := summaryFn(p)
			s.Datapoints = append(s.Datapoints, &DataPoint{&v, &t})
		}
		resp.Series = append(resp.Series, s)
	}
	return resp, nil
}
