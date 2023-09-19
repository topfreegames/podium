package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/topfreegames/podium/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func NewLogMiddleware(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return (&logMiddleware{logger: logger}).Intercept
}

type logMiddleware struct {
	logger *zap.Logger
}

func (mw *logMiddleware) Intercept(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	l := mw.logger.With(
		zap.String("source", "request"),
	)

	//all except latency to string
	var statusCode int
	var latency time.Duration
	var startTime, endTime time.Time

	startTime = time.Now()

	h, err := handler(ctx, req)

	//no time.Since in order to format it well after
	endTime = time.Now()
	latency = endTime.Sub(startTime)

	_, statusCode = getStatusCodeFromError(err)

	method := info.FullMethod
	reqLog := l.With(
		zap.String("method", method),
		zap.Time("endTime", endTime),
		zap.Int("statusCode", statusCode),
		zap.Duration("latency", latency),
	)

	//request failed
	if statusCode > 399 && statusCode < 500 {
		log.D(reqLog, "Request failed.")
		return h, err
	}

	//request is ok, but server failed
	if statusCode > 499 {
		log.D(reqLog, "Response failed.")
		return h, err
	}

	//Everything went ok
	log.D(reqLog, "Request successful.")
	return h, err
}

func getStatusCodeFromError(err error) (*status.Status, int) {
	var statusCode int
	st, ok := status.FromError(err)
	if !ok {
		statusCode = http.StatusInternalServerError
	} else {
		statusCode = runtime.HTTPStatusFromCode(st.Code())
	}
	return st, statusCode
}
