package timeseries

import (
	"errors"
	"fmt"
	"log"
	"sync"

	"timeengine/ae"
	"timeengine/timeseries/points"
)

var _ = log.Println

func encodeSerieDef(def *SerieDef) string {
	return fmt.Sprintf("%s@%d@%d@%d", def.M, def.R, def.T, def.To)
}

func GetData(c ae.Context, req *GetReq) (*GetResp, error) {
	resp := &GetResp{}
	// TODO: make each call in a separate go routine.
	defs := make(map[string]*SerieDef)
	data := make(map[string][]points.StatsDataPoint)
	// Some series may be duplicated since they may have different
	// summary functions (serie.S). So find the unique series.
	for _, serie := range req.Serie {
		k := encodeSerieDef(serie)
		defs[k] = serie
		data[k] = nil
	}
	// Now request the data for each unique definition.
	dataMutex := sync.Mutex{}
	// Go routines put a nil in the done channel if everything went ok
	// otherwise they put an error.
	done := make(chan error)
	for key, serie := range defs {
		go func(k string, s *SerieDef) {
			res := points.SelectFrameSize(s.R)
			pts, err := points.GetPoints(c, s.M,
				points.NewSpan(res, s.T, s.To))
			if err != nil {
				done <- errors.New("Error while getting data: " + err.Error())
				return
			}
			dataMutex.Lock()
			data[k] = pts
			dataMutex.Unlock()
			done <- nil
		}(key, serie)
	}
	// Wait until all the go routines are done.
	for _ = range defs {
		err := <-done
		if err != nil {
			return nil, err
		}
	}

	// Compute the outpug for each requested Serie, using the
	// data retrieved.
	for _, serie := range req.Serie {
		k := encodeSerieDef(serie)
		res := points.SelectFrameSize(serie.R)
		pts, _ := data[k]
		maxPoints := int(serie.To - serie.T)
		if res > 1 {
			maxPoints /= int(res)
		}
		s := &SerieData{
			Target:     serie.M + "@" + serie.S,
			Datapoints: make([]*DataPoint, 0, maxPoints),
		}
		summaryFn := points.GetSummarySelector(serie.S)
		for _, p := range pts {
			t := float64(p.GetTs())
			v := summaryFn(p)
			s.Datapoints = append(s.Datapoints, &DataPoint{&v, &t})
		}
		resp.Series = append(resp.Series, s)
	}
	return resp, nil
}
