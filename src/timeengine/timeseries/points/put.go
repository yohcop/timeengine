package points

import (
	"time"

	"timeengine/ae"
	//"strconv"
	//"net/url"
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
	now := time.Now().Unix()
	updateKeys := make([]string, 0, len(pts))
	//taskValues := make(map[string][]int64)

	for _, p := range pts {
		keys = append(keys, keyAt(p.m, p.t))
		updateKeys = append(updateKeys, MetricUpdateKey(now, p.t, p.m))
		//taskValues[p.m] = minMaxForTask(taskValues, p.m, p.t)
	}

	//tasks := make([]url.Values, 0)
	//for k, limits := range taskValues {
	//  v := url.Values{}
	//  v.Set("m", k)
	//  v.Set("f", strconv.FormatInt(limits[0], 10))
	//  v.Set("t", strconv.FormatInt(limits[1], 10))
	//  tasks = append(tasks, v)
	//}
	//c.AddTasks("aggregate", "/queue/aggregate", tasks)
	c.PutMulti("MU", updateKeys, make([]MetricUpdate, len(updateKeys)))
	return c.PutMulti("P", keys, pts)
}
