package timeengine

import (
	"net/http"

	"timeengine/compat"
	"timeengine/dashboard"
	"timeengine/namespace"
	"timeengine/timeseries"
	"timeengine/ui"
	"timeengine/users"
)

func init() {
	// UI pages.
	http.HandleFunc("/", ui.Index)
	http.HandleFunc("/dashboards", ui.Dashboards)
	http.HandleFunc("/dashboard/edit", ui.DashboardEditor)
	http.HandleFunc("/namespaces", ui.Namespaces)
	http.HandleFunc("/push", ui.PushPage)
	http.HandleFunc("/debug", ui.DebugPage)

	// Test page. Verifies that the user is logged in, and can send
	// data. Mostly for use in shell scripts.
	http.HandleFunc("/checkauth", checkUser)

	// Api stuff. Doesn't render a UI, but handles ajax calls.
	http.HandleFunc("/api/timeseries/put", timeseries.PutDataPoints)
	http.HandleFunc("/api/timeseries/get", timeseries.GetDataPoints)

	http.HandleFunc("/api/namespace/new", namespace.NewNamespace)
	http.HandleFunc("/api/namespace/list", namespace.ListNamespaces)

	http.HandleFunc("/api/dashboard/new", dashboard.NewDashboard)
	http.HandleFunc("/api/dashboard/list", dashboard.ListDashboards)
	http.HandleFunc("/api/dashboard/save", dashboard.SaveDashboard)
	http.HandleFunc("/api/dashboard/get", dashboard.GetDashboard)
	http.HandleFunc("/api/dashboard/delete", dashboard.DeleteDashboard)

	// Backward compatible with graphite:
	// get a dashboard, and a tiny subset of the json renderer.
	http.HandleFunc("/render", compat.Render)  // HU?
	http.HandleFunc("/render/", compat.Render)

	// Task and queues handlers
	http.HandleFunc("/tasks/summarize60", timeseries.SummarizeCron)
	http.HandleFunc(timeseries.SummarizeQueueUrl, timeseries.SummarizeTask)

	http.HandleFunc(timeseries.UpdateSchemaMapUrl, timeseries.UpdateSchemaMap)
}

func checkUser(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return
	}
	w.Write([]byte("ok"))
}
