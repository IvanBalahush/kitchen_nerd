package database

import (
	"context"
	"kitchen_nerd"
	"kitchen_nerd/recipes"
	"kitchen_nerd/tokens"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"

	"kitchen_nerd/users"
)

const (
	notFound = "no rows in result set"
)

var (
	// Error is the default kitchen nerd error class.
	Error = errs.Class("kitchen nerd db error")
)

// ensures that database implements kitchen_nerd.DB.
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
			status                  VARCHAR                  NOT NULL,
			name 	         		VARCHAR                  NOT NULL,
			password_hash    		BYTEA                    NOT NULL,
			last_login              TIMESTAMP WITH TIME ZONE NOT NULL,
			created_at       		TIMESTAMP WITH TIME ZONE NOT NULL
        );
		CREATE TABLE IF NOT EXISTS users_tokens (
		    id              UUID      PRIMARY KEY    NOT NULL,
		    user_id         UUID                     NOT NULL,
		    token           VARCHAR   UNIQUE         NOT NULL,
		    expired_at      TIMESTAMP WITH TIME ZONE NOT NULL,
		    created_at      TIMESTAMP WITH TIME ZONE NOT NULL
		);
CREATE TABLE IF NOT EXISTS recipes (
    id           UUID        PRIMARY KEY,
    title        VARCHAR     NOT NULL,
    photo        VARCHAR,
    description  TEXT,
    instructions TEXT,
    created_at   TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ingredients (
    id   UUID    PRIMARY KEY,
    name VARCHAR NOT NULL,
    unit VARCHAR NOT NULL,
    UNIQUE (name, unit)
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    id            UUID            PRIMARY KEY,
    name          VARCHAR         NOT NULL,
    recipe_id     UUID REFERENCES recipes(id),
    quantity      DOUBLE PRECISION,
    unit          VARCHAR,
    optional      BOOLEAN,
    UNIQUE(name, recipe_id)
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

func (db *database) Tokens() tokens.DB {
	return &tokensDB{pool: db.pool}
}

func (db *database) Recipes() recipes.DB {
	return &recipesDB{pool: db.pool}
}
