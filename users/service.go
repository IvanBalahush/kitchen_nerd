package users

import (
	"context"
	"golang.org/x/crypto/bcrypt"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"kitchen_nerd/pkg/logger"
)

var (
	// ErrUsers indicates that there was an error in the service.
	ErrUsers = errs.Class("users service error")

	// ErrInvalidPassword should be returned when users password is invalid.
	ErrInvalidPassword = errs.Class("password is invalid")

	// ErrWrongCredentials indicates that user entered wrong credentials.
	ErrWrongCredentials error = errs.New("wrong credentials")
)

// Service is handling users related logic.
//
// architecture: Service
type Service struct {
	log   logger.Logger
	users DB
}

// NewService is a constructor for users service.
func NewService(log logger.Logger, users DB) *Service {
	return &Service{
		log:   log,
		users: users,
	}
}

// Create creates a user.
func (service *Service) Create(ctx context.Context, name, email, password string) (err error) {
	//_, err = service.users.GetByEmail(ctx, email)
	//if err == nil {
	//	return ErrEmailAddressAlreadyInUse.New("user with such email address already exists: %s", email)
	//} else {
	//	if err.Error() != "user does not exist: no rows in result set" {
	//		return ErrUsers.Wrap(err)
	//	}
	//}

	if !IsPasswordValid(password) {
		return ErrInvalidPassword.New("the password must contain at least one lowercase (a-z) letter, one uppercase (A-Z) letter, one digit (0-9) and one special character.")
	}

	id := uuid.New()

	user := User{
		ID:           id,
		Name:         name,
		Email:        email,
		PasswordHash: []byte(password),
		CreatedAt:    time.Now().UTC(),
	}

	if err = user.EncodePass(); err != nil {
		return ErrUsers.Wrap(err)
	}

	return ErrUsers.Wrap(service.users.Create(ctx, &user))
}

func (service *Service) Login(ctx context.Context, email string, password string) error {
	session := NewSession(email, password)
	authToken, err := service.LoginToken(ctx, session)
	if err != nil {
		return ErrUsers.Wrap(err)
	}
}

func (service *Service) LoginToken(ctx context.Context, session Session) (interface{}, error) {
	user, err := service.users.GetByEmail(ctx, session.Email)
	if err != nil {
		return "", ErrWrongCredentials
	}

	if err = bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(session.Password)); err != nil {
		return "", ErrWrongCredentials
	}

	token, err := service.AddSession(ctx, session)
	return token, ErrUsers.Wrap(err)
}

// Session contains user sign-in fields.
type Session struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewSession(email string, password string) *Session {
	return &Session{Email: email, Password: password}
}

// AddSession creates a session.
func (service *Service) AddSession(ctx context.Context, session Session) (string, error) {
	var user *User
	var err error
	user, err = service.users.GetByEmail(ctx, session.Email)
	if err != nil {
		return "", ErrWrongCredentials
	}

	token, err := rand.RandStr(36, rand.TypeToken)
	if err != nil {
		return "", ErrUsers.Wrap(err)
	}

	userToken := tokens.UserToken{
		ID:             uuid.New(),
		UserID:         user.ID,
		Token:          token,
		ExpiredAt:      util.CurrentTime().Add(service.tokensConfig.TokenExpirationTime),
		OSName:         session.OSName,
		OSVersion:      session.OSVersion,
		Browser:        session.Browser,
		BrowserVersion: session.BrowserVersion,
		Device:         session.Device,
		IsMobile:       session.IsMobile,
		IsApp:          session.IsApp,
		CreatedAt:      util.CurrentTime(),
	}

	err = service.tokens.AddToken(ctx, userToken)
	if err != nil {
		return "", ErrAuth.Wrap(err)
	}

	if err = service.users.UpdateLastLogin(ctx, user.ID); err != nil {
		return "", ErrAuth.Wrap(err)
	}

	return token, nil
}
