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
		points.MetricUpdateDatastoreType, "", to, 5000, &objs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	maxBeforeFlush := 100
	metrics := make([]*points.Metric, 0, len(keys))
	metricKeys := make([]string, 0, maxBeforeFlush)
	toRemove := make([]string, 0, maxBeforeFlush)
	tasks := make([]*ae.Task, 0, maxBeforeFlush)
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
			tasks = append(tasks, &ae.Task{Url:v})
		}
		if len(metrics) >= maxBeforeFlush {
			if err := addTasks(c, tasks, metrics, metricKeys, toRemove); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			metrics = make([]*points.Metric, 0, maxBeforeFlush)
			metricKeys = make([]string, 0, maxBeforeFlush)
			toRemove = make([]string, 0, maxBeforeFlush)
			tasks = make([]*ae.Task, 0, maxBeforeFlush)
		}
	}
	if len(tasks) > 0 {
		if err := addTasks(c, tasks, metrics, metricKeys, toRemove); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func addTasks(c ae.Context,
	tasks []*ae.Task, metrics []*points.Metric,
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

// NOTE: these can't be done in separate tasks (i.e. separate tasks for each
// summary size, for each time series). The reason being that a higher level
// summary needs the data from a lower level summary already computed.
func SummarizeTask(w http.ResponseWriter, r *http.Request) {
	metricKey := r.FormValue("m")
	_, summary, metric, err :=
		points.MetricUpdateKeyDecode(metricKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c := &impl.Appengine{appengine.NewContext(r)}
	for _, res := range points.AvailableSummarySizes[1:] {
		var err error
		if res.USecs() < oneMinute {
			_, err = points.BuildSummaries(
				c, metric, points.NewSpan(res, summary, summary+oneMinute-res.USecs()))
		} else {
			_, err = points.BuildSummaries(
				c, metric, points.NewSpan(res, summary, summary))
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
