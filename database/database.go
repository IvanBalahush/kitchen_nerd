package database

import (
	"context"
	"kitchen_nerd"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"

	"kitchen_nerd/users"
)

const (
	notFound = "no rows in result set"
)

var (
	// Error is the default fastwallet error class.
	Error = errs.Class("fastwallet db error")
)

// ensures that database implements fastwallet.DB.
var _ kitchen_nerd.DB = (*database)(nil)

// database combines access to different database tables with a record
// of the db driver, db implementation, and db source URL.
//
// architecture: Master Database
type database struct {
	pool *pgxpool.Pool
}

// New returns kitchenNerd.DB postgresql implementation.
func New(ctx context.Context, databaseURL string) (kitchen_nerd.DB, error) {
	pool, err := pgxpool.Connect(ctx, databaseURL)
	if err != nil {
		return nil, Error.Wrap(err)
	}

	return &database{pool: pool}, nil
}

// CreateSchema create schema for all tables and databases.
func (db *database) CreateSchema(ctx context.Context) (err error) {
	createTableQuery := `
        CREATE TABLE IF NOT EXISTS users (
			id               		UUID PRIMARY KEY         NOT NULL,
			email            		VARCHAR                  NOT NULL,
			name 	         		VARCHAR                  NOT NULL,
			password_hash    		BYTEA                    NOT NULL,
			created_at       		TIMESTAMP WITH TIME ZONE NOT NULL
        );
	`

	_, err = db.pool.Exec(ctx, createTableQuery)
	if err != nil {
		return Error.Wrap(err)
	}

	return nil
}

// Close closes underlying db connection.
func (db *database) Close() {
	db.pool.Close()
}

// Users provides access to users db.
func (db *database) Users() users.DB {
	return &usersDB{pool: db.pool}
}
