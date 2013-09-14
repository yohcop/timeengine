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

// An optional point.
type optPoint struct {
	V       float64 `datastore:"v,noindex"`
	Present bool    `datastore:"p,noindex"`
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

const frameDatastoreType = "F"

type frame struct {
	Stats stats `datastore:"s,noindex"`

	// Childrens stored only contain
	Children []stats `datastore:"c,noindex"`

	// The rest is NOT stored in the datastore or memcache.
	// Instead it can be derived from the key.
	metric string    // metric
	fs     FrameSize // resolution
	kf     KeyFrame  // timestamp
}

func aggregateKeyName(metric string, r FrameSize, kf KeyFrame) string {
	return fmt.Sprintf("%s@%d@%010d", metric, r, kf)
}

func decodeAggregateKeyName(key string) (
	metric string, fs FrameSize, kf KeyFrame, err error) {
	parts := strings.Split(key, "@")
	if len(parts) == 3 {
		metric = parts[0]
		ifs, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return "", 0, 0, err
		}
		ikf, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return "", 0, 0, err
		}
		return metric, FrameSize(ifs), KeyFrame(ikf), nil
	}
	return "", 0, 0, errors.New("Bad key")
}

func newFrame(metric string, fs FrameSize, kf KeyFrame,
	using []StatsDataPoint) *frame {
	s := *newStats(kf.Ts())
	children := make([]stats, 0)
	for _, f := range using {
		s.update(f.GetAvg(), f.GetSum(), f.GetMin(), f.GetMax())
		children = append(children, *copyStats(f))
	}
	return &frame{
		// TODO: Avg should be normalized based on duration of each value.
		Stats:    s,
		Children: children,

		metric: metric,
		fs:     fs,
		kf:     kf,
	}
}

func (f *frame) Key() string {
	return aggregateKeyName(f.metric, f.fs, f.kf)
}

func (f *frame) GetTs() int64    { return f.kf.Ts() }
func (f *frame) GetAvg() float64 { return f.Stats.Avg }
func (f *frame) GetSum() float64 { return f.Stats.Sum }
func (f *frame) GetMin() float64 { return f.Stats.Min }
func (f *frame) GetMax() float64 { return f.Stats.Max }
func (f *frame) HasStats() bool { return true }

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
