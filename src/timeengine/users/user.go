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

func IsAuthorized(r *http.Request) (bool, *User, error) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	return IsUserAuthorized(r, u)
}

func IsUserAuthorized(r *http.Request, u *user.User) (bool, *User, error) {
	if u == nil {
		return false, nil, nil
	}
	// Admins are always authorized.
	if u.Admin {
		user, err := FindOrNewUser(u, r)
		return true, user, err
	}
	c := appengine.NewContext(r)
	user, err := FindUser(c, u.Email)
	if err != nil || user == nil {
		return false, nil, err
	}
	return true, user, nil
}

func RedirectToLogin(w http.ResponseWriter, r *http.Request) error {
	c := appengine.NewContext(r)
	url, err := user.LoginURL(c, r.URL.String())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Location", url)
	w.WriteHeader(http.StatusFound)
	return nil
}

// TODO: change r to appengine.Context. A lot of appengine.Context are created
// everywhere in here, it's not necessary.
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
