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
	"strconv"
	"strings"

	"github.com/buger/jsonparser"
	"github.com/labstack/echo"
	"github.com/topfreegames/podium/leaderboard"
	"go.uber.org/zap"
)

var notFoundError = "Could not find data for member"
var noPageSizeProvidedError = "strconv.ParseInt: parsing \"\": invalid syntax"
var defaultPageSize = 20

func serializeMember(member *leaderboard.Member, position int) map[string]interface{} {
	memberData := map[string]interface{}{
		"publicID": member.PublicID,
		"score":    member.Score,
		"rank":     member.Rank,
	}
	if member.PreviousRank != 0 {
		memberData["previousRank"] = member.PreviousRank
	}
	if position >= 0 {
		memberData["position"] = position
	}
	return memberData
}

func serializeMembers(members []*leaderboard.Member, includePosition bool) []map[string]interface{} {
	serializedMembers := make([]map[string]interface{}, len(members))
	for i, member := range members {
		if includePosition {
			serializedMembers[i] = serializeMember(member, i)
		} else {
			serializedMembers[i] = serializeMember(member, -1)
		}
	}
	return serializedMembers
}

// UpsertMemberScoreHandler is the handler responsible for creating or updating the member score
func UpsertMemberScoreHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberScoreHandler"),
		)
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		var payload setScorePayload
		prevRank := false
		prevRankStr := c.QueryParam("prevRank")
		if prevRankStr != "" && prevRankStr == "true" {
			prevRank = true
		}
		scoreTTL := c.QueryParam("scoreTTL")

		err := WithSegment("Payload", c, func() error {
			b, err := GetRequestBody(c)
			if err != nil {
				app.AddError()
				return err
			}
			if _, err := jsonparser.GetInt(b, "score"); err != nil {
				app.AddError()
				if _, t, _, err := jsonparser.Get(b, "score"); err == nil {
					return fmt.Errorf("invalid type for score: %v", t)
				}
				return fmt.Errorf("score is required")
			}
			if err := LoadJSONPayload(&payload, c, lg); err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var member *leaderboard.Member
		err = WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			member, err = l.SetMemberScore(memberPublicID, payload.Score, prevRank, scoreTTL)

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWithError(err, c)
		}

		return SucceedWith(serializeMember(member, -1), c)
	}
}

// IncrementMemberScoreHandler is the handler responsible for incrementing the member score
func IncrementMemberScoreHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "IncrementMemberScoreHandler"),
		)
		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		scoreTTL := c.QueryParam("scoreTTL")

		var payload incrementScorePayload

		err := WithSegment("Payload", c, func() error {
			if err := LoadJSONPayload(&payload, c, lg); err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var member *leaderboard.Member
		err = WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			member, err = l.IncrementMemberScore(memberPublicID, payload.Increment, scoreTTL)

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWithError(err, c)
		}

		return SucceedWith(serializeMember(member, -1), c)
	}
}

//RemoveMemberHandler removes a member from a leaderboard
func RemoveMemberHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "RemoveMemberHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		err := WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			err := l.RemoveMember(memberPublicID)

			if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{}, c)
	}
}

//RemoveMembersHandler removes several members from a leaderboard
func RemoveMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "RemoveMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		ids := c.QueryParam("ids")

		if ids == "" {
			app.AddError()
			return FailWith(400, "Member IDs are required using the 'ids' querystring parameter", c)
		}

		memberIDs := strings.Split(ids, ",")
		idsInter := make([]interface{}, len(memberIDs))
		for i, v := range memberIDs {
			idsInter[i] = v
		}

		err := WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			err := l.RemoveMembers(idsInter)

			if err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{}, c)
	}
}

// GetMemberHandler is the handler responsible for retrieving a member score and rank
func GetMemberHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberHandler"),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")

		var member *leaderboard.Member
		status := 404
		err := WithSegment("Model", c, func() error {
			var err error
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			member, err = l.GetMember(memberPublicID, order)

			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				app.AddError()
				status = 500
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(serializeMember(member, -1), c)
	}
}

