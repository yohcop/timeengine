package ae

import (
	"net/url"
)

type Context interface {
	PutMulti(kind string, keys []string, els interface{}) error
	DsGetBetweenKeys(
		kind, from, to string, limit int, els interface{}) (
		keys []string, err error)
  DeleteMulti(kind string, keys[]string) error
	AddTasks(queue, parh string, tasks []url.Values) error
}
