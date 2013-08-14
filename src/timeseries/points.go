package timeseries

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"appengine"
	"appengine/datastore"
)

const (
	s = 1
	m = 60
	h = 60 * 60
)

// To build a TimeSlice, use selectTimeSlice(int)
type TimeSlice int

// On 500px:
// 1 year: 12h intervals.
// 1 month: ~1.5h intervals.
// 1 week: ~20 minutes intervals.
// 1 day: ~2.5 minutes intervals.
var AvailableAggregatesSecs = []TimeSlice{
	1 * s,

	1 * m,  // 60    | 1440 => 7200
	15 * m, // 900   | 360  => 7560

	1 * h,  // 3600  | 24   => 7584
	12 * h, // 43200 | 2
}

func (t *TimeSlice) Memcached() bool {
	return true //int(*t) > 1*m
}

func selectTimeSlice(res int) TimeSlice {
	for i, ts := range AvailableAggregatesSecs[1:] {
		if res < int(ts) {
			return AvailableAggregatesSecs[i]
		}
	}
	return AvailableAggregatesSecs[len(AvailableAggregatesSecs)-1]
}

func lowerTimeSlice(res TimeSlice) TimeSlice {
	if res == AvailableAggregatesSecs[0] {
		return res
	}
	for i, ts := range AvailableAggregatesSecs[1:] {
		if res == ts {
			return AvailableAggregatesSecs[i]
		}
	}
	return AvailableAggregatesSecs[0]
}

func keyAt(m string, r TimeSlice, t int64) string {
	return fmt.Sprintf("%s@%d@%d", m, r, slice(r, t))
}

func key(c appengine.Context, m string, r TimeSlice, t int64) *datastore.Key {
	return keyStr(c, keyAt(m, r, t))
}

func keyStr(c appengine.Context, k string) *datastore.Key {
	return datastore.NewKey(c, "P", k, 0, nil)
}

// Generate all the keys in between from and to.
func genKeyIDs(m string, r TimeSlice, from, to int64) []string {
	last := slice(r, to)
	keys := make([]string, 0, numPoints(from, to, r))
	for at := slice(r, from); at < last; at += int64(r) {
		keys = append(keys, keyAt(m, r, at))
	}
	return keys
}

func slice(r TimeSlice, at int64) int64 {
	return at - (at % int64(r))
}

func numPoints(from, to int64, res TimeSlice) int {
	return int(slice(res, to)-slice(res, from)) / int(res)
}

func decodePointKey(k *datastore.Key, p *P) {
	decodePointStrKey(k.StringID(), p)
}

func decodeKey(k string) (m string, r int, t int64, err error) {
	parts := strings.Split(k, "@")
	if len(parts) == 3 {
		m = parts[0]
		r, _ = strconv.Atoi(parts[1])
		t, _ = strconv.ParseInt(parts[2], 10, 64)
		return m, r, t, nil
	}
	return "", 0, 0, errors.New("Bad key")
}

func decodePointStrKey(k string, p *P) {
	m, r, t, err := decodeKey(k)
	if err == nil {
		p.k = k
		p.m = m
		p.r = r
		p.t = t
	}
}

func putRawPoints(c appengine.Context, pts []*P) error {
	keys := make([]*datastore.Key, 0, len(pts))
	for _, p := range pts {
		keys = append(keys, key(c, p.m, selectTimeSlice(p.r), p.t))
	}
	_, err := datastore.PutMulti(c, keys, pts)
	if err != nil {
		return err
	}
	return nil
}

func getRawPoints(c appengine.Context, r TimeSlice, t, to int64, m string) ([]*P, error) {
	limit := numPoints(t, to, r)
	q := datastore.NewQuery("P").Order("__key__")
	q = q.Filter("__key__ >=", key(c, m, r, t))
	q = q.Filter("__key__ <=", key(c, m, r, to))
	q = q.Limit(limit)

	pts := make([]*P, 0, limit)
	keys, err := q.GetAll(c, &pts)
	if err != nil {
		return nil, err
	}
	for i, p := range pts {
		decodePointKey(keys[i], p)
	}
	return pts, nil
}