// GetMemberRankHandler is the handler responsible for retrieving a member rank
func GetMemberRankHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		status := 404
		rank := 0
		err := WithSegment("Model", c, func() error {
			var err error
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			rank, err = l.GetRank(memberPublicID, order)

			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				app.AddError()
				status = 500
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"publicID": memberPublicID,
			"rank":     rank,
		}, c)
	}
}

//GetMemberRankInManyLeaderboardsHandler returns the member rank in several leaderboards at once
func GetMemberRankInManyLeaderboardsHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetMemberRankInManyLeaderboardsHandler"),
		)
		memberPublicID := c.Param("memberPublicID")
		ids := c.QueryParam("leaderboardIds")
		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		if ids == "" {
			app.AddError()
			return FailWith(400, "Leaderboard IDs are required using the 'leaderboardIds' querystring parameter", c)
		}

		leaderboardIDs := strings.Split(ids, ",")
		serializedScores := make([]map[string]interface{}, len(leaderboardIDs))

		status := 404
		err := WithSegment("Model", c, func() error {
			for i, leaderboardID := range leaderboardIDs {
				l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
				member, err := l.GetMember(memberPublicID, order)
				if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
					app.AddError()
					status = 404
					return fmt.Errorf("Leaderboard not found or member not found in leaderboard.")
				} else if err != nil {
					app.AddError()
					status = 500
					return err
				}
				serializedScores[i] = map[string]interface{}{
					"leaderboardID": leaderboardID,
					"rank":          member.Rank,
					"score":         member.Score,
				}
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// GetAroundMemberHandler retrieves a list of member score and rank centered in the given member
func GetAroundMemberHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetAroundMemberHandler"),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}
		getLastIfNotFound := false
		getLastIfNotFoundStr := c.QueryParam("getLastIfNotFound")
		if getLastIfNotFoundStr == "true" {
			getLastIfNotFound = true
		}

		leaderboardID := c.Param("leaderboardID")
		memberPublicID := c.Param("memberPublicID")
		pageSize, err := GetPageSize(app, c, defaultPageSize)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var members []*leaderboard.Member
		status := 404
		err = WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, pageSize, lg)
			members, err = l.GetAroundMe(memberPublicID, order, getLastIfNotFound)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				app.AddError()
				status = 404
				return fmt.Errorf("Member not found.")
			} else if err != nil {
				app.AddError()
				status = 500
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
		}, c)
	}
}

// GetTotalMembersHandler is the handler responsible for returning the total number of members in a leaderboard
func GetTotalMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetTotalMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")

		count := 0
		err := WithSegment("Model", c, func() error {
			var err error
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			count, err = l.TotalMembers()

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"count": count,
		}, c)
	}
}

// GetTopMembersHandler retrieves onePage of member score and rank
func GetTopMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetTopMembersHandler"),
		)

		leaderboardID := c.Param("leaderboardID")
		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		pageNumber, err := GetIntRouteParam(app, c, "pageNumber", 1)
		if err != nil {
			app.AddError()
			return FailWith(400, err.Error(), c)
		}

		pageSize, err := GetPageSize(app, c, defaultPageSize)
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		var members []*leaderboard.Member
		err = WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, pageSize, lg)
			members, err = l.GetLeaders(pageNumber, order)

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
		}, c)
	}
}

