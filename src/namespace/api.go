package namespace

import (
	"encoding/json"
	"math/rand"
	"net/http"

	"users"

	"appengine"
	"appengine/datastore"
)

func NewNamespace(w http.ResponseWriter, r *http.Request) {
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
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
	namespace := &Ns{
		S: randString(10),
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
	user, err := users.AuthUser(w, r)
	if user == nil || err != nil {
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

func randString(n int) string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	for i := range bytes {
		bytes[i] = alphabet[rand.Intn(len(alphabet))]
	}
	return string(bytes)
}
