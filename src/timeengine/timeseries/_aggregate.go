package timeseries

// Aggregates
//
// Aggregates combine a bunch of datapoints into a single entry
// in the datastore or memcache.

/*
import (
	"log"
	"time"

	"appengine"
)

var _ = log.Println

func genAggregate(c appengine.Context, ts TimeSlice, from, to int64, m string, f AggFunc) ([]*P, error) {
	if ts == 1 {
		return getRawPoints(c, from, to, m)
	}
	//log.Printf("GenAggregate @%d [%d, %d], %s\n", ts, from, to, m)
	keys := genCacheKeyIDs(m, ts, from, to)
	//log.Printf("%s: %d keys @%d to check from the cache\n",
	//    m, len(keys), ts)
	pts := []*P{}
	missing_keys := []string{}
	if ts.Memcached() {
		// First, try to get from memcache.
		pts, missing_keys = getPtsFromCache(c, keys)
	} else {
		missing_keys = keys
	}

	newPts := make([]*WrappedP, 0, len(missing_keys))
	lower := lowerTimeSlice(ts)
	now := time.Now().Unix()
	//log.Printf("We need to generate %d missing aggregates. res: %d, lower: %d\n",
	//	len(missing_keys), ts, lower)
	for _, k := range missing_keys {
		//log.Printf("Generating %s\n", k)
		_, _, t, err := decodeCacheKey(k)
		if err != nil {
			//log.Printf("Error for key %s: %s\n", k, err.Error())
			continue
		}
		to := t + int64(ts)
		lowerPts, _ := genAggregate(c, lower, t, to, m, f)
		v := f(lowerPts)
		wp := &WrappedP{k: k}
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
		// Do not store if in newPts if they are too close to <now>
		if t < now-int64(ts) {
			newPts = append(newPts, wp)
		}
	}

	// Save the newPts in memcache
	if ts.Memcached() {
		addPtsInCache(c, newPts)
	}
	//log.Printf("Number of points: %d\n", len(pts))
	//log.Printf("%#v\n", pts[0])
	sortPtsByDate(pts)
	return pts, nil
}
*/
