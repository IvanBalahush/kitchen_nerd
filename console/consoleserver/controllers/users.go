package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/zeebo/errs"

	"kitchen_nerd/pkg/logger"
	"kitchen_nerd/users"
)

var (
	// ErrUsers is an internal error type for users controller.
	ErrUsers = errs.Class("users controller error")
)

// Users is a mvc controller that handles all users related views.
type Users struct {
	log   logger.Logger
	users *users.Service
}

// NewUsers is a constructor for users controller.
func NewUsers(log logger.Logger, users *users.Service) *Users {
	usersController := &Users{
		log:   log,
		users: users,
	}

	return usersController
}

// UserFields contains user registration fields.
type UserFields struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

// IsValid check the request for all conditions.
func (uf *UserFields) IsValid() bool {
	switch {
	case uf.FirstName == "":
		return false
	case uf.LastName == "":
		return false
	case uf.Email == "":
		return false
	case uf.Password == "":
		return false
	default:
		return true
	}
}

// TokenResponse contains token.
type TokenResponse struct {
	Token      string `json:"token"`
	Enabled2FA bool   `json:"enabled_2fa"`
}

// Register creates a new user account.
func (u *Users) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request UserFields
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		u.serveError(w, http.StatusBadRequest, ErrUsers.Wrap(err))
		return
	}
	defer r.Body.Close()

	if !request.IsValid() {
		u.serveError(w, http.StatusBadRequest, ErrUsers.New("did not fill in all the fields"))
		return
	}

	err := u.users.Create(ctx, fmt.Sprintf("%s %s", request.FirstName, request.LastName), request.Email, request.Password)
	if err != nil {
		u.log.Error("Unable to register new user", ErrUsers.Wrap(err))
		u.serveError(w, http.StatusInternalServerError, ErrUsers.Wrap(err))
		return
	}
}

// serveError replies to request with specific code and error.
func (u *Users) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	if err = json.NewEncoder(w).Encode(response); err != nil {
		u.log.Error("failed to write json error response", ErrUsers.Wrap(err))
	}
}
