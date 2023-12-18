package tokens

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"

	"kitchen_nerd/pkg/rand"
)

// ErrNoToken indicates that token does not exist.
var ErrNoToken = errs.Class("token does not exist")

type DB interface {
	// GetToken returns user's token from the database.
	GetToken(ctx context.Context, token string) (UserToken, error)
	// GetTokenByID returns user's token from the database.
	GetTokenByID(ctx context.Context, id uuid.UUID) (UserToken, error)
	// AddToken inserts a token in the database.
	AddToken(ctx context.Context, token *UserToken) error
	// DeleteToken removes a token from the database.
	DeleteToken(ctx context.Context, token string) error
	// DeleteTokenByUserId removes a token from the database by user id.
	DeleteTokenByUserId(ctx context.Context, id uuid.UUID) error
	// ListActiveSessions gets all sessions by user id form the database.
	ListActiveSessions(ctx context.Context, userID uuid.UUID) ([]UserToken, error)
	// DeleteSessionToken removes a token from the database by session id.
	DeleteSessionToken(ctx context.Context, userId, sessionId uuid.UUID) error
	//// AddAdminSession inserts an access token in tha database.
	//AddAdminSession(ctx context.Context, session AdminSession) error
	//// GetAdminSession returns an admin token from the database.
	//GetAdminSession(ctx context.Context, token string) (AdminSession, error)
	// DeleteAdminSession removes an admin session from the database.
	//DeleteAdminSession(ctx context.Context, token string) error
}

type UserToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"userID"`
	Username  string    `json:"username"`
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expiredAt"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewUserToken(
	userID uuid.UUID,
	username string,
) (*UserToken, error) {
	token, err := rand.RandStr(36, rand.TypeToken)
	if err != nil {
		return nil, ErrTokens.Wrap(err)
	}

	return &UserToken{
		ID:        uuid.New(),
		UserID:    userID,
		Username:  username,
		Token:     token,
		ExpiredAt: time.Now().Add(time.Minute * 60 * 24 * 5),
		CreatedAt: time.Now(),
	}, nil
}
