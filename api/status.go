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
	"encoding/json"

	"github.com/kataras/iris"
)

// StatusHandler is the handler responsible for reporting podium status
func StatusHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		c.Set("route", "Status")
		payload := map[string]interface{}{
			"app": map[string]interface{}{
				"errorRate": app.Errors.Rate(),
			},
		}

		payloadJSON, _ := json.Marshal(payload)
		c.SetStatusCode(iris.StatusOK)
		c.Write(string(payloadJSON))
	}
}
