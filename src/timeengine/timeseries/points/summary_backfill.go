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
	from := r.FormValue("from")
	to := r.FormValue("to")
	continue_tasks := r.FormValue("continue")

	// Fetch points.
	pts := make([]*P, 0)
	keys, err := c.DsGetBetweenKeys("P", from, to, n, &pts)
	// Accept the old datapoints. (datastructure doesn't match anymore)
	if _, ok := err.(*datastore.ErrFieldMismatch); err != nil && !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, p := range pts {
		decodePointStrKey(keys[i], p)
	}

	if err = QueueSummaries(c, pts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(keys) == 0 {
		w.Write([]byte("No keys returned"))
		return
	}

	v := url.Values{}
	v.Set("from", keys[len(keys)-1])
	v.Set("to", to)
	v.Set("f", function)
	v.Set("continue", continue_tasks)
	task := &ae.Task{Url: v}

	if continue_tasks == "yes" && len(keys) > 1 {
		err = c.AddTasks("map", path, []*ae.Task{task})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		for _, p := range pts {
		  w.Write([]byte(p.Key() + "\n"))
		}
	}

	// All good.
	s := fmt.Sprintf("Next: %#v", v)
	w.Write([]byte(s))
}
