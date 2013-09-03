package points

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Metric struct {
	Name string `datastore:"m,noindex"`
}

// Key = frame(updateTime, 60s)@frame(ts, 60s)@metricname
type MetricUpdate struct {
}

func MetricUpdateKey(now, ts int64, metric string) string {
	res := SelectFrameSize(60 * s)
	return fmt.Sprintf("%d@%d@%s", int64(res.KeyFrame(now)),
		int64(res.KeyFrame(ts)), metric)
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
