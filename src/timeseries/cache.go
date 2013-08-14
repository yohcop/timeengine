package timeseries

import (
	"appengine"
	"appengine/memcache"
	"log"
)

var _ = log.Println

type WrappedP struct {
	P     *P
	HasPt bool

  k string
}

func getPtsFromCache(c appengine.Context, keys []string) (pts []*P, missing_keys []string) {
	items, err := memcache.GetMulti(c, keys)
	if err != nil {
		return
	}
	pts = make([]*P, 0, len(items))
	for k, item := range items {
		wp := &WrappedP{}
		memcache.Gob.Unmarshal(item.Value, wp)
		if wp.HasPt {
			decodePointStrKey(k, wp.P)
			pts = append(pts, wp.P)
		}
	}

	// In missing, we don't include the ones that had no point in the
	// wrapping object, since we don't want to try to recompute them.
	missing := make([]string, 0, len(keys)-len(pts))
	for _, k := range keys {
		if _, ok := items[k]; !ok {
			missing = append(missing, k)
		}
	}
	return pts, missing
}

func addPtsInCache(c appengine.Context, pts []*WrappedP) {
	items := make([]*memcache.Item, 0, len(pts))
	for _, wp := range pts {
		if v, err := memcache.Gob.Marshal(wp); err == nil {
			i := &memcache.Item{
				//Key:   keyAt(p.m, TimeSlice(p.r), p.t),
				Key:   wp.k,
				Value: v,
			}
			items = append(items, i)
		}
	}
	memcache.AddMulti(c, items)
}
