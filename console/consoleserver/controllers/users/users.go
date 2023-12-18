package users_controller

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/zeebo/errs"
	"html/template"
	"log"
	"net/http"

	"kitchen_nerd/users"
)

var (
	// ErrUsers is an internal error type for users controller
	ErrUsers = errs.Class("users controller error")
)

type Templates struct {
	Profile *template.Template
}

type Users struct {
	users *users.Service

	templates Templates
}

func NewUsers(users *users.Service, templates Templates) *Users {
	usersController := &Users{
		users:     users,
		templates: templates,
	}

	return usersController
}

func (c *Users) Profile(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.Profile.Execute(w, nil); err != nil {
			http.Error(w, "could not execute user's profile tepmlate", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		ctx := r.Context()
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			c.serveError(w, http.StatusBadRequest, err)
			return
		}

		profile, err := c.users.GetProfile(ctx, id)
		if err != nil {
			c.serveError(w, http.StatusBadRequest, err)
			return
		}
		if err = json.NewEncoder(w).Encode(profile); err != nil {
			log.Println("failed to write json error response", ErrUsers.Wrap(err))
			return
		}
	}
}

// serveError replies to request with specific code and error.
func (c *Users) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Println("failed to write json error response", ErrUsers.Wrap(err))
	}
}
