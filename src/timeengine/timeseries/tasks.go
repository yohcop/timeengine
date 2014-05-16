package timeseries

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"timeengine/ae"
	"timeengine/ae/impl"
	"timeengine/timeseries/points"

	"appengine"
)

var _ = log.Println

const SummarizeQueueUrl = "/queue/summarize"

// NOTE: these can't be done in separate tasks (i.e. separate tasks for each
// summary size, for each time series). The reason being that a higher level
// summary needs the data from a lower level summary already computed.
func SummarizeTask(w http.ResponseWriter, r *http.Request) {
	metric := r.FormValue("m")
	ts, err := strconv.ParseInt(r.FormValue("ts"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ssInt, err := strconv.ParseInt(r.FormValue("ss"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ss := points.SelectSummarySize(ssInt)

	c := &impl.Appengine{appengine.NewContext(r)}
	_, err = points.BuildSummaries(c, metric, points.NewSpan(ss, ts, ts))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// So far so good. This summary was computed without errors.
	// Let's schedule the next lower resolution (bigger span) summary.
	nextSs := ss.LowerRes()
	if nextSs == ss {
		// We are done, this was the lower res already.
		return
	}

	v := url.Values{}
	v.Set("m", metric)
	v.Set("ts", fmt.Sprint(ts))
	v.Set("ss", fmt.Sprint(nextSs))

	updateKey := points.MetricUpdateKey(ts, metric, nextSs)
	runAfter := time.Duration(70 * time.Second)
	task := &ae.Task{
		Url:      v,
		Name:     &updateKey,
		RunAfter: &runAfter,
	}
	if err := c.AddTasks("summary", "/queue/summarize", []*ae.Task{task}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
