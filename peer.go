package kitchen_nerd

import (
	"context"
	"errors"
	"kitchen_nerd/recipes"
	"log"
	"net"

	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"kitchen_nerd/console/consoleserver"
	"kitchen_nerd/tokens"
	"kitchen_nerd/users"
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database.
type DB interface {
	// Users provides access to users db.
	Users() users.DB

	// Recipes provides access to recipes db.
	Recipes() recipes.DB

	// Tokens provides access to tokens db.
	Tokens() tokens.DB

	// Close closes underlying db connection.
	Close()

	// CreateSchema create tables.
	CreateSchema(ctx context.Context) error
}

// Config is the global configuration for kitchen nerd.
type Config struct {
	DatabaseURL    string `env:"DATABASE_URL,notEmpty"`
	ServerAddress  string `env:"SERVER_ADDRESS,notEmpty"`
	StaticDir      string `env:"STATIC_DIR,notEmpty"`
	ExportDataPath string `env:"EXPORT_DATA_PATH,notEmpty"`
}

type KitchenNerd struct {
	Config   Config
	Database DB

	// Users exposes users related logic.
	Users struct {
		Service *users.Service
	}

	Recipes struct {
		Service *recipes.Service
	}

	// Console web server with web UI.
	Console struct {
		Listener net.Listener
		Endpoint *consoleserver.Server
	}
}

// New is a constructor for KitchenNerd.
func New(config Config, db DB) (kitchenNerd *KitchenNerd, err error) {
	kitchenNerd = &KitchenNerd{
		Database: db,
	}

	{ // users setup.
		kitchenNerd.Users.Service = users.NewService(db.Users(), db.Tokens())
	}

	{ // recipes setup.
		kitchenNerd.Recipes.Service = recipes.NewService(db.Recipes())
	}

	{ // console setup.
		kitchenNerd.Console.Listener, err = net.Listen("tcp", config.ServerAddress)
		if err != nil {
			return nil, err
		}
		log.Println("123")
		cfg := consoleserver.Config{
			ServerAddress: config.ServerAddress,
			StaticDir:     config.StaticDir,
		}

		kitchenNerd.Console.Endpoint, err = consoleserver.NewServer(
			cfg,
			kitchenNerd.Console.Listener,
			kitchenNerd.Users.Service,
			kitchenNerd.Recipes.Service,
		)
		if err != nil {
			log.Fatal(err)
		}
	}

	return kitchenNerd, nil
}

// Run runs kitchenNerd until it's either closed or it errors.
func (kitchenNerd *KitchenNerd) Run(ctx context.Context) error {
	group, ctx := errgroup.WithContext(ctx)

	// start kitchenNerd servers as a separate goroutine.
	group.Go(func() error {
		return ignoreCancel(kitchenNerd.Console.Endpoint.Run(ctx))
	})

	return group.Wait()
}

// Close closes all the resources.
func (kitchenNerd *KitchenNerd) Close() error {
	var errlist errs.Group

	errlist.Add(kitchenNerd.Console.Endpoint.Close())

	return errlist.Err()
}

// we ignore cancellation and stopping errors since they are expected.
func ignoreCancel(err error) error {
	if errors.Is(err, context.Canceled) {
		return nil
	}

	return err
}
