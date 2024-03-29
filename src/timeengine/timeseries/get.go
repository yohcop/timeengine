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
	// Some series may be duplicated since they may have different
	// summary functions (serie.S). So find the unique series.
	defs := make(map[string]*SerieDef)
	for _, serie := range req.Serie {
		k := encodeSerieDef(serie)
		defs[k] = serie
	}

	// Now request the data for each unique definition.
	data := make(map[string][]points.StatsDataPoint)
	dataMutex := sync.Mutex{}
	// Go routines put a nil in the done channel if everything went ok
	// otherwise they put an error.
	done := make(chan error)
	for key, serie := range defs {
		go func(k string, s *SerieDef) {
			res := points.SelectSummarySize(s.R)
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

	// Wait until all the go routines are done, and check for errors.
	for _ = range defs {
		err := <-done
		if err != nil {
			return nil, err
		}
	}

	// Compute the output for each requested Serie, using the
	// data retrieved by the go routines.
	for _, serie := range req.Serie {
		k := encodeSerieDef(serie)
		res := points.SelectSummarySize(serie.R)
		pts, _ := data[k]
		maxPoints := serie.To - serie.T
		if res > 1 {
			maxPoints /= int64(res)
		} else {
			// res is in microseconds. so if res >1, the division is
			// enough. Otherwise, we need to convert to seconds from
			// microseconds.
			maxPoints /= 1000000
		}
		s := &SerieData{
			Target:     serie.M,
			Datapoints: make([]*DataPoint, 0, int(maxPoints)),
		}
		if len(pts) > 0 && pts[0].HasStats() {
			s.Target += "@" + serie.S
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
