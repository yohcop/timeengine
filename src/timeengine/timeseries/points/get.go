package points

import (
	"log"

	"timeengine/ae"
)

var _ = log.Println

func GetPoints(c ae.Context, metric string, span *Span) ([]StatsDataPoint, error) {
	if span.fs.USecs() == 0 || span.to.Ts()-span.from.Ts() <= 120*s {
		return getRawPoints(c, metric, span.from.Ts(), span.to.Ts())
	}
	return getFromFrames(c, metric, span)
}

func getRawPoints(c ae.Context, metric string, from, to int64) ([]StatsDataPoint, error) {
	log.Println("getRawPoints", from, to)
	pts := make([]*P, 0)
	keys, err := c.DsGetBetweenKeys("P",
		keyAt(metric, from), keyAt(metric, to), -1, &pts)
	if err != nil {
		log.Println("Error there", keyAt(metric, from), err.Error())
		return nil, err
	}
	stats := make([]StatsDataPoint, len(pts))
	for i, p := range pts {
		decodePointStrKey(keys[i], p)
		stats[i] = p
	}
	return stats, nil
}

func getFromFrames(c ae.Context, metric string, span *Span) (
	[]StatsDataPoint, error) {
	log.Println("getFromFrames", span.from, span.to)

	pts := make([]*frame, 0)
	keys, err := c.DsGetBetweenKeys("F",
		aggregateKeyName(metric, span.fs, span.from),
		aggregateKeyName(metric, span.fs, span.to),
		-1, &pts)
	if err != nil {
		return nil, err
	}
	stats := make([]StatsDataPoint, len(pts))
	for i, p := range pts {
		metric, fsize, kf, err := decodeAggregateKeyName(keys[i])
		if err != nil {
			return nil, err
		}
		p.metric = metric
		p.fs = fsize
		p.kf = kf
		stats[i] = p
	}
	return stats, nil
}
