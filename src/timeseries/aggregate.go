package timeseries

import (
	"log"
  "time"

	"appengine"
)

var _ = log.Println

type AggFunc func([]*P) *float64

func Avg(pts []*P) *float64 {
  if len(pts) == 0 {
    return nil
  }
	s := 0.0
	for _, n := range pts {
		s += n.V
	}
	res := s / float64(len(pts))
	return &res
}

func genAggregate(c appengine.Context, ts TimeSlice, from, to int64, m string, f AggFunc) ([]*P, error) {
	if ts == 1 {
		return getRawPoints(c, ts, from, to, m)
	}
  //log.Printf("GenAggregate @%d [%d, %d], %s\n", ts, from, to, m)
	keys := genKeyIDs(m, ts, from, to)
  pts := []*P{}
  missing_keys := []string{}
  if (ts.Memcached()) {
	  // First, try to get from memcache.
	  pts, missing_keys = getPtsFromCache(c, keys)
  } else {
    missing_keys = keys
  }

  log.Printf("We need to generate %d missing aggregates\n",
      len(missing_keys))
	newPts := make([]*WrappedP, 0, len(missing_keys))
	lower := lowerTimeSlice(ts)
  now := time.Now().Unix()
	for _, k := range missing_keys {
		_, _, t, _ := decodeKey(k)
		to := t + int64(ts)
		lowerPts, _ := genAggregate(c, lower, t, to, m, f)
		v := f(lowerPts)
    wp := &WrappedP{k:k}
		if v != nil {
			pt := &P{
				V: *v,
        t: t,
			}
      wp.HasPt = true
      wp.P = pt
			pts = append(pts, pt)
    } else {
      wp.HasPt = false
    }
    // TODO: Do not store if in newPts if they are too close to
    // <now>
    if t < now - int64(ts) {
  	  newPts = append(newPts, wp)
    }
	}

	// Save the newPts in memcache
  if (ts.Memcached()) {
	  addPtsInCache(c, newPts)
  }
  log.Printf("Number of points: %d\n", len(pts))
  sortPtsByDate(pts)
	return pts, nil
}