// GetTopPercentageHandler retrieves top x % members in the leaderboard
func GetTopPercentageHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetTopPercentageHandler"),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		percentageStr := c.Param("percentage")
		percentage, err := strconv.ParseInt(percentageStr, 10, 32)
		if err != nil {
			app.AddError()
			return FailWith(400, fmt.Sprintf("Invalid percentage provided: %s", err.Error()), c)
		}
		if percentage == 0 {
			app.AddError()
			return FailWith(400, "Percentage must be a valid integer between 1 and 100.", c)
		}

		var members []*leaderboard.Member
		status := 400
		err = WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, defaultPageSize, lg)
			members, err = l.GetTopPercentage(int(percentage), app.Config.GetInt("api.maxReturnedMembers"), order)

			if err != nil {
				if err.Error() == "Percentage must be a valid integer between 1 and 100." {
					app.AddError()
					status = 400
					return err
				}

				app.AddError()
				status = 500
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(status, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"members": serializeMembers(members, false),
		}, c)
	}
}

// GetMembersHandler retrieves several members at once
func GetMembersHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "GetMembersHandler"),
		)

		order := c.QueryParam("order")
		if order == "" || (order != "asc" && order != "desc") {
			order = "desc"
		}

		leaderboardID := c.Param("leaderboardID")
		ids := c.QueryParam("ids")
		if ids == "" {
			app.AddError()
			return FailWith(400, "Member IDs are required using the 'ids' querystring parameter", c)
		}

		memberIDs := strings.Split(ids, ",")

		var members []*leaderboard.Member
		err := WithSegment("Model", c, func() error {
			var err error
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, defaultPageSize, lg)
			members, err = l.GetMembers(memberIDs, order)

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		notFound := []string{}

		for _, memberID := range memberIDs {
			found := false
			for _, member := range members {
				if member.PublicID == memberID {
					found = true
					break
				}
			}
			if !found {
				notFound = append(notFound, memberID)
			}
		}

		return SucceedWith(map[string]interface{}{
			"members":  serializeMembers(members, true),
			"notFound": notFound,
		}, c)
	}
}

// UpsertMemberLeaderboardsScoreHandler sets the member score for all leaderboards
func UpsertMemberLeaderboardsScoreHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "UpsertMemberLeaderboardsScoreHandler"),
		)
		memberPublicID := c.Param("memberPublicID")
		scoreTTL := c.QueryParam("scoreTTL")

		var payload setScoresPayload

		prevRank := false
		prevRankStr := c.QueryParam("prevRank")
		if prevRankStr != "" && prevRankStr == "true" {
			prevRank = true
		}

		err := WithSegment("Payload", c, func() error {
			b, err := GetRequestBody(c)
			if err != nil {
				app.AddError()
				return err
			}
			if _, err := jsonparser.GetInt(b, "score"); err != nil {
				app.AddError()
				return fmt.Errorf("score is required")
			}
			if err := LoadJSONPayload(&payload, c, lg); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(400, err.Error(), c)
		}

		serializedScores := make([]map[string]interface{}, len(payload.Leaderboards))

		err = WithSegment("Model", c, func() error {
			for i, leaderboardID := range payload.Leaderboards {
				l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
				member, err := l.SetMemberScore(memberPublicID, payload.Score, prevRank, scoreTTL)

				if err != nil {
					app.AddError()
					return err
				}
				serializedScore := serializeMember(member, -1)
				serializedScore["leaderboardID"] = leaderboardID
				serializedScores[i] = serializedScore
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{
			"scores": serializedScores,
		}, c)
	}
}

// RemoveLeaderboardHandler is the handler responsible for removing a leaderboard
func RemoveLeaderboardHandler(app *App) func(c echo.Context) error {
	return func(c echo.Context) error {
		lg := app.Logger.With(
			zap.String("handler", "RemoveLeaderboardHandler"),
		)
		leaderboardID := c.Param("leaderboardID")

		err := WithSegment("Model", c, func() error {
			l := leaderboard.NewLeaderboard(app.RedisClient.Trace(c.StdContext()), leaderboardID, 0, lg)
			err := l.RemoveLeaderboard()

			if err != nil {
				app.AddError()
				return err
			}
			return nil
		})
		if err != nil {
			return FailWith(500, err.Error(), c)
		}

		return SucceedWith(map[string]interface{}{}, c)
	}
}
