package middlewares

import (
	"context"
	"encoding/base64"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewBasicAuthMiddleware(username, password string) grpc.UnaryServerInterceptor {
	return (&basicAuthMiddleware{username: username, password: password}).Intercept
}

type basicAuthMiddleware struct {
	username string
	password string
}

func (mw *basicAuthMiddleware) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "basic")
	if err != nil {
		return nil, err
	}
	if token != base64.StdEncoding.EncodeToString([]byte(mw.username+":"+mw.password)) {
		return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
	}
	return ctx, nil
}
