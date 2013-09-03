package points

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Keep names short, saves db space on appengine.
//
// Stores data points.
//
// Offline processes can derive coarser resolutions as well. From 60 records at
// a resolution of 1 second, we can derive a 60 second average (or max, etc)
// point for the same metric name.
//
// The key is M@T, where M is actually "<namespace>*<metric name>"
// This way we can query for all the points in a range by key, and
// never need another index.
type P struct {
	// Value.
	V float64 `datastore:"v,noindex"`

	// The following unexported fields are not serialized, or stored
	// in memcache, since they can be derived from the key name.

	// Key
	k string
	// Timestamp, Unix time.
	t int64
	// Metric name.
	m string
	// Namespace.
	ns string
}

func NewP(value float64, timestamp int64, metric string) *P {
	return &P{V: value, t: timestamp, m: metric}
}

func (p *P) Timestamp() int64 {
	return p.t
}

func (p *P) Key() string {
	return keyAt(p.m, p.t)
}

func keyAt(m string, t int64) string {
	return fmt.Sprintf("%s@%d", m, t)
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

func (p *P) GetTs() int64    { return p.t }
func (p *P) GetAvg() float64 { return p.V }
func (p *P) GetSum() float64 { return p.V }
func (p *P) GetMin() float64 { return p.V }
func (p *P) GetMax() float64 { return p.V }
