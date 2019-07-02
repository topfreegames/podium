// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.uber.org/zap"

	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

//healthCheckHandler is the handler responsible for validating that the app is still up
func (app *App) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	workingString := app.Config.GetString("healthcheck.workingText")
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	res, err := app.Leaderboards.Ping(context.Background())
	var msg string
	if err != nil || res != "PONG" {
		w.WriteHeader(http.StatusInternalServerError)
		msg = newFailMsg(fmt.Sprintf("Error connecting to redis: %v", err))
	} else {
		msg = workingString
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		app.Logger.Error("Error writing /healthcheck response", zap.Error(err))
	}
}

func (app *App) HealthCheck(ctx context.Context, in *api.HealthCheckRequest) (*api.HealthCheckResponse, error) {
	var res string

	err := withSegment("redis", ctx, func() error {
		var err error
		res, err = app.Leaderboards.Ping(ctx)
		if err != nil {
			return fmt.Errorf("Error trying to ping redis: %v", err)
		} else if res != "PONG" {
			return fmt.Errorf("Redis return = %s, want PONG", res)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	workingString := strings.TrimSpace(app.Config.GetString("healthcheck.workingText"))

	return &api.HealthCheckResponse{WorkingString: workingString}, nil
}
