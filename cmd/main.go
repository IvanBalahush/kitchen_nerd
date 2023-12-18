package main

import (
	"context"
	"kitchen_nerd"
	"kitchen_nerd/database"
	"log"
	"os"

	"github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
	"github.com/spf13/cobra"
	"github.com/zeebo/errs"
)

// Error is a default error type for kitchen_nerd cli.
var Error = errs.Class("kitchen_nerd cli error")

// Config contains configurable values for kitchen nerd project.
type Config struct {
	Database            string `json:"database"`
	kitchen_nerd.Config `json:"config"`
}

// commands.
var (
	// kitchen_nerd root cmd.
	rootCmd = &cobra.Command{
		Use:   "kitchen_nerd",
		Short: "cli for interacting with kitchen nerd project",
	}

	runCmd = &cobra.Command{
		Use:         "run",
		Short:       "runs the program",
		RunE:        cmdRun,
		Annotations: map[string]string{"type": "run"},
	}
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
	config := new(kitchen_nerd.Config)

	err = env.Parse(config)
	if err != nil {
		log.Println("could not parse env to config:", err)
		return
	}

	db, err := database.New(ctx, config.DatabaseURL)
	if err != nil {
		//log.Error("Error starting master database on kitchen_nerd service", Error.Wrap(err))
		return Error.Wrap(err)
	}
	defer db.Close()

	// TODO: remove for production.
	err = db.CreateSchema(ctx)
	if err != nil {
		log.Println("Error creating schema", Error.Wrap(err))
	}

	kitchenNerd, err := kitchen_nerd.New(*config, db)
	if err != nil {
		log.Println("Error starting kitchen_nerd service", Error.Wrap(err))
		return Error.Wrap(err)
	}

	runError := kitchenNerd.Run(ctx)
	closeError := kitchenNerd.Close()

	return Error.Wrap(errs.Combine(runError, closeError))
}
