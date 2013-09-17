package points

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

var _ = log.Println

const (
	s = 1000000
	m = 60 * s
	h = 60 * m
)

// A SummarySize represent a resolution.
// To build a SummarySize, use SelectSummarySize(int)
type SummarySize int64

// A SummaryKey is a timestamp normalized at a given SummarySize.
// Use SummarySize.SummaryKey to make one.
type SummaryKey int64

type Span struct {
	ss       SummarySize
	from, to SummaryKey
}

var InvalidSummarySize SummarySize = SummarySize(-1)
var RawSummarySize SummarySize = SummarySize(0)
var InvalidSummaryKey SummaryKey = SummaryKey(-1)

// On 500px:
// 1 year: 12h intervals.
// 1 month: ~1.5h intervals.
// 1 week: ~20 minutes intervals.
// 1 day: ~2.5 minutes intervals.
var AvailableSummarySizes = []SummarySize{
	//      // =secs | per day
	0,      // 0.001 | n/a - as many as pushed.
	1 * m,  // 60    | 1440
	1 * h,  // 3600  | 24
	24 * h, // 86400 | 1
}

type stats struct {
	Ts  int64   `datastore:"t,noindex"`
	Avg float64 `datastore:"a,noindex"`
	Sum float64 `datastore:"s,noindex"`
	Min float64 `datastore:"i,noindex"`
	Max float64 `datastore:"x,noindex"`

	count  float64
	sumAvg float64
}

type StatsDataPoint interface {
	GetTs() int64

	HasStats() bool
	GetAvg() float64
	GetSum() float64
	GetMin() float64
	GetMax() float64
}

const summaryDatastoreType = "S"

type summary struct {
	Stats stats `datastore:"s,noindex"`

	Children []stats `datastore:"c,noindex"`

	// The rest is NOT stored in the datastore or memcache.
	// Instead it can be derived from the key.
	metric string      // metric
	ss     SummarySize // resolution
	sk     SummaryKey  // timestamp
}

// === SummarySize ==========================================

func SelectSummarySize(res int64) SummarySize {
	for i, ts := range AvailableSummarySizes[1:] {
		if res == int64(ts) {
			return ts
		}
		if res < int64(ts) {
			return AvailableSummarySizes[i]
		}
	}
	return AvailableSummarySizes[len(AvailableSummarySizes)-1]
}

func (r SummarySize) USecs() int64 {
	return int64(r)
}

func (r SummarySize) SummaryKey(ts int64) SummaryKey {
	if r <= 1 {
		return SummaryKey(ts)
	}
	return SummaryKey(ts - (ts % int64(r)))
}

func (r SummarySize) IsRaw() bool {
	return r == RawSummarySize
}

func (r SummarySize) IsValidNotRaw() bool {
	return r > RawSummarySize
}

func (r SummarySize) higherRes() SummarySize {
	if r == AvailableSummarySizes[0] {
		return r
	}
	for i, ts := range AvailableSummarySizes[1:] {
		if r == ts {
			return AvailableSummarySizes[i]
		}
	}
	return AvailableSummarySizes[0]
}

func (r SummarySize) lowerRes() SummarySize {
	var l = len(AvailableSummarySizes) - 1
	if r == AvailableSummarySizes[l] {
		return r
	}
	for i, ts := range AvailableSummarySizes {
		if r == ts && i < l-1 {
			return AvailableSummarySizes[i+1]
		}
	}
	return AvailableSummarySizes[l]
}

func (r SummarySize) numPoints() int {
	if r <= 1 {
		return 0
	}
	lower := r.lowerRes()
	return int(r / lower)
}

// === SummaryKey ===========================================

func (sk SummaryKey) Ts() int64 {
	return int64(sk)
}

// === Span ===============================================

func NewSpan(ss SummarySize, from, to int64) *Span {
	return &Span{ss, ss.SummaryKey(from), ss.SummaryKey(to)}
}

func (s *Span) HigherRes() *Span {
	l := s.ss.higherRes()
	from := l.SummaryKey(s.from.Ts())
	to := l.SummaryKey(s.to.Ts() + s.ss.USecs() - 1)
	return &Span{l, from, to}
}

func (s *Span) NumSummaries() int {
	return 1 + int((s.to.Ts()-s.from.Ts())/s.ss.USecs())
}

// === Summary ================================================

func summaryKeyName(metric string, r SummarySize, sk SummaryKey) string {
	// We convert the summary size in seconds to save bytes in the key.
	return fmt.Sprintf("%s@%d@%010d", metric, int64(r)/s, sk)
}

func decodeSummaryKeyName(key string) (
	metric string, ss SummarySize, sk SummaryKey, err error) {
	parts := strings.Split(key, "@")
	if len(parts) == 3 {
		metric = parts[0]
		iss, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", 0, 0, err
		}
		isk, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", 0, 0, err
		}
		return metric, SummarySize(iss * s), SummaryKey(isk), nil
	}
	return "", 0, 0, errors.New("Bad key")
}

func newSummary(metric string, ss SummarySize, sk SummaryKey,
	using []StatsDataPoint) *summary {
	s := newStats(sk.Ts())
	children := make([]stats, 0)
	for _, f := range using {
		s.update(f.GetAvg(), f.GetSum(), f.GetMin(), f.GetMax())
		children = append(children, *copyStats(f))
	}
	return &summary{
		// TODO: Avg should be normalized based on duration of each value.
		Stats:    *s,
		Children: children,

		metric: metric,
		ss:     ss,
		sk:     sk,
	}
}

func (f *summary) Key() string {
	return summaryKeyName(f.metric, f.ss, f.sk)
}

func (f *summary) GetTs() int64    { return f.sk.Ts() }
func (f *summary) GetAvg() float64 { return f.Stats.Avg }
func (f *summary) GetSum() float64 { return f.Stats.Sum }
func (f *summary) GetMin() float64 { return f.Stats.Min }
func (f *summary) GetMax() float64 { return f.Stats.Max }
func (f *summary) HasStats() bool  { return true }

// === stats ==============================================

func newStats(ts int64) *stats {
	return &stats{
		Ts:     ts,
		Avg:    0,
		Sum:    0,
		Min:    math.MaxFloat64,
		Max:    -math.MaxFloat64,
		count:  0,
		sumAvg: 0,
	}
}

func copyStats(s StatsDataPoint) *stats {
	return &stats{
		Ts:     s.GetTs(),
		Avg:    s.GetAvg(),
		Sum:    s.GetSum(),
		Min:    s.GetMin(),
		Max:    s.GetMax(),
		count:  0,
		sumAvg: 0,
	}
}

func (s *stats) update(avg, sum, min, max float64) *stats {
	s.sumAvg += avg
	s.Sum += sum
	s.Min = math.Min(s.Min, min)
	s.Max = math.Max(s.Max, max)

	s.count += 1
	s.Avg = s.sumAvg / s.count
	return s
}

func (s *stats) GetTs() int64    { return s.Ts }
func (s *stats) GetAvg() float64 { return s.Avg }
func (s *stats) GetSum() float64 { return s.Sum }
func (s *stats) GetMin() float64 { return s.Min }
func (s *stats) GetMax() float64 { return s.Max }
