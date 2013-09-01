package dashboard

import (
	"errors"
	"log"
	"regexp"
	"strings"

	"appengine"
	"appengine/datastore"
)

var _ = log.Println

func ValidDashboard(name string) (string, error) {
	name = strings.ToLower(name)
	if regexp.MustCompile("[a-z0-9-_.]+").Match([]byte(name)) {
		return name, nil
	}
	return "", errors.New(
		"Invalid dashboard. Should match /[a-z0-9-_.]+/")
}

func DashboardKey(c appengine.Context, name string) *datastore.Key {
	return datastore.NewKey(c, "Dash", name, 0, nil)
}

func GetDashboard(c appengine.Context, name string) *Dashboard {
	// No memcache for dashboards as long as we don't update the
	// cache on edits.
	/*
		if item := getDashFromMemcache(c, name); item != nil {
			return item
		}
	*/
	return getDashFromDatastore(c, name)
}

/*
func getDashFromMemcache(c appengine.Context, name string) *Dashboard {
	item := &Dashboard{}
	if _, err := memcache.Gob.Get(c, name, item); err == nil {
		return item
	}
	return nil
}
*/

func getDashFromDatastore(c appengine.Context, name string) *Dashboard {
	ts := &Dashboard{}
	if err := datastore.Get(c, DashboardKey(c, name), ts); err != nil {
		return nil
	}
	return ts
}
