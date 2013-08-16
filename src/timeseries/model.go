package timeseries

// Keep names short, saves db space on appengine.
//
// Stores data points.
//
// Offline processes can derive coarser resolutions as well. From 60 records at
// a resolution of 1 second, we can derive a 60 second average (or max, etc)
// point for the same metric name.
//
// The key is M@T, where M is actually "<namespace>*<metric name>"
// This way we can query for all the points in a range by key, and
// never need another index.
type P struct {
	// Value.
	V float64 `datastore:",noindex"`

	// The following unexported fields are not serialized, or stored
	// in memcache, since they can be derived from the key name.

	// Key
	k string
	// Timestamp, Unix time.
	t int64
	// Metric name.
	m string
	// Namespace.
	ns string
}
