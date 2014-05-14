package timeseries

import (
	"log"
	"net/http"

	"timeengine/ae/impl"
	"timeengine/timeseries/points"

	"appengine"
)

var _ = log.Println

const SummarizeQueueUrl = "/queue/summarize"
const oneMinute = int64(60 * 1000000)

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
