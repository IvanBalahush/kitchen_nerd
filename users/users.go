package users

import (
	"context"
	"html/template"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
	"golang.org/x/crypto/bcrypt"
)

// ErrNoUser indicates that user does not exist.
var ErrNoUser = errs.Class("user does not exist")

// Status represents types of user rights.
type Status string

const (
	// StatusAdmin allows the user to add content, edit information.
	StatusAdmin Status = "admin"
	// StatusUser default type of user with read-only possibility.
	StatusUser Status = "user"
)

// User describes user entity.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	Status       Status    `json:"status"`
	PasswordHash []byte    `json:"passwordHash"`
	LastLogin    time.Time `json:"lastLogin"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Profile describes a user's available fields to check.
type Profile struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Status Status    `json:"status"`
}

func NewProfile(ID uuid.UUID, name string, status Status) *Profile {
	return &Profile{ID: ID, Name: name, Status: status}
}

// UserTemplates holds all users related templates.
type UserTemplates struct {
	List   *template.Template
	Create *template.Template
	Update *template.Template
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
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
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

// Session contains user sign-in fields.
type Session struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewSession(email string, password string) *Session {
	return &Session{Email: email, Password: password}
}
