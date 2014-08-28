package users

import (
	"encoding/json"
	"net/http"

	"appengine"
	"appengine/datastore"
)

func NewUser(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := IsAuthorized(r); !ok {
		return
	}

	email := r.FormValue("email")
	c := appengine.NewContext(r)
	// Check if the namespace already exists.
	if u, err := FindUser(c, email); u != nil || err == nil {
		http.Error(w, "User exists", http.StatusBadRequest)
		return
	}

	key := datastore.NewKey(c, "User", email, 0, nil)
	u := &User{
		Email: email,
	}
	if _, err := datastore.Put(c, key, u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type UserData struct {
	Email string
}

type UserListResp struct {
	Users []*UserData
}

func ListUsers(w http.ResponseWriter, r *http.Request) {
	if ok, _, _ := IsAuthorized(r); !ok {
		return
	}

	c := appengine.NewContext(r)
	q := datastore.NewQuery("User").Order("__key__")
	uss := make([]*User, 0)
	keys, err := q.GetAll(c, &uss)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	resp := &UserListResp{}
	for i := range uss {
		resp.Users = append(resp.Users, &UserData{
			Email:   keys[i].StringID(),
		})
	}
	s, _ := json.Marshal(resp)
	w.Write(s)
}
