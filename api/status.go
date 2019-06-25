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
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/golang/protobuf/ptypes/empty"
	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

// StatusHandler is the handler responsible for reporting podium status
func (app *App) statusHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	payload := map[string]interface{}{
		"app": map[string]interface{}{
			"errorRate": app.Errors.Rate(),
		},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		errMsg := fmt.Sprintf("JSON marshaling failed: %v", err)
		app.Logger.Error(errMsg)
		data = []byte(newFailMsg(errMsg))
		w.WriteHeader(http.StatusInternalServerError)
	}

	if _, err := w.Write([]byte(data)); err != nil {
		app.Logger.Error("Error writing /status response", zap.Error(err))
	}
}

func (app *App) Status(ctx context.Context, in *empty.Empty) (*api.StatusResponse, error) {
	return &api.StatusResponse{ErrorRate: app.Errors.Rate()}, nil
}
