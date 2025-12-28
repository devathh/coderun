package ssoservice

import (
	"context"

	ssopb "github.com/devathh/coderun/rest-gateway/api/sso/v1"
)

type SSOClient interface {
	Register(context.Context, *ssopb.RegisterRequest) (*ssopb.Token, error)
	Login(context.Context, *ssopb.LoginRequest) (*ssopb.Token, error)
	Refresh(context.Context, *ssopb.RefreshRequest) (*ssopb.Token, error)
	UpdateUser(context.Context, *ssopb.UpdateRequest, string) error
	GetUserByID(context.Context, *ssopb.GetByIDRequest) (*ssopb.User, error)
}
