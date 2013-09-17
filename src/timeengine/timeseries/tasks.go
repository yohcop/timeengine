package timeseries

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"timeengine/ae/impl"
	"timeengine/timeseries/points"

	"appengine"
)

var _ = log.Println

func Summarize60sTask(w http.ResponseWriter, r *http.Request) {
	c := &impl.Appengine{appengine.NewContext(r)}

	res := points.SelectSummarySize(60)

	// Fetch work.
	now := time.Now().Unix()
	to := fmt.Sprintf("%d@9999999999", int64(res.SummaryKey(now))-60)

	objs := make([]points.MetricUpdate, 0)
	// Get the top ones. Since they are sorted by when they were
	// inserted, we get the oldest ones.
	keys, err := c.DsGetBetweenKeys("MU", "", to, -1, &objs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	metrics := make([]*points.Metric, 0, len(keys))
	metricKeys := make([]string, 0, len(keys))

	toRemove := make([]string, 0, len(keys))
	for _, k := range keys {
		if _, summary, metric, err := points.MetricUpdateKeyDecode(k); err == nil {
			metrics = append(metrics, &points.Metric{})
			metricKeys = append(metricKeys, metric)
			// Only generate 1 summary for now. We can be smarter by agregating
			// multiple keys, but it will rarely be useful as we generate
			// those every minute anyway.
			for _, res := range points.AvailableSummarySizes[1:] {
				_, err := points.BuildSummaries(c, metric, points.NewSpan(res, summary, summary))
				if err != nil {
					log.Println(err.Error())
				} else {
					toRemove = append(toRemove, k)
				}
			}

			c.PutMulti("M", metricKeys, metrics)
		}
	}
	c.DeleteMulti("MU", toRemove)
}
