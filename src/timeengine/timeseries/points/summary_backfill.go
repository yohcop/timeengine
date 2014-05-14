package points

import (
	"fmt"
	"net/http"
	"net/url"

	"timeengine/ae"
	"timeengine/ae/impl"

	"appengine"
	"appengine/datastore"
)

func BackfilSummaries(w http.ResponseWriter, r *http.Request, path, function string) {
	c := &impl.Appengine{appengine.NewContext(r)}
	n := 5000
	key := r.FormValue("m")
	pts := make([]*P, 0)

	// Fetch points.
	keys, err := c.DsGetBetweenKeys("P", key, "", n, &pts)
	// Accept the old datapoints. (datastructure doesn't match anymore)
	if _, ok := err.(*datastore.ErrFieldMismatch); err != nil && !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = QueueSummaries(c, pts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	v := url.Values{}
	v.Set("m", keys[len(keys)-1])
	v.Set("f", function)
	task := &ae.Task{Url: v}

	continue_tasks := r.FormValue("c")
	if continue_tasks != "no" && len(keys) > 1 {
		err = c.AddTasks("map", path, []*ae.Task{task})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// All good.
	s := fmt.Sprintf("Next: %#v", v)
	w.Write([]byte(s))
}
