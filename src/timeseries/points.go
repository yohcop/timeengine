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
	//      // secs  | /day
	1 * s, // 1     |

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

func keyAt(m string, t int64) string {
	return fmt.Sprintf("%s@%d", m, t)
}

func key(c appengine.Context, m string, t int64) *datastore.Key {
	return keyStr(c, keyAt(m, t))
}

func keyStr(c appengine.Context, k string) *datastore.Key {
	return datastore.NewKey(c, "P", k, 0, nil)
}

func decodePointKey(k *datastore.Key, p *P) {
	decodePointStrKey(k.StringID(), p)
}

func decodeKey(k string) (m string, t int64, err error) {
	parts := strings.Split(k, "@")
	if len(parts) == 2 {
		m = parts[0]
		t, _ = strconv.ParseInt(parts[1], 10, 64)
		return m, t, nil
	}
	return "", 0, errors.New("Bad key")
}

func decodePointStrKey(k string, p *P) {
	m, t, err := decodeKey(k)
	if err == nil {
		p.k = k
		p.m = m
		p.t = t
	}
}

func putRawPoints(c appengine.Context, pts []*P) error {
	keys := make([]*datastore.Key, 0, len(pts))
	for _, p := range pts {
		keys = append(keys, key(c, p.m, p.t))
	}
	_, err := datastore.PutMulti(c, keys, pts)
	if err != nil {
		return err
	}
	return nil
}

func getRawPoints(c appengine.Context, from, to int64, m string) ([]*P, error) {
	q := datastore.NewQuery("P")
	q = q.Order("__key__")
	q = q.Filter("__key__ >=", key(c, m, from))
	q = q.Filter("__key__ <=", key(c, m, to))
	q = q.Limit(-1)

	pts := make([]*P, 0)
	keys, err := q.GetAll(c, &pts)
	if err != nil {
		return nil, err
	}
	for i, p := range pts {
		decodePointKey(keys[i], p)
	}
	return pts, nil
}
