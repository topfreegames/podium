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
	"fmt"

	"github.com/kataras/iris"
	"github.com/topfreegames/podium/leaderboard"
)

var notFoundError = "redigo: nil returned"
var defaultLimit = 20

type setScorePayload struct {
	Score int
}

func serializeUser(user *leaderboard.User) map[string]interface{} {
	return map[string]interface{}{
		"publicID": user.PublicID,
		"score":    user.Score,
		"rank":     user.Rank,
	}
}

func serializeUsers(users []leaderboard.User) []map[string]interface{} {
	serializedUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		serializedUsers[i] = serializeUser(&user)
	}
	return serializedUsers
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

		SucceedWith(serializeUser(&user), c)
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

		SucceedWith(serializeUser(&user), c)
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

// GetAroundUserHandler is the handler responsible for remover a user score
func GetAroundUserHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		leaderboardID := c.Param("leaderboardID")
		userPublicID := c.Param("userPublicID")
		limit, err := c.URLParamInt("limit")
		if err != nil && err.Error() == "strconv.ParseInt: parsing \"\": invalid syntax" {
			limit = defaultLimit
		} else if err != nil {
			FailWith(400, fmt.Sprintf("Invalid limit provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, limit)
		users, err := l.GetAroundMe(userPublicID)

		if err != nil && err.Error() == notFoundError {
			FailWith(404, "User not found.", c)
			return
		} else if err != nil {
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"users": serializeUsers(users),
		}, c)
	}
}
