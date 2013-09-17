package impl

import (
	"log"
	"net/url"

	"appengine"
	"appengine/datastore"
	"appengine/taskqueue"
)

var _ = log.Println

type Appengine struct {
	C appengine.Context
}

func (ae *Appengine) DsGetBetweenKeys(kind, from, to string, limit int, els interface{}) (keys []string, err error) {
	q := datastore.NewQuery(kind)
	q = q.Order("__key__")
	q = q.Filter("__key__ >=", datastore.NewKey(ae.C, kind, from, 0, nil))
	q = q.Filter("__key__ <=", datastore.NewKey(ae.C, kind, to, 0, nil))
	if limit > 0 {
		q = q.Limit(limit)
	}

	ks, err := q.GetAll(ae.C, els)
	for _, k := range ks {
		keys = append(keys, k.StringID())
	}
	if err != nil {
		log.Println("--------", kind, from, to, limit, "==>", len(ks), err.Error())
	}
	return keys, err
}

func (ae *Appengine) PutMulti(kind string, keys []string, els interface{}) error {
	ks := make([]*datastore.Key, len(keys))
	for i, p := range keys {
		ks[i] = datastore.NewKey(ae.C, kind, p, 0, nil)
	}
	_, err := datastore.PutMulti(ae.C, ks, els)
	if err != nil {
		log.Println("====>", kind, err.Error())
	}
	return err
}

func (ae *Appengine) DeleteMulti(kind string, keys []string) error {
	ks := make([]*datastore.Key, len(keys))
	for i, p := range keys {
		ks[i] = datastore.NewKey(ae.C, kind, p, 0, nil)
	}
	return datastore.DeleteMulti(ae.C, ks)
}

func (ae *Appengine) AddTasks(queue, path string, tasks []url.Values) error {
	aeTasks := make([]*taskqueue.Task, len(tasks))
	for i, values := range tasks {
		aeTasks[i] = taskqueue.NewPOSTTask(path, values)
	}
	_, err := taskqueue.AddMulti(ae.C, aeTasks, queue)
	return err
}
