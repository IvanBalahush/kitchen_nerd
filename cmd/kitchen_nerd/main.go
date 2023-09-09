package main

import (
	"context"
	"kitchen_nerd"
	"kitchen_nerd/database"
	"kitchen_nerd/pkg/logger/zaplog"
	"os"

	"github.com/dmitrymomot/go-env"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
)

// Error is a default error type for fastwallet cli.
var Error = errs.Class("fastwallet cli error")

var (
	databaseConnectionString = env.MustString("DATABASE_URL")

	// Console server
	consoleServerAddress = env.MustString("CONSOLE_SERVER_ADDRESS")
)

// Config contains configurable values for kitchen nerd project.
type Config struct {
	Database            string `json:"database"`
	kitchen_nerd.Config `json:"config"`
}

// commands.
var (
	// fastwallet root cmd.
	rootCmd = &cobra.Command{
		Use:   "fastwallet",
		Short: "cli for interacting with fastwallet project",
	}

	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
	runCfg Config
)

func init() {
	rootCmd.AddCommand(runCmd)
}

func main() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func cmdRun(_ *cobra.Command, _ []string) (err error) {
	ctx := context.Background()
	log := zaplog.NewLog()
	runCfg = createConfig()

	db, err := database.New(ctx, runCfg.Database)
	if err != nil {
		//log.Error("Error starting master database on fastwallet service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer db.Close()

	// TODO: remove for production.
	err = db.CreateSchema(ctx)
	if err != nil {
		log.Error("Error creating schema", Error.Wrap(err))
	}

	kitchenNerd, err := kitchen_nerd.New(log, runCfg.Config, db)
	if err != nil {
		log.Error("Error starting fastwallet service", Error.Wrap(err))
		return Error.Wrap(err)
	}

	runError := kitchenNerd.Run(ctx)
	closeError := kitchenNerd.Close()

	return Error.Wrap(errs.Combine(runError, closeError))
}

// createConfig creates config using env.
func createConfig() (config Config) {
	// Database
	config.Database = databaseConnectionString

	//Console
	config.Console.Server.Address = consoleServerAddress

	return config
}
