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
	"github.com/kataras/iris"
	"github.com/topfreegames/podium/leaderboard"
)

var notFoundError = "redigo: nil returned"

type setScorePayload struct {
	Score int
}

// UpsertUserScoreHandler is the handler responsible for creating or updating the user score
func UpsertUserScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		var payload setScorePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0)
		user, err := l.SetUserScore(userPublicID, payload.Score)

		if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": user.PublicID,
			"score":    user.Score,
			"rank":     user.Rank,
		}, c)
	}
}

// RemoveUserHandler is the handler responsible for removing a user score
func RemoveUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0)
		_, err := l.RemoveMember(userPublicID)

		if err != nil && err.Error() != notFoundError {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// GetUserHandler is the handler responsible for retrieving a user score and rank
func GetUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0)
		user, err := l.GetMember(userPublicID)

		if err != nil && err.Error() == notFoundError {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": user.PublicID,
			"score":    user.Score,
			"rank":     user.Rank,
		}, c)
	}
}

// GetUserRankHandler is the handler responsible for remover a user score
func GetUserRankHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0)
		rank, err := l.GetRank(userPublicID)

		if err != nil && err.Error() == notFoundError {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": userPublicID,
			"rank":     rank,
		}, c)
	}
}
