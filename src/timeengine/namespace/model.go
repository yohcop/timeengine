package namespace

type Ns struct {
	// Namespace.
	// Must stay immutable.
	// This is unexported, since it's used as the key, we don't need
	// to store it.
	name string

	// Secret key, needed to send values to this time series.
	// It's not really a secret, it's mostly just to avoid polluting
	// another namespace unintentionally with a typo (e.g. pushing data
	// to my.namespace.268 instead of the intended my.namespace.288).
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
