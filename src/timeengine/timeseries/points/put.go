package points

import (
	"time"

	"timeengine/ae"
	//"strconv"
	"net/url"
)

func minMaxForTask(tasks map[string][]int64, key string, ts int64) []int64 {
	if v, ok := tasks[key]; ok {
		if ts < v[0] {
			v[0] = ts
		} else if ts > v[1] {
			v[1] = ts
		}
		return v
	}
	return []int64{ts, ts}
}

func PutRawPoints(c ae.Context, pts []*P) error {
	keys := make([]string, 0, len(pts))
	for _, p := range pts {
		keys = append(keys, keyAt(p.m, p.t))
	}

	if err := QueueSummaries(c, pts); err != nil {
		c.Logf("Error queuing summaries: %s", err.Error())
		// This will be retried if we return an error here.
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	return c.PutMulti(pointDatastoreType, keys, pts)
}

func QueueSummaries(c ae.Context, pts []*P) error {
	now := time.Now().Unix() * s
	updateKeys := make(map[string]bool)
	tasks := make([]*ae.Task, 0)

	for _, p := range pts {
		updateKey := MetricUpdateKey(now, p.t, p.m)
		if _, present := updateKeys[updateKey]; !present {
			updateKeys[updateKey] = true

			v := url.Values{}
			v.Set("m", updateKey)

			runAfter := time.Duration(60 * time.Second)
			task := &ae.Task{
				Url:      v,
				Name:     &updateKey,
				RunAfter: &runAfter,
			}
			tasks = append(tasks, task)
		}
	}

	return c.AddTasks("summary", "/queue/summarize", tasks)
}
