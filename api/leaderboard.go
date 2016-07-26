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
	"strings"

	"github.com/kataras/iris"
	"github.com/topfreegames/podium/leaderboard"
	"github.com/uber-go/zap"
)

var notFoundError = "Could not find data for member"
var noPageSizeProvidedError = "strconv.ParseInt: parsing \"\": invalid syntax"
var defaultPageSize = 20

type setScorePayload struct {
	Score int
}

type setScoresPayload struct {
	Score        int
	Leaderboards []string
}

func serializeMember(member *leaderboard.Member) map[string]interface{} {
	return map[string]interface{}{
		"publicID": member.PublicID,
		"score":    member.Score,
		"rank":     member.Rank,
	}
}

func serializeMembers(members []*leaderboard.Member) []map[string]interface{} {
	serializedMembers := make([]map[string]interface{}, len(members))
	for i, member := range members {
		serializedMembers[i] = serializeMember(member)
	}
	return serializedMembers
}

// UpsertMemberScoreHandler is the handler responsible for creating or updating the member score
func UpsertMemberScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberScoreHandler"),
		)
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		var payload setScorePayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			app.AddError()
			FailWith(400, err.Error(), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		member, err := l.SetMemberScore(memberPublicID, payload.Score)

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeMember(member), c)
	}
}

// RemoveMemberHandler is the handler responsible for removing a member score
func RemoveMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "RemoveMemberHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		err := l.RemoveMember(memberPublicID)

		if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{}, c)
	}
}

// GetMemberHandler is the handler responsible for retrieving a member score and rank
func GetMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		member, err := l.GetMember(memberPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(serializeMember(member), c)
	}
}

// GetMemberRankHandler is the handler responsible for retrieving a member rank
func GetMemberRankHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		rank, err := l.GetRank(memberPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"publicID": memberPublicID,
			"rank":     rank,
		}, c)
	}
}

// GetAroundMemberHandler retrieves a list of member score and rank centered in the given member
func GetAroundMemberHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetAroundMemberHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		} else if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
			app.AddError()
			FailWith(400, fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", app.Config.GetInt("api.maxReturnedMembers"), pageSize), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		members, err := l.GetAroundMe(memberPublicID)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			app.AddError()
			FailWith(404, "Member not found.", c)
			return
		} else if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members),
		}, c)
	}
}

// GetTotalMembersHandler is the handler responsible for returning the total number of members in a leaderboard
func GetTotalMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTotalMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
		count, err := l.TotalMembers()

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"count": count,
		}, c)
	}
}

// GetTopMembersHandler retrieves onePage of member score and rank
func GetTopMembersHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		pageNumber, err := c.ParamInt("pageNumber")
		if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageNumber provided: %s", err.Error()), c)
			return
		}
		pageSize, err := c.URLParamInt("pageSize")
		if err != nil && err.Error() == noPageSizeProvidedError {
			pageSize = defaultPageSize
		} else if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid pageSize provided: %s", err.Error()), c)
			return
		} else if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
			app.AddError()
			FailWith(400, fmt.Sprintf("Max pageSize allowed: %d. pageSize requested: %d", app.Config.GetInt("api.maxReturnedMembers"), pageSize), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, pageSize, lg)
		members, err := l.GetLeaders(pageNumber)

		if err != nil {
			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members),
		}, c)
	}
}

// GetTopPercentageHandler retrieves top x % members in the leaderboard
func GetTopPercentageHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "GetTopPercentageHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		percentage, err := c.ParamInt("percentage")
		if err != nil {
			app.AddError()
			FailWith(400, fmt.Sprintf("Invalid percentage provided: %s", err.Error()), c)
			return
		}

		l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, defaultPageSize, lg)
		members, err := l.GetTopPercentage(percentage, app.Config.GetInt("api.maxReturnedMembers"))

		if err != nil {
			if err.Error() == "Percentage must be a valid integer between 1 and 100." {
				app.AddError()
				FailWith(400, err.Error(), c)
				return
			}

			app.AddError()
			FailWith(500, err.Error(), c)
			return
		}

		SucceedWith(map[string]interface{}{
			"members": serializeMembers(members),
		}, c)
	}
}

// UpsertMemberLeaderboardsScoreHandler sets the member score for all leaderboards
func UpsertMemberLeaderboardsScoreHandler(app *App) func(c *iris.Context) {
	return func(c *iris.Context) {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberLeaderboardsScoreHandler"),
		)
		memberPublicID := c.Param("memberPublicID")

		var payload setScoresPayload
		if err := LoadJSONPayload(&payload, c); err != nil {
			FailWith(400, err.Error(), c)
			return
		}

		serializedScores := make([]map[string]interface{}, len(payload.Leaderboards))

		for i, leaderboardID := range payload.Leaderboards {
			l := leaderboard.NewLeaderboard(app.RedisClient, leaderboardID, 0, lg)
			member, err := l.SetMemberScore(memberPublicID, payload.Score)

			if err != nil {
				app.AddError()
				FailWith(500, err.Error(), c)
				return
			}
			serializedScore := serializeMember(member)
			serializedScore["leaderboardID"] = leaderboardID
			serializedScores[i] = serializedScore
		}

		SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}
