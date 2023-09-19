package middlewares

import (
	"context"
	"encoding/base64"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewMetricsMiddleware(options ...Option) grpc.UnaryServerInterceptor {
	return (&metricsMiddleware{username: username, password: password}).Intercept
}

type Option func(*Options) error

type Options struct {
	// Namespace to prepend to all metrics, events and service checks name.
	Namespace string
	// Tags are global tags to be applied to every metrics, events and service checks.
	Tags []string
	Rate
}


	config.SetDefault("extensions.dd.namespace", "middleware_dev.")
	config.SetDefault("extensions.dd.rate", "1")

type metricsMiddleware struct {
	username string
	password string
}

func (mw *metricsMiddleware) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	startTime := time.Now()
	h, err := handler(ctx, req)
	timeUsed := time.Since(startTime)
	_, st := getStatusCodeFromError(err)
	method := info.FullMethod

	tags := []string{
		fmt.Sprintf("method:%s", method),
		fmt.Sprintf("status:%d", st),
	}

	if err := app.DDStatsD.Timing("response_time_milliseconds", timeUsed, tags...); err != nil {
		app.Logger.Error("DDStatsD Timing", zap.Error(err))
	}

	return h, err
}