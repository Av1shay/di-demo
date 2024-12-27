package authentication

import (
	"context"
	"errors"
	"github.com/Av1shay/di-demo/pkg/types"
	"github.com/brianvoe/gofakeit/v7"
)

var demoUser = types.User{
	ID:        gofakeit.UUID(),
	Name:      gofakeit.Name(),
	Email:     gofakeit.Email(),
	AccountID: gofakeit.UUID(),
	Token:     gofakeit.UUID(),
}

type ctxKey string

var userCtxKey = ctxKey("user")

var UsrNotFoundErr = errors.New("user not found")

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (s *Client) Authenticate(ctx context.Context, token string) (*types.User, error) {
	// Here we would normally call the authentication service to validate the token
	// and get the user details
	user := demoUser
	return &user, nil
}

func (s *Client) ContextWithUserAuth(ctx context.Context, token string) (context.Context, error) {
	usr, err := s.Authenticate(ctx, token)
	if err != nil {
		return nil, err
	}
	return ContextWithUser(ctx, usr), nil
}

func ContextWithUser(ctx context.Context, usr *types.User) context.Context {
	return context.WithValue(ctx, userCtxKey, usr)
}

func UserFromContext(ctx context.Context) (*types.User, error) {
	if usr, ok := ctx.Value(userCtxKey).(*types.User); ok {
		return usr, nil
	}
	return nil, UsrNotFoundErr
}
