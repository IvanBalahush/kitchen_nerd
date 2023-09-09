package users

import (
	"context"
	"time"
	"unicode"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// User describes user entity.
type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"passwordHash"`
	CreatedAt    time.Time `json:"createdAt"`
}

// EncodePass encode the password and generate "hash" to store from users password.
func (user *User) EncodePass() error {
	hash, err := bcrypt.GenerateFromPassword(user.PasswordHash, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = hash

	return nil
}

// DB exposes access to users db.
//
// architecture: DB
type DB interface {
	// Create creates a user and writes to the database.
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

// IsPasswordValid check the password for all conditions.
func IsPasswordValid(s string) bool {
	var number, upper bool
	letters := 0
	for _, c := range s {
		switch {
		case unicode.IsNumber(c):
			number = true
		case unicode.IsUpper(c):
			upper = true
		case unicode.IsLetter(c) || c == ' ':
			letters++
		}
	}

	return len(s) >= 8 && letters >= 1 && number && upper
}
