package ae

import (
	"time"

	"net/url"
)

type Task struct {
	Url url.Values

	// Optional fields:
	Name     *string
	RunAfter *time.Duration
	RunAt    *time.Time
}

type Context interface {
	PutMulti(kind string, keys []string, els interface{}) error
	DsGetBetweenKeys(
		kind, from, to string, limit int, els interface{}) (
		keys []string, err error)
	DeleteMulti(kind string, keys []string) error
	AddTasks(queue, parh string, tasks []*Task) error
	Logf(fmt string, args ...interface{})
}
