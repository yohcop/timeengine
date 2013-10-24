package timeseries

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"timeengine/ae"
	"timeengine/ae/impl"
	"timeengine/timeseries/points"

	"appengine"
)

var _ = log.Println

const SummarizeQueueUrl = "/queue/summarize"
const oneMinute = int64(60 * 1000000)

func SummarizeCron(w http.ResponseWriter, r *http.Request) {
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

	maxBeforeFlush := 100
	metrics := make([]*points.Metric, 0, len(keys))
	metricKeys := make([]string, 0, maxBeforeFlush)
	toRemove := make([]string, 0, maxBeforeFlush)
	tasks := make([]url.Values, 0, maxBeforeFlush)
	for _, k := range keys {
		if _, _, metric, err := points.MetricUpdateKeyDecode(k); err == nil {
			// The points.Metric is basically empty. all is contained in the key.
			metrics = append(metrics, &points.Metric{})
			// Note: we may have duplicated metricKeys.
			// TODO: dedupe.
			metricKeys = append(metricKeys, metric)
			toRemove = append(toRemove, k)
			v := url.Values{}
			v.Set("m", k)
			tasks = append(tasks, v)
		}
		if len(metrics) >= maxBeforeFlush {
			// 500 is the max. flush now.
			// toRemove grows slightly faster since there are 3 summary sizes.
			if err := addTasks(c, tasks, metrics, metricKeys, toRemove); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			metrics = make([]*points.Metric, 0, maxBeforeFlush)
			metricKeys = make([]string, 0, maxBeforeFlush)
			toRemove = make([]string, 0, maxBeforeFlush)
			tasks = make([]url.Values, 0, maxBeforeFlush)
		}
	}
	if len(tasks) > 0 {
		if err := addTasks(c, tasks, metrics, metricKeys, toRemove); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func addTasks(c ae.Context,
	tasks []url.Values, metrics []*points.Metric,
	metricKeys, toRemove []string) error {
	if err := c.AddTasks("summary", SummarizeQueueUrl, tasks); err != nil {
		return err
	}
	if err := c.PutMulti(points.MetricDatastoreType, metricKeys, metrics); err != nil {
		return err
	}
	if err := c.DeleteMulti(points.MetricUpdateDatastoreType, toRemove); err != nil {
		return err
	}
	return nil
}

func SummarizeTask(w http.ResponseWriter, r *http.Request) {
	metricKey := r.FormValue("m")
	_, summary, metric, err :=
		points.MetricUpdateKeyDecode(metricKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	c := &impl.Appengine{appengine.NewContext(r)}
	for _, res := range points.AvailableSummarySizes[1:] {
		_, err := points.BuildSummaries(
			c, metric, points.NewSpan(res, summary, summary))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
