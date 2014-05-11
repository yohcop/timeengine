package timeseries

import (
	"fmt"
	"net/http"
	"net/url"

	"timeengine/ae/impl"

	"appengine"
	"appengine/datastore"
)

type update_stats struct {
	Ts  int64   `datastore:"t,noindex"`
	Avg float64 `datastore:"a,noindex"`
	Sum float64 `datastore:"s,noindex"`
	Min float64 `datastore:"i,noindex"`
	Max float64 `datastore:"x,noindex"`
}
type update_summary struct {
	Stats update_stats `datastore:"s,noindex"`
}

const UpdateSchemaMapUrl = "/tasks/updateschemamap"

func UpdateSchemaMap(w http.ResponseWriter, r *http.Request) {
	c := &impl.Appengine{appengine.NewContext(r)}
	n := 50

	key := r.FormValue("m")

	summaries := make([]update_summary, 0)
	keys, err := c.DsGetBetweenKeys("S", key, "", n, &summaries)

	if _, ok := err.(*datastore.ErrFieldMismatch); err != nil && !ok {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Put the structs, this will erase the old fields that are not in
	// update_summary.
	err = c.PutMulti("S", keys, summaries)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(keys) < n {
		// Stop here, we're done.
		return
	}

	v := url.Values{}
	v.Set("m", keys[len(keys)-1])
	err = c.AddTasks("map", UpdateSchemaMapUrl, []url.Values{v})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// All good.
	s := fmt.Sprintf("%#v --- %v", summaries, keys)
	w.Write([]byte(s))
}
