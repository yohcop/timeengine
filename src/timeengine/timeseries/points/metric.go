package points

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Metric struct {
}

// Key = summary(updateTime, 60s)@summary(ts, 60s)@metricname
type MetricUpdate struct {
}

const MetricDatastoreType = "M"
const MetricUpdateDatastoreType = "MU"

// === MetricUpdate ==============================================

func MetricUpdateKey(now, ts int64, metric string) string {
	res := SelectSummarySize(60 * s)
	return fmt.Sprintf("%d@%d@%s", int64(res.SummaryKey(now)),
		int64(res.SummaryKey(ts)), metric)
}

func MetricUpdateKeyDecode(key string) (at, ts int64, metric string, err error) {
	parts := strings.Split(key, "@")
	if len(parts) == 3 {
		at, _ = strconv.ParseInt(parts[0], 10, 64)
		ts, _ = strconv.ParseInt(parts[1], 10, 64)
		m := parts[2]
		return at, ts, m, nil
	}
	return 0, 0, "", errors.New("Bad key")
}
