package middlewares

import (
	"context"
	"fmt"
	"runtime/debug"

	"google.golang.org/grpc"
)

type ErrorHandler func(err error, stack []byte)

func NewRecoverMiddleware(errorHandler ErrorHandler) grpc.UnaryServerInterceptor {
	return (&recoverMiddleware{onError: errorHandler}).Intercept
}

type recoverMiddleware struct {
	onError ErrorHandler
}

func (mw *recoverMiddleware) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			eError, ok := err.(error)
			if !ok {
				eError = fmt.Errorf(fmt.Sprintf("%v", err))
			}
			mw.onError(eError, debug.Stack())
		}
	}()
	return handler(ctx, req)
}
