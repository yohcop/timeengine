package points

import (
	"log"

	"timeengine/ae"
)

var _ = log.Println

func BuildFrames(c ae.Context, metric string, span *Span) ([]*frame, error) {
	if span.fs.IsRaw() {
		return nil, nil
	}

	// Get the frames at a higher resolution
	finerFrames, err := GetPoints(c, metric, span.HigherRes())
	if err != nil {
		return nil, err
	}

	newFrames := make([]*frame, 0, span.NumFrames())
	newKeys := make([]string, 0, span.NumFrames())

	start := span.from.Ts()
	end := span.to.Ts()
	step := int64(span.fs)
	for at := start; at <= end; at += step {
		relevantFrames := make([]StatsDataPoint, 0)
		for _, finer := range finerFrames {
			if finer.GetTs() >= at && finer.GetTs() < at+step {
				relevantFrames = append(relevantFrames, finer)
			}
		}
		// Now we have all the relevant higher-res frames for the
		// frame we are trying to build.
		kf := span.fs.KeyFrame(at)
		newFrame := newFrame(metric, span.fs, kf, relevantFrames)
		newFrames = append(newFrames, newFrame)
		newKeys = append(newKeys, aggregateKeyName(metric, span.fs, kf))
	}

	err = c.PutMulti("F", newKeys, newFrames)
	return newFrames, err
}
