package timeengine

import (
	"net/http"

	"timeengine/timeseries/points"
)

const MapperUrl = "/tasks/map"

func Mapper(w http.ResponseWriter, r *http.Request) {
	function := r.FormValue("f")
	switch function {
	case "backfillSummaries": points.BackfilSummaries(w, r, MapperUrl, function)
	}
}
