package users

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"net/http"
)

type User struct {
	Email string

	appengineUser *user.User
	this          *datastore.Key
	Id            string `datastore:"-"`
}

func (u *User) Key() *datastore.Key {
	return u.this
}

func LogoutURL(c appengine.Context) string {
	url, _ := user.LogoutURL(c, "/")
	return url
}

func AuthUser(w http.ResponseWriter, r *http.Request) (*User, error) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		url, err := user.LoginURL(c, r.URL.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return nil, err
		}
		w.Header().Set("Location", url)
		w.WriteHeader(http.StatusFound)
		return nil, nil
	}
	return FindOrNewUser(u, r)
}

func FindOrNewUser(appengineUser *user.User, r *http.Request) (*User, error) {
	c := appengine.NewContext(r)

	user := &User{appengineUser: appengineUser}
	k := datastore.NewKey(c, "User", appengineUser.Email, 0, nil)

	err := datastore.Get(c, k, user)
	if err == nil {
		user.this = k
		user.Id = k.Encode()
		return user, nil
	}
	user.Email = appengineUser.Email
	k, err = datastore.Put(c, k, user)
	user.this = k
	user.Id = k.Encode()
	return user, err
}

func FindUser(c appengine.Context, userId string) (*User, error) {
	userKey, err := datastore.DecodeKey(userId)
	if err != nil {
		return nil, err
	}
	user := &User{}
	if err := datastore.Get(c, userKey, user); err != nil {
		return nil, err
	}
	user.this = userKey
	user.Id = userKey.Encode()
	return user, nil
}
