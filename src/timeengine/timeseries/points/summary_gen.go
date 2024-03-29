package points

import (
	"errors"
	"fmt"
	"log"

	"timeengine/ae"
)

var _ = log.Println

func BuildSummaries(c ae.Context, metric string, span *Span) ([]*summary, error) {
	if span.ss.IsRaw() {
		return nil, nil
	}

	// Get the summaries at a higher resolution
	finerSummaries, err := GetPoints(c, metric, span.HigherRes())
	if err != nil {
		return nil, err
	}

	newSummaries := make([]*summary, 0, span.NumSummaries())
	newKeys := make([]string, 0, span.NumSummaries())

	start := span.from.Ts()
	end := span.to.Ts()
	step := int64(span.ss)
	for at := start; at <= end; at += step {
		relevantSummaries := make([]StatsDataPoint, 0)
		for _, finer := range finerSummaries {
			if finer.GetTs() >= at && finer.GetTs() < at+step {
				relevantSummaries = append(relevantSummaries, finer)
			}
		}
		if len(relevantSummaries) == 0 {
			errStr := fmt.Sprintf("No relevant summaries for %s, "+
				"start=%d end=%d step=%d at=%d [got %d summaries]",
				metric, start, end, step, at, len(finerSummaries))
			c.Logf(errStr)
			// try again later.
			return nil, errors.New(errStr)
		}
		// Now we have all the relevant higher-res summaries for the
		// summary we are trying to build.
		sk := span.ss.SummaryKey(at)
		newSummary := newSummary(metric, span.ss, sk, relevantSummaries)
		newSummaries = append(newSummaries, newSummary)
		newKeys = append(newKeys, summaryKeyName(metric, span.ss, sk))
	}

	if len(newKeys) > 0 {
		err = c.PutMulti(summaryDatastoreType, newKeys, newSummaries)
	}
	return newSummaries, err
}
