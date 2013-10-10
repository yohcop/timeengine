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

const oneMinute = int64(60 * 1000000)

func Summarize60sTask(w http.ResponseWriter, r *http.Request) {
	c := &impl.Appengine{appengine.NewContext(r)}

	res := points.SelectSummarySize(oneMinute)

	// Fetch work.
	now := time.Now().Unix() * 1000000
	to := fmt.Sprintf("%d@9", res.SummaryKey(now).Ts()-oneMinute)

	objs := make([]points.MetricUpdate, 0)
	// Get the top ones. Since they are sorted by when they were
	// inserted, we get the oldest ones.
	keys, err := c.DsGetBetweenKeys(
		points.MetricUpdateDatastoreType, "", to, 1000, &objs)
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
					log.Println("Task error:", err.Error())
				} else {
					toRemove = append(toRemove, k)
				}
			}
			if len(metricKeys) >= 500 || len(toRemove) >= 100 {
				// 500 is the max. flush now.
				// toRemove grows slightly faster since there are 3 summary sizes.
				if err := c.PutMulti(points.MetricDatastoreType, metricKeys, metrics); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if err := c.DeleteMulti(points.MetricUpdateDatastoreType, toRemove); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				metrics = make([]*points.Metric, 0, len(keys))
				metricKeys = make([]string, 0, len(keys))
				toRemove = make([]string, 0, len(keys))
			}
		}
	}
	if err := c.PutMulti(points.MetricDatastoreType, metricKeys, metrics); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := c.DeleteMulti(points.MetricUpdateDatastoreType, toRemove); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
