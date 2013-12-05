package namespace

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"appengine"
	"appengine/datastore"
	"appengine/memcache"
)

var _ = log.Println

func ValidNamespace(ns string) (string, error) {
	ns = strings.ToLower(ns)
	if regexp.MustCompile("[a-z0-9-_.]+").Match([]byte(ns)) {
		return ns, nil
	}
	return "", errors.New(
		"Invalid namespace. Should match /[a-z0-9-_]+/")
}

func MetricName(ns, metric string) string {
	return fmt.Sprintf("%s*%s", ns, metric)
}

func NsKey(c appengine.Context, ns string) *datastore.Key {
	return datastore.NewKey(c, "Ns", ns, 0, nil)
}

func VerifyNamespace(c appengine.Context, ns, secret string) bool {
	if ns == "test" && secret == "test" {
		return true
	}
	if namespace := getNs(c, ns); namespace != nil {
		return secret == namespace.S
	}
	return false
}

func getNs(c appengine.Context, ns string) *Ns {
	if item := getNsFromMemcache(c, ns); item != nil {
		return item
	}
	item := getNsFromDatastore(c, ns)
	if item != nil {
		putNsInMemcache(c, ns, item)
	}
	return item
}

func getNsFromMemcache(c appengine.Context, ns string) *Ns {
	item := &Ns{}
	if _, err := memcache.Gob.Get(c, ns, item); err == nil {
		return item
	}
	return nil
}

func getNsFromDatastore(c appengine.Context, ns string) *Ns {
	ts := &Ns{}
	if err := datastore.Get(c, NsKey(c, ns), ts); err != nil {
		return nil
	}
	return ts
}

func putNsInMemcache(c appengine.Context, ns string, obj *Ns) {
	// Ignore any return value at this point. If setting the namespace
	// object in memcache fails, whatever.
	memcache.Gob.Set(c, &memcache.Item{
		Key:        ns,
		Object:     obj,
		Expiration: time.Hour,
	})
}
