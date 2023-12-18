package auth

import (
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"kitchen_nerd/tokens"
	"log"
	"net/http"
	"strings"

	"github.com/zeebo/errs"

	"kitchen_nerd/users"
)

var (
	// ErrAuth is an internal error type for auth controller.
	ErrAuth = errs.Class("auth controller error")
)

const (
	bearerPrefix = "Bearer "

	KeyUserID   = "user_id"
	KeyToken    = "token"
	KeyUsername = "username"
)

// Templates holds all users auth related templates.
type Templates struct {
	Register *template.Template
	Login    *template.Template
}

// Auth is a mvc controller that handles all auth related views.
type Auth struct {
	users  *users.Service
	tokens *tokens.Service

	templates Templates
}

// NewAuth is a constructor for users controller.
func NewAuth(users *users.Service, templates Templates) *Auth {
	usersController := &Auth{
		users:     users,
		templates: templates,
	}

	return usersController
}

// Register creates a new user account.
func (c *Auth) Register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.Register.Execute(w, nil); err != nil {
			http.Error(w, "could not execute register user template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
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
			log.Println("Unable to register new user", ErrAuth.Wrap(err))
			c.serveError(w, http.StatusConflict, ErrAuth.Wrap(err))
			return
		}
	}
}

func (c *Auth) Login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		if err := c.templates.Login.Execute(w, nil); err != nil {
			http.Error(w, "could not execute login user template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
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

		authToken, err := c.users.Login(ctx, request.Email, request.Password)
		if err != nil {
			if errors.As(users.ErrWrongCredentials, &err) {
				c.serveError(w, http.StatusBadRequest, ErrAuth.Wrap(err))
				return
			}
			log.Println("Unable to login a user", ErrAuth.Wrap(err))
			c.serveError(w, http.StatusInternalServerError, ErrAuth.Wrap(err))
			return
		}

		if err = json.NewEncoder(w).Encode(authToken); err != nil {
			log.Println("failed to write json error response", ErrAuth.Wrap(err))
		}
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
		log.Println("failed to write json error response", ErrAuth.Wrap(err))
	}
}

// AuthMiddleware performs token check
func (c *Auth) AuthMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		ctx := r.Context()

		token := r.Header.Get("Authorization")
		if token == "" {
			handler.ServeHTTP(w, r.Clone(ctx))
			return

			//c.serveError(w, http.StatusUnauthorized, ErrAuth.New("Authorization header is not set"))
			//return
		}

		token = strings.TrimPrefix(token, bearerPrefix)

		userToken, err := c.tokens.GetToken(ctx, token)
		if err != nil {
			switch {
			case tokens.ErrNoToken.Has(err):
				c.serveError(w, http.StatusUnauthorized, ErrAuth.Wrap(err))
			case ErrAuth.Has(err):
				c.serveError(w, http.StatusUnauthorized, ErrAuth.Wrap(err))
			default:
				c.serveError(w, http.StatusInternalServerError, ErrAuth.Wrap(err))
			}

			return
		}

		user, err := c.users.Get(ctx, userToken.UserID)
		if err != nil {
			c.serveError(w, http.StatusInternalServerError, ErrAuth.Wrap(err))
			return
		}
		// Добавим информацию о пользователе в контекст
		ctx = context.WithValue(ctx, KeyUserID, userToken.UserID)
		ctx = context.WithValue(ctx, KeyToken, userToken.Token)
		ctx = context.WithValue(ctx, KeyUsername, user.Name)

		// Передаем контекст в обработчик
		handler.ServeHTTP(w, r.Clone(ctx))
	})
}
