package auth

import (
	"encoding/json"
	"net/http"

	"github.com/zeebo/errs"

	"kitchen_nerd/pkg/logger"
	"kitchen_nerd/users"
)

var (
	// ErrAuth is an internal error type for users controller.
	ErrAuth = errs.Class("auth controller error")
)

// Auth is a mvc controller that handles all auth related views.
type Auth struct {
	log   logger.Logger
	users *users.Service
}

// NewAuth is a constructor for users controller.
func NewAuth(log logger.Logger, users *users.Service) *Auth {
	usersController := &Auth{
		log:   log,
		users: users,
	}

	return usersController
}

// TokenResponse contains token.
type TokenResponse struct {
	Token      string `json:"token"`
	Enabled2FA bool   `json:"enabled_2fa"`
}

// Register creates a new user account.
func (c *Auth) Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	ctx := r.Context()

	var request RegistrationRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.serveError(w, http.StatusBadRequest, ErrAuth.Wrap(err))
		return
	}
	defer r.Body.Close()

	if !request.IsValid() {
		c.serveError(w, http.StatusBadRequest, ErrAuth.New("did not fill in all the fields"))
		return
	}

	err := c.users.Create(ctx, request.UserName, request.Email, request.Password)
	if err != nil {
		c.log.Error("Unable to register new user", ErrAuth.Wrap(err))
		c.serveError(w, http.StatusInternalServerError, ErrAuth.Wrap(err))
		return
	}
}

func (c *Auth) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		c.serveError(w, http.StatusBadRequest, ErrAuth.Wrap(err))
		return
	}

	if !request.IsValid() {
		c.serveError(w, http.StatusBadRequest, ErrAuth.New("did not fill in all the fields"))
		return
	}

	err := c.users.Login(ctx, request.Email, request.Password)
	if err != nil {
		c.log.Error("Unable to login a user", ErrAuth.Wrap(err))
		c.serveError(w, http.StatusInternalServerError, ErrAuth.Wrap(err))
		return
	}
}

// serveError replies to request with specific code and error.
func (c *Auth) serveError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)

	var response struct {
		Error string `json:"error"`
	}

	response.Error = err.Error()

	if err = json.NewEncoder(w).Encode(response); err != nil {
		c.log.Error("failed to write json error response", ErrAuth.Wrap(err))
	}
}
