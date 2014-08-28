package namespace

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"

	"timeengine/users"

	"appengine"
	"appengine/datastore"
)

func NewNamespace(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := users.IsAuthorized(r); !ok {
		return
	}

	ns, err := ValidNamespace(r.FormValue("ns"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	c := appengine.NewContext(r)
	// Check if the namespace already exists.
	if getNs(c, ns) != nil {
		http.Error(w, "Namespace exists", http.StatusBadRequest)
		return
	}

	key := NsKey(c, ns)
	var secret string
	if ns == "test" {
		secret = "test"
	} else {
		secret, err = randString(10)
		if err != nil {
			http.Error(w, err.Error(),
				http.StatusInternalServerError)
			return
		}
	}
	namespace := &Ns{
		S: secret,
		D: false,
	}
	if _, err := datastore.Put(c, key, namespace); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type Namespace struct {
	Name   string
	Secret string
	First  int64
	Last   int64
	Delete bool
}

type NsListResp struct {
	Namespaces []*Namespace
}

func ListNamespaces(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := users.IsAuthorized(r); !ok {
		return
	}

	c := appengine.NewContext(r)
	q := datastore.NewQuery("Ns").Order("__key__")
	nss := make([]*Ns, 0)
	keys, err := q.GetAll(c, &nss)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := &NsListResp{}
	for i, ns := range nss {
		resp.Namespaces = append(resp.Namespaces, &Namespace{
			Name:   keys[i].StringID(),
			Secret: ns.S,
			First:  ns.F,
			Last:   ns.L,
			Delete: ns.D,
		})
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}

func randString(n int) (string, error) {
	var bytes = make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bytes), nil
}
