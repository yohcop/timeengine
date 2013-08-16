package timeseries

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

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

func cacheKeyAt(m string, r TimeSlice, t int64) string {
	return fmt.Sprintf("%s@%d@%d", m, r, slice(r, t))
}

func decodeCacheKey(k string) (m string, r int, t int64, err error) {
	parts := strings.Split(k, "@")
	if len(parts) == 3 {
		m = parts[0]
		r, _ = strconv.Atoi(parts[1])
		t, _ = strconv.ParseInt(parts[2], 10, 64)
		return m, r, t, nil
	}
	return "", 0, 0, errors.New("Bad key")
}

func decodeCachePointStrKey(k string, p *P) {
	m, _, t, err := decodeCacheKey(k)
	if err == nil {
		p.k = k
		p.m = m
		p.t = t
	}
}

// Generate all the memcache keys in between from and to at the given
// resolution
func genCacheKeyIDs(m string, r TimeSlice, from, to int64) []string {
	last := slice(r, to)
	keys := make([]string, 0, numPoints(from, to, r))
	for at := slice(r, from); at < last; at += int64(r) {
		keys = append(keys, cacheKeyAt(m, r, at))
	}
	return keys
}

func slice(r TimeSlice, at int64) int64 {
	return at - (at % int64(r))
}

func numPoints(from, to int64, res TimeSlice) int {
	return int(slice(res, to)-slice(res, from)) / int(res)
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
			decodeCachePointStrKey(k, wp.P)
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
