package users

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"kitchen_nerd/tokens"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

var (
	// ErrUsers indicates that there was an error in the service.
	ErrUsers = errs.Class("users service error")

	// ErrInvalidPassword should be returned when users password is invalid.
	ErrInvalidPassword = errs.Class("password is invalid")

	// ErrWrongCredentials indicates that user entered wrong credentials.
	ErrWrongCredentials = errs.New("wrong credentials")

	// ErrEmailAddressAlreadyInUse indicates that user with current email already exists.
	ErrEmailAddressAlreadyInUse = errs.New("user with such email address already exists")
)

// Service is handling users related logic.
//
// architecture: Service
type Service struct {
	tokens tokens.DB
	users  DB
}

// NewService is a constructor for users service.
func NewService(users DB, tokens tokens.DB) *Service {
	return &Service{
		tokens: tokens,
		users:  users,
	}
}

// Create creates a user.
func (service *Service) Create(ctx context.Context, name, email, password string) (err error) {
	_, err = service.users.GetByEmail(ctx, email)
	if err == nil {
		return ErrEmailAddressAlreadyInUse
	} else {
		if err.Error() != "user does not exist: no rows in result set" {
			return ErrUsers.Wrap(err)
		}
	}

	if !IsPasswordValid(password) {
		return ErrInvalidPassword.New("the password must contain at least one lowercase (a-z) letter, one uppercase (A-Z) letter, one digit (0-9) and one special character.")
	}

	id := uuid.New()

	user := User{
		ID:           id,
		Name:         name,
		Status:       StatusUser,
		Email:        strings.ToLower(email),
		PasswordHash: []byte(password),
		LastLogin:    time.Now().UTC(),
		CreatedAt:    time.Now().UTC(),
	}

	if err = user.EncodePass(); err != nil {
		return ErrUsers.Wrap(err)
	}

	return ErrUsers.Wrap(service.users.Create(ctx, &user))
}

func (service *Service) Login(ctx context.Context, email string, password string) (*tokens.UserToken, error) {
	session := NewSession(email, password)
	authToken, err := service.LoginToken(ctx, session)
	if err != nil {
		return nil, ErrUsers.Wrap(err)
	}

	return authToken, err
}

func (service *Service) LoginToken(ctx context.Context, session *Session) (*tokens.UserToken, error) {
	user, err := service.users.GetByEmail(ctx, session.Email)
	if err != nil {
		return nil, ErrWrongCredentials
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(session.Password)); err != nil {
		return nil, ErrWrongCredentials
	}

	token, err := service.AddSession(ctx, session)

	return token, ErrUsers.Wrap(err)
}

// AddSession creates a session.
func (service *Service) AddSession(ctx context.Context, session *Session) (*tokens.UserToken, error) {
	var user *User
	var err error
	user, err = service.users.GetByEmail(ctx, session.Email)
	if err != nil {
		return nil, ErrWrongCredentials
	}

	userToken, err := tokens.NewUserToken(user.ID, user.Name)
	if err != nil {
		return nil, err
	}

	err = service.tokens.AddToken(ctx, userToken)
	if err != nil {
		return nil, ErrUsers.Wrap(err)
	}

	if err = service.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return nil, ErrUsers.Wrap(err)
	}

	return userToken, nil
}

func (service *Service) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	user, err := service.users.Get(ctx, id)
	return user, ErrUsers.Wrap(err)
}

func (service *Service) GetProfile(ctx context.Context, id uuid.UUID) (*Profile, error) {
	user, err := service.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	profile := NewProfile(id, user.Name, user.Status)
	return profile, nil
}
