package consoleserver

import (
	"context"
	"errors"
	"kitchen_nerd/console/consoleserver/controllers/auth"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"kitchen_nerd/console/consoleserver/controllers"
	"kitchen_nerd/pkg/logger"
	"kitchen_nerd/users"
)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("console web server error")
)

// Config contains configuration for console web server.
type Config struct {
	Address string `json:"address"`
}

// Server represents console web server.
//
// architecture: Endpoint
type Server struct {
	log    logger.Logger
	config Config

	listener net.Listener
	server   http.Server

	users *users.Service
}

// NewServer is a constructor for console web server.
func NewServer(config Config, log logger.Logger, listener net.Listener, users *users.Service) *Server {
	server := &Server{
		log:      log,
		config:   config,
		listener: listener,
		users:    users,
	}

	authController := auth.NewAuth(server.log, server.users)
	usersController := controllers.NewUsers(server.log, server.users)

	router := mux.NewRouter()
	router.Use(cors.AllowAll().Handler)

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", authController.Login).Methods(http.MethodPost)
	authRouter.HandleFunc("/register", authController.Register).Methods(http.MethodPost)
	usersRouter := router.PathPrefix("/users").Subrouter()
	usersRouter.HandleFunc("/register", usersController.Register).Methods(http.MethodPost)

	server.server = http.Server{
		Handler: router,
	}

	return server
}

// Run starts the server that host webapp and api endpoint.
func (server *Server) Run(ctx context.Context) (err error) {
	ctx, cancel := context.WithCancel(ctx)
	var group errgroup.Group
	group.Go(func() error {
		<-ctx.Done()
		return Error.Wrap(server.server.Shutdown(context.Background()))
	})
	group.Go(func() error {
		defer cancel()
		err := server.server.Serve(server.listener)
		isCancelled := errs.IsFunc(err, func(err error) bool { return errors.Is(err, context.Canceled) })
		if isCancelled || errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		return Error.Wrap(err)
	})

	return Error.Wrap(group.Wait())
}

// Close closes server and underlying listener.
func (server *Server) Close() error {
	return Error.Wrap(server.server.Close())
}

// appHandler is web app http handler function.
func (server *Server) appHandler(w http.ResponseWriter, r *http.Request) {
	header := w.Header()

	header.Set("Content-Type", "text/html; charset=UTF-8")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "same-origin")
}
