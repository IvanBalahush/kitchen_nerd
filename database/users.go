package database

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"

	"kitchen_nerd/users"
)

const (
	createUser = `INSERT INTO users(id, email, name, phone, status, email_normalized, password_hash, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
)

// ErrUsers indicates that there was an error in the database.
var ErrUsers = errs.Class("users repository error")

// usersDB provides access to users db.
//
// architecture: Database
type usersDB struct {
	pool *pgxpool.Pool
}

// Create creates a user and writes to the database.
func (usersDB *usersDB) Create(ctx context.Context, user *users.User) error {
	emailNormalized := strings.ToUpper(user.Email)

	_, err := usersDB.pool.Exec(ctx, createUser, user.ID, user.Email, user.Name, emailNormalized, user.PasswordHash, user.CreatedAt)

	return ErrUsers.Wrap(err)
}
