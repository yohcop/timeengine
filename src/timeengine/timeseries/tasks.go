package timeseries

import (
	"net/http"
	"time"
  "fmt"
  "log"

	"timeengine/ae/impl"
	"timeengine/timeseries/points"

	"appengine"
)

var _ = log.Println

func Aggregate60sTask(w http.ResponseWriter, r *http.Request) {
	c := &impl.Appengine{appengine.NewContext(r)}

	res := points.SelectFrameSize(60)

	// Fetch work.
	now := time.Now().Unix()
	to := fmt.Sprintf("%d@9999999999", int64(res.KeyFrame(now)) - 60)

	objs := make([]points.MetricUpdate, 0)
  // Get the top ones. Since they are sorted by when they were
  // inserted, we get the oldest ones.
	keys, err := c.DsGetBetweenKeys("MU", "", to, -1, &objs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

  metrics := make([]*points.Metric, 0, len(keys))
  toRemove := make([]string, 0, len(keys))
  for _, k := range keys {
    if _, frame, metric, err := points.MetricUpdateKeyDecode(k); err == nil {
      metrics = append(metrics, &points.Metric{metric})
      // Only generate 1 frame for now. We can be smarter by agregating
      // multiple keys, but it will rarely be useful as we generate
      // those every minute anyway.
      for _, res := range points.AvailableFrameSizes[1:] {
        _, err := points.BuildFrames(c, metric, points.NewSpan(res, frame, frame))
        if err != nil {
          log.Println(err.Error())
        } else {
          toRemove = append(toRemove, k)
        }
      }
    }
  }
  c.DeleteMulti("MU", toRemove)
}
