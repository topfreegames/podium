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
	"net/http"

	"github.com/labstack/echo"
)

// StatusHandler is the handler responsible for reporting podium status
func StatusHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		c.Set("route", "Status")
		payload := map[string]interface{}{
			"app": map[string]interface{}{
				"errorRate": app.Errors.Rate(),
			},
		}
		return c.JSON(http.StatusOK, payload)
	}
}
