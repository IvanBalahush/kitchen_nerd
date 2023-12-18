package database

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/zeebo/errs"

	"kitchen_nerd/users"
)

// ensures that usersDB implements users.DB.
var _ users.DB = (*usersDB)(nil)

const (
	fields     = "id, email, name, status, password_hash, last_login, created_at"
	createUser = `INSERT INTO users(` + fields + `)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
)

// ErrUsers indicates that there was an error in the database.
var ErrUsers = errs.Class("users repository error")

// usersDB provides access to users db.
//
// architecture: Database
type usersDB struct {
	pool *pgxpool.Pool
}

func (usersDB *usersDB) Get(ctx context.Context, id uuid.UUID) (*users.User, error) {
	query := `
SELECT 
	id, 
	email, 
	name,
	password_hash,
	last_login,
	created_at
FROM users WHERE email=$1`

	user := new(users.User)
	err := usersDB.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.LastLogin,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, users.ErrNoUser.Wrap(err)
		}

		return user, ErrUsers.Wrap(err)
	}
	return user, nil
}

func (usersDB *usersDB) GetByEmail(ctx context.Context, email string) (*users.User, error) {
	query := `
SELECT 
	id, 
	email, 
	name,
	password_hash,
	last_login,
	created_at
FROM users WHERE email=$1`

	user := new(users.User)
	err := usersDB.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.PasswordHash,
		&user.LastLogin,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user, users.ErrNoUser.Wrap(err)
		}

		return user, ErrUsers.Wrap(err)
	}
	return user, nil
}

// Create creates a user and writes to the database.
func (usersDB *usersDB) Create(ctx context.Context, user *users.User) error {
	_, err := usersDB.pool.Exec(ctx, createUser, user.ID, user.Email, user.Name, user.Status, user.PasswordHash, user.LastLogin, user.CreatedAt)

	return ErrUsers.Wrap(err)
}

// UpdateLastLogin updates last login time.
func (usersDB *usersDB) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET last_login=$1 WHERE id=$2`
	result, err := usersDB.pool.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return ErrUsers.Wrap(err)
	}

	rowNum := result.RowsAffected()
	if rowNum == 0 {
		return users.ErrNoUser.New("")
	}

	return nil
}
