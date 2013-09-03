package timeengine

import (
	"net/http"

	"timeengine/compat"
	"timeengine/dashboard"
	"timeengine/namespace"
	"timeengine/timeseries"
	"timeengine/ui"
)

func init() {
	// Us pages.
	http.HandleFunc("/", ui.Index)
	http.HandleFunc("/dashboards", ui.Dashboards)
	http.HandleFunc("/dashboard/edit", ui.DashboardEditor)
	http.HandleFunc("/namespaces", ui.Namespaces)

	// Api stuff. Doesn't render a UI, but handles ajax calls.
	http.HandleFunc("/api/timeseries/put", timeseries.PutDataPoints)
	http.HandleFunc("/api/timeseries/get", timeseries.GetDataPoints)

	http.HandleFunc("/api/namespace/new/", namespace.NewNamespace)
	http.HandleFunc("/api/namespace/list/", namespace.ListNamespaces)

	http.HandleFunc("/api/dashboard/new/", dashboard.NewDashboard)
	http.HandleFunc("/api/dashboard/list/", dashboard.ListDashboards)
	http.HandleFunc("/api/dashboard/save/", dashboard.SaveDashboard)
	http.HandleFunc("/api/dashboard/get", dashboard.GetDashboard)

	// Backward compatible with graphite:
	// get a dashboard, and a tiny subset of the json renderer.
	http.HandleFunc("/render/", compat.Render)

	// Task queues handlers
	http.HandleFunc("/tasks/aggregateminute", timeseries.Aggregate60sTask)
}
