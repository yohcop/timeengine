package points

import (
	"fmt"
)

type Metric struct {
}

// Key = summary(updateTime, 60s)@summary(ts, 60s)@metricname
type MetricUpdate struct {
}

const MetricDatastoreType = "M"

// === MetricUpdate ==============================================

func MetricUpdateKey(ts int64, metric string, ss SummarySize) string {
	return fmt.Sprintf("%d@%d@%s", int64(ss.SummaryKey(ts)), int64(ss), metric)
}
