package kitchen_nerd

import (
	"context"
	"errors"
	"kitchen_nerd/console/consoleserver"
	"log"
	"net"

	"github.com/zeebo/errs"
	"golang.org/x/sync/errgroup"

	"kitchen_nerd/pkg/logger"
	"kitchen_nerd/users"
)

// DB provides access to all databases and database related functionality.
//
// architecture: Master Database.
type DB interface {
	// Users provides access to users db.
	Users() users.DB

	// Close closes underlying db connection.
	Close()

	// CreateSchema create tables.
	CreateSchema(ctx context.Context) error
}

// Config is the global configuration for kitchen nerd.
type Config struct {
	Console struct {
		Server consoleserver.Config `json:"server"`
	} `json:"console"`
}

type KitchenNerd struct {
	Config   Config
	Log      logger.Logger
	Database DB

	// Users exposes users related logic.
	Users struct {
		Service *users.Service
	}

	// Console web server with web UI.
	Console struct {
		Listener net.Listener
		Endpoint *consoleserver.Server
	}
}

// New is a constructor for KitchenNerd.
func New(logger logger.Logger, config Config, db DB) (kitchenNerd *KitchenNerd, err error) {
	kitchenNerd = &KitchenNerd{
		Log:      logger,
		Database: db,
	}

	{ // users setup.
		kitchenNerd.Users.Service = users.NewService(logger, db.Users())
	}

	{ // console setup.
		kitchenNerd.Console.Listener, err = net.Listen("tcp", config.Console.Server.Address)
		if err != nil {
			return nil, err
		}
		log.Println("123")
		kitchenNerd.Console.Endpoint = consoleserver.NewServer(
			config.Console.Server,
			logger,
			kitchenNerd.Console.Listener,
			kitchenNerd.Users.Service,
		)
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
