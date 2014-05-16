package points

import (
	"fmt"
	"time"

	"net/url"
	"timeengine/ae"
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

	// This will be retried if we return an error.
	if err := c.PutMulti(pointDatastoreType, keys, pts); err != nil {
		c.Logf("Error saving datapoints: %s", err.Error())
		return err
	}
	if err := QueueSummaries(c, pts); err != nil {
		c.Logf("Error queuing summaries: %s", err.Error())
		return err
	}
	return nil
}

func QueueSummaries(c ae.Context, pts []*P) error {
	updateKeys := make(map[string]bool)
	tasks := make([]*ae.Task, 0)
	ss := AvailableSummarySizes[1]

	for _, p := range pts {
		updateKey := MetricUpdateKey(p.t, p.m, ss)
		if _, present := updateKeys[updateKey]; !present {
			updateKeys[updateKey] = true

			v := url.Values{}
			v.Set("m", p.m)
			v.Set("ts", fmt.Sprint(p.t))
			v.Set("ss", fmt.Sprint(ss))

			// We decide to wait at least 10 seconds before starting this task.
			// It's best if all the points that will go in that summary are saved when
			// this happens. If they are not, and some come later, a new task will be
			// posted anyway. We can run in a race condition though, between the task
			// existence check (a task for this datapoint exists), if that task already
			// started running. It's unclear if the task is removed from the queue at the
			// execution start or end on appengine docs.
			runAfter := time.Duration(10 * time.Second)
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
