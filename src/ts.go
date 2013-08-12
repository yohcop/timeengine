package timeengine

// Keep names short, saves db space on appengine.
//
// Stores data points.
// For metrics expected at every second:
// 1 1234567890 foo.bar.baz.cpu_usage 0.45
// For metrics expected at every minute:
// 60 1234567860 foo.bar.baz.mem_usage 1345
// The time is always modulo resolution:
// if timestamp is ts, and resolution is R, then:
// T = ts - (ts % R)
//
// Offline processes can derive coarser resolutions as well. From 60 records at
// a resolution of 1 second, we can derive a 60 second average (or max, etc)
// point for the same metric name. In that case R is set to false.
// This is important in case the metrics values change, and we need to recompute
// coarser resolutions.
//
// When a point is saved, if there exist an AggregateDefinition where
// Metric == P.M and FromRes == P.R and LastPoint > P.T, then a portion
// of this metric must be recomputed, and an AggregateRequest is created.
// Maybe also delete the relevant p entry?
//
// The key is M@R@T, where M is actually "<namespace>#<metric name>"
// This way we can query for all the points in a range by key, and
// never need another index.
type P struct {
	// Value.
	V float64 `datastore:",noindex"`

  // The following unexported fields are not serialized, or stored
  // in memcache, since they can be derived from the key name.

  // Key
  k string
	// Resolution, in seconds.
	r int
	// Timestamp, Unix time.
	t int64
	// Metric name.
	m string
  // Namespace.
  ns string
}

type Ns struct {
  // Namespace.
  // Must stay immutable.
  // This is unexported, since it's used as the key, we don't need
  // to store it.
  name string

  // Secret key, needed to send values to this time series.
  // It's not really a secret, it's mostly just to avoid poluting
  // another namespace unintentionally.
  // Must stay immutable.
  S string

  // First point (timestamp) in this time series
  F int64
  // Last point (timestamp) in this time series
  L int64

  // If set to true, a cron job will delete all the points in this
  // series.
  D bool
}
