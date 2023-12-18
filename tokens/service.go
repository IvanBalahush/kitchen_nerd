package tokens

import (
	"context"
	"github.com/zeebo/errs"
	"time"
)

// ErrTokens indicates that there was an error in the service.
var ErrTokens = errs.Class("users service error")

// Config defines configuration for tokens.
type Config struct {
	TokenExpirationTime time.Duration `json:"tokenExpirationTime"`
}

// Service is handling tokens related logic.
//
// architecture: Service
type Service struct {
	config Config
	tokens DB
}

// GetToken returns UserToken by token.
func (service *Service) GetToken(ctx context.Context, token string) (UserToken, error) {
	userToken, err := service.tokens.GetToken(ctx, token)

	return userToken, ErrTokens.Wrap(err)
}

// NewService is a constructor for tokens service.
func NewService(config Config, tokens DB) *Service {
	return &Service{
		config: config,
		tokens: tokens,
	}
}
