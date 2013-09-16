package points

import (
	"log"

	"timeengine/ae"
)

var _ = log.Println

func GetPoints(c ae.Context, metric string, span *Span) ([]StatsDataPoint, error) {
	if span.ss.USecs() == 0 || span.to.Ts()-span.from.Ts() <= 120*s {
		return getRawPoints(c, metric, span.from.Ts(), span.to.Ts())
	}
	return getFromSummaries(c, metric, span)
}

func getRawPoints(c ae.Context, metric string, from, to int64) ([]StatsDataPoint, error) {
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

func getFromSummaries(c ae.Context, metric string, span *Span) (
	[]StatsDataPoint, error) {
	pts := make([]*summary, 0)
	keys, err := c.DsGetBetweenKeys(summaryDatastoreType,
		summaryKeyName(metric, span.ss, span.from),
		summaryKeyName(metric, span.ss, span.to),
		-1, &pts)
	if err != nil {
		return nil, err
	}
	stats := make([]StatsDataPoint, len(pts))
	for i, p := range pts {
		metric, ssize, sk, err := decodeSummaryKeyName(keys[i])
		if err != nil {
			return nil, err
		}
		p.metric = metric
		p.ss = ssize
		p.sk = sk
		stats[i] = p
	}
	return stats, nil
}
