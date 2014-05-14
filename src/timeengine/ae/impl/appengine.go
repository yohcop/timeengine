package impl

import (
	"encoding/hex"
	"log"

	"timeengine/ae"

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
	if len(from) > 0 {
		q = q.Filter("__key__ >=", datastore.NewKey(ae.C, kind, from, 0, nil))
	}
	if len(to) > 0 {
		q = q.Filter("__key__ <=", datastore.NewKey(ae.C, kind, to, 0, nil))
	}
	if limit > 0 {
		q = q.Limit(limit)
	}

	ks, err := q.GetAll(ae.C, els)
	for _, k := range ks {
		keys = append(keys, k.StringID())
	}
	if err != nil {
		log.Println("--------", kind, from, to, limit, "==>", len(ks), err.Error())
		ae.C.Errorf("Error fetching data: %s [%s, %s] %d (len=%d): %s", kind, from, to, limit, len(ks), err.Error())
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

// Enqueues all the tasks.
func (ae *Appengine) pushTasks(queue, path string, tasks []*ae.Task) error {
	aeTasks := make([]*taskqueue.Task, len(tasks))
	for i, task := range tasks {
		aeTask := taskqueue.NewPOSTTask(path, task.Url)

		if task.Name != nil {
			// Make sure the name is valid..."
			aeTask.Name = hex.EncodeToString([]byte(*task.Name))
		}

		if task.RunAfter != nil {
			aeTask.Delay = *task.RunAfter
		} else if task.RunAt != nil {
			aeTask.ETA = *task.RunAt
		}

		aeTasks[i] = aeTask
	}
	_, err := taskqueue.AddMulti(ae.C, aeTasks, queue)
	// Ignore if the tasks were already added.
	if multi, ok := err.(appengine.MultiError); err != nil && ok {
	  // If one of the errors isn't ErrTaskAlreadyAdded, return all the errors.
		for _, erri := range multi {
			if erri != taskqueue.ErrTaskAlreadyAdded {
				return err
			}
		}
	} else if err != nil && !ok && err != taskqueue.ErrTaskAlreadyAdded {
	  // If this is not a MultiError, and is not ErrTaskAlreadyAdded, return that error.
		return err
	}
	return nil
}

func (ae *Appengine) AddTasks(queue, path string, tasks []*ae.Task) error {
	// Push at max. 100 tasks at a time.
	for len(tasks) > 0 {
		max := 100
		if max > len(tasks) {
			max = len(tasks)
		}
		err := ae.pushTasks(queue, path, tasks[:max])
		if err != nil {
			return err
		}
		tasks = tasks[max:]
	}
	return nil
}

func (ae *Appengine) Logf(fmt string, args ...interface{}) {
	ae.C.Errorf(fmt, args...)
}
