package consoleserver

import (
	"context"
	"errors"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"
	"html/template"
	"kitchen_nerd/console/consoleserver/controllers/auth"
	recipes_controller "kitchen_nerd/console/consoleserver/controllers/recipes"
	users_controller "kitchen_nerd/console/consoleserver/controllers/users"
	"kitchen_nerd/recipes"
	"net"
	"net/http"
	"path/filepath"

	"kitchen_nerd/users"
)

var (
	// Error is an error class that indicates internal http server error.
	Error = errs.Class("console web server error")
)

// Config contains configuration for console web server.
type Config struct {
	DatabaseURL    string `env:"DATABASE_URL,notEmpty"`
	ServerAddress  string `env:"SERVER_ADDRESS,notEmpty"`
	StaticDir      string `env:"STATIC_DIR,notEmpty"`
	ExportDataPath string `env:"EXPORT_DATA_PATH,notEmpty"`
}

// Server represents console web server.
//
// architecture: Endpoint
type Server struct {
	config Config

	listener net.Listener
	server   http.Server

	templates struct {
		auth    auth.Templates
		recipes recipes_controller.Templates
		users   users_controller.Templates
	}
}

// NewServer is a constructor for console web server.
func NewServer(config Config, listener net.Listener, users *users.Service, recipes *recipes.Service) (*Server, error) {
	server := &Server{
		config:   config,
		listener: listener,
	}

	err := server.initializeTemplates()
	if err != nil {
		return nil, err
	}

	authController := auth.NewAuth(users, server.templates.auth)
	recipesController := recipes_controller.NewRecipes(recipes, server.templates.recipes)
	usersController := users_controller.NewUsers(users, server.templates.users)
	router := mux.NewRouter()
	router.Use(cors.AllowAll().Handler)

	authRouter := router.PathPrefix("/auth").Subrouter()
	authRouter.HandleFunc("/login", authController.Login).Methods(http.MethodGet, http.MethodPost)
	authRouter.HandleFunc("/register", authController.Register).Methods(http.MethodGet, http.MethodPost)

	recipesRouter := router.PathPrefix("/recipes").Subrouter()
	recipesRouter.Use(authController.AuthMiddleware)
	recipesRouter.HandleFunc("/list", recipesController.List).Methods(http.MethodGet, http.MethodPost)
	recipesRouter.HandleFunc("/id/{id}", recipesController.Get).Methods(http.MethodGet, http.MethodPost)

	usersRouter := router.Path("/users").Subrouter()
	usersRouter.HandleFunc("/{id}", usersController.Profile).Methods(http.MethodGet, http.MethodPost)
	recipesRouter.Use()
	//adminRecipesRouter := router.PathPrefix("/admin-recipes")
	recipesRouter.HandleFunc("/add", recipesController.Create).Methods(http.MethodGet, http.MethodPost)
	//recipesRouter.HandleFunc("/id/{id}/update", recipesController.Update).Methods(http.MethodGet, http.MethodPost)
	recipesRouter.HandleFunc("/id/{id}/delete", recipesController.Delete).Methods(http.MethodPost)
	web := http.FileServer(http.Dir(server.config.StaticDir))
	router.PathPrefix("/web/").Handler(http.StripPrefix("/web/", web))

	server.server = http.Server{
		Handler: router,
	}

	return server, nil
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

// initializeTemplates initializes and caches templates for managers controller.
func (server *Server) initializeTemplates() (err error) {
	server.templates.auth.Register, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "auth", "register.html"))
	if err != nil {
		return err
	}
	server.templates.auth.Login, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "auth", "login.html"))
	if err != nil {
		return err
	}

	server.templates.recipes.Create, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "admins", "create.html"))
	if err != nil {
		return err
	}
	server.templates.recipes.List, err = template.ParseFiles(filepath.Join(server.config.StaticDir, "index.html"))
	if err != nil {
		return err
	}

	return nil
}

// appHandler is web app http handler function.
func (server *Server) appHandler(w http.ResponseWriter, r *http.Request) {
	header := w.Header()

	header.Set("Content-Type", "text/html; charset=UTF-8")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Referrer-Policy", "same-origin")
}
