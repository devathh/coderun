package ssoclient

import (
	"context"

	ssopb "github.com/devathh/coderun/rest-gateway/api/sso/v1"
	customerrors "github.com/devathh/coderun/rest-gateway/pkg/errors"
	"google.golang.org/grpc/metadata"
)

type SSOClient struct {
	client ssopb.SSOClient
}

func New(client ssopb.SSOClient) (*SSOClient, error) {
	if client == nil {
		return nil, customerrors.ErrNilArgs
	}

	return &SSOClient{
		client: client,
	}, nil
}

func (sc *SSOClient) Register(ctx context.Context, req *ssopb.RegisterRequest) (*ssopb.Token, error) {
	resp, err := sc.client.Register(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sc *SSOClient) Login(ctx context.Context, req *ssopb.LoginRequest) (*ssopb.Token, error) {
	resp, err := sc.client.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sc *SSOClient) Refresh(ctx context.Context, req *ssopb.RefreshRequest) (*ssopb.Token, error) {
	resp, err := sc.client.Refresh(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sc *SSOClient) UpdateUser(ctx context.Context, req *ssopb.UpdateRequest, token string) error {
	md := metadata.MD{}
	md.Set("session", token)

	_, err := sc.client.UpdateUser(metadata.NewOutgoingContext(ctx, md), req)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SSOClient) GetUserByID(ctx context.Context, req *ssopb.GetByIDRequest) (*ssopb.User, error) {
	resp, err := sc.client.GetUserByID(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (sc *SSOClient) GetSelf(ctx context.Context, token string) (*ssopb.User, error) {
	md := metadata.MD{}
	md.Set("session", token)

	resp, err := sc.client.GetSelf(metadata.NewOutgoingContext(ctx, md), nil)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
