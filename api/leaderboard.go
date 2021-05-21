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
	"math"
	"strings"

	lmodel "github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

const (
	notFoundError   = "Could not find data for member"
	defaultPageSize = 20
)

func validateBulkUpsertScoresRequest(req *api.BulkUpsertScoresRequest) error {
	for _, m := range req.MemberScores.Members {
		if m.PublicID == "" {
			return status.Errorf(codes.InvalidArgument, "publicID is required")
		}
	}
	return nil
}

// BulkUpsertScores is the handler responsible for creating or updating members score.
func (app *App) BulkUpsertScores(ctx context.Context, req *api.BulkUpsertScoresRequest) (*api.BulkUpsertScoresResponse, error) {
	if err := validateBulkUpsertScoresRequest(req); err != nil {
		return nil, err
	}

	lg := app.Logger.With(
		zap.String("handler", "BulkUpsertScores"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	members := make([]*lmodel.Member, len(req.MemberScores.Members))

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Setting member scores.")
		for i, ms := range req.MemberScores.Members {
			members[i] = &lmodel.Member{Score: int64(ms.Score), PublicID: ms.PublicID}
		}

		if err := app.Leaderboards.SetMembersScore(ctx, req.LeaderboardId, members, req.PrevRank, getScoreTTL(req.ScoreTTL)); err != nil {
			lg.Error("Setting member scores failed.", zap.Error(err))
			app.AddError()
			//TODO: Turn all these LeaderboardExpiredError verifications into a middleware
			if _, ok := err.(*service.LeaderboardExpiredError); ok {
				return status.Errorf(codes.InvalidArgument, err.Error())
			}
			return err
		}
		lg.Debug("Setting member scores succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	responses := make([]*api.BulkUpsertScoresResponse_Member, len(members))

	for i, m := range members {
		responses[i] = &api.BulkUpsertScoresResponse_Member{
			PublicID:     m.PublicID,
			Score:        float64(m.Score),
			Rank:         int32(m.Rank),
			PreviousRank: int32(m.PreviousRank),
			ExpireAt:     int32(m.ExpireAt),
		}
	}

	return &api.BulkUpsertScoresResponse{Success: true, Members: responses}, nil
}

func getScoreTTL(scoreTTL int32) string {
	if scoreTTL == 0 {
		return ""
	}

	return fmt.Sprint(scoreTTL)
}

// UpsertScore is the handler responsible for creating or updating the member score.
func (app *App) UpsertScore(ctx context.Context, req *api.UpsertScoreRequest) (*api.UpsertScoreResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "UpsertScore"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	var member *lmodel.Member
	err := withSegment("Model", ctx, func() error {
		lg.Debug("Setting member score.", zap.Int64("score", int64(req.ScoreChange.Score)))

		var err error
		member, err = app.Leaderboards.SetMemberScore(
			ctx, req.LeaderboardId, req.MemberPublicId, int64(req.ScoreChange.Score), req.PrevRank, getScoreTTL(req.ScoreTTL))

		if err != nil {
			lg.Error("Setting member score failed.", zap.Error(err))
			app.AddError()
			if _, ok := err.(*service.LeaderboardExpiredError); ok {
				return status.Errorf(codes.InvalidArgument, err.Error())
			}

			return err
		}
		lg.Debug("Setting member score succeeded.")
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &api.UpsertScoreResponse{
		Success:      true,
		PublicID:     member.PublicID,
		Score:        float64(member.Score),
		Rank:         int32(member.Rank),
		PreviousRank: int32(member.PreviousRank),
		ExpireAt:     int32(member.ExpireAt),
	}, nil
}

// IncrementScore is the handler responsible for incrementing the member score.
func (app *App) IncrementScore(ctx context.Context, req *api.IncrementScoreRequest) (*api.IncrementScoreResponse, error) {
	if req.Body.Increment == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "increment is required")
	}

	lg := app.Logger.With(
		zap.String("handler", "IncrementScore"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	var member *lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Incrementing member score.", zap.Int64("increment", int64(req.Body.Increment)))
		member, err = app.Leaderboards.IncrementMemberScore(context.Background(), req.LeaderboardId, req.MemberPublicId,
			int(req.Body.Increment), getScoreTTL(req.ScoreTTL))

		if err != nil {
			lg.Error("Member score increment failed.", zap.Error(err))
			app.AddError()
			if _, ok := err.(*service.LeaderboardExpiredError); ok {
				return status.Errorf(codes.InvalidArgument, err.Error())
			}

			return err
		}
		lg.Debug("Member score increment succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.IncrementScoreResponse{
		Success:      true,
		PublicID:     member.PublicID,
		Score:        float64(member.Score),
		Rank:         int32(member.Rank),
		PreviousRank: int32(member.PreviousRank),
		ExpireAt:     int32(member.ExpireAt),
	}, nil
}

//TODO: Make this function use RemoveMembers
// RemoveMember removes a member from a leaderboard.
func (app *App) RemoveMember(ctx context.Context, req *api.RemoveMemberRequest) (*api.RemoveMemberResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "RemoveMember"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing member.")

		//TODO: implement an operation that checks and if exists removes the member atomically, removing the need to check an error string.
		if err := app.Leaderboards.RemoveMember(ctx, req.LeaderboardId, req.MemberPublicId); err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member removal failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Member removal succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.RemoveMemberResponse{Success: true}, nil

}

// RemoveMembers removes several members from a leaderboard.
func (app *App) RemoveMembers(ctx context.Context, req *api.RemoveMembersRequest) (*api.RemoveMembersResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "RemoveMembers"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	if req.Ids == "" {
		app.AddError()
		return nil, status.Errorf(codes.InvalidArgument, "Member IDs are required using the 'ids' querystring parameter")
	}

	memberIDs := strings.Split(req.Ids, ",")
	idsInter := make([]string, len(memberIDs))
	for i, v := range memberIDs {
		idsInter[i] = v
	}

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing members.", zap.String("ids", req.Ids))

		if err := app.Leaderboards.RemoveMembers(ctx, req.LeaderboardId, idsInter); err != nil && !strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Members removal failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Members removal succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.RemoveMembersResponse{Success: true}, nil
}

func getOrder(order string) string {
	if order == "" || (order != "asc" && order != "desc") {
		return "desc"
	}
	return order
}

// GetMember is the handler responsible for retrieving a member score and rank.
func (app *App) GetMember(ctx context.Context, req *api.GetMemberRequest) (*api.GetMemberResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetMember"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	order := getOrder(req.Order)

	var member *lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting member.")
		//TODO: Add a NotFound error on the library
		member, err = app.Leaderboards.GetMember(ctx, req.LeaderboardId, req.MemberPublicId, order, req.ScoreTTL)
		switch {
		case err != nil && strings.HasPrefix(err.Error(), notFoundError):
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.Errorf(codes.NotFound, "Member not found.")
		case err != nil:
			lg.Error("Get member failed.")
			app.AddError()
			return err
		}
		lg.Debug("Getting member succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetMemberResponse{
		Success:      true,
		PublicID:     member.PublicID,
		Score:        float64(member.Score),
		Rank:         int32(member.Rank),
		PreviousRank: int32(member.PreviousRank),
		ExpireAt:     int32(member.ExpireAt),
	}, nil
}

// GetRank is the handler responsible for retrieving a member rank.
func (app *App) GetRank(ctx context.Context, req *api.GetRankRequest) (*api.GetRankResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetRank"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	order := getOrder(req.Order)

	var rank int
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting rank.")
		rank, err = app.Leaderboards.GetRank(ctx, req.LeaderboardId, req.MemberPublicId, order)

		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.Errorf(codes.NotFound, "Member not found.")
		} else if err != nil {
			lg.Error("Getting rank failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting rank succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetRankResponse{
		Success:  true,
		PublicID: req.MemberPublicId,
		Rank:     int32(rank),
	}, nil
}

// GetRankMultiLeaderboards returns the member rank req several leaderboards at once.
func (app *App) GetRankMultiLeaderboards(ctx context.Context, req *api.GetRankMultiLeaderboardsRequest) (*api.GetRankMultiLeaderboardsResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetRankMultiLeaderboards"),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	order := getOrder(req.Order)

	if req.LeaderboardIds == "" {
		app.AddError()
		return nil, status.Errorf(codes.InvalidArgument, "Leaderboard IDs are required using the 'leaderboardIds' querystring parameter")
	}

	leaderboardIDs := strings.Split(req.LeaderboardIds, ",")
	serializedScores := make([]*api.GetRankMultiLeaderboardsResponse_Member, len(leaderboardIDs))

	err := withSegment("Model", ctx, func() error {
		for i, leaderboardID := range leaderboardIDs {
			lg.Debug("Getting member rank on leaderboard.", zap.String("leaderboard", leaderboardID))
			member, err := app.Leaderboards.GetMember(ctx, leaderboardID, req.MemberPublicId, order, req.ScoreTTL)
			if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
				lg.Error("Member not found.", zap.Error(err))
				app.AddError()
				return status.Errorf(codes.NotFound, "Leaderboard not found or member not found req leaderboard.")
			} else if err != nil {
				lg.Error("Getting member rank on leaderboard failed.", zap.Error(err))
				app.AddError()
				return err
			}
			lg.Debug("Getting member rank on leaderboard succeeded.")
			serializedScores[i] = &api.GetRankMultiLeaderboardsResponse_Member{
				LeaderboardID: leaderboardID,
				Rank:          int32(member.Rank),
				Score:         float64(member.Score),
				ExpireAt:      int32(member.ExpireAt),
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetRankMultiLeaderboardsResponse{
		Success: true,
		Scores:  serializedScores,
	}, nil
}

// GetAroundMember retrieves a list of member score and rank centered req the given member.
func (app *App) GetAroundMember(ctx context.Context, req *api.GetAroundMemberRequest) (*api.GetAroundMemberResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetAroundMember"),
		zap.String("leaderboard", req.LeaderboardId),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	order := getOrder(req.Order)

	pageSize := getPageSize(int(req.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.Errorf(codes.InvalidArgument, msg)
	}

	var members []*lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting members around player.")
		members, err = app.Leaderboards.GetAroundMe(ctx, req.LeaderboardId, pageSize, req.MemberPublicId, order,
			req.GetLastIfNotFound)
		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.Errorf(codes.NotFound, "Member not found.")
		} else if err != nil {
			lg.Error("Getting members around player failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting members around player succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetAroundMemberResponse{
		Success: true,
		Members: newMemberRankResponseList(members),
	}, nil
}

func getPageSize(pageSize int) int {
	if pageSize == 0 {
		pageSize = defaultPageSize
	}

	return pageSize
}

// GetAroundScore retrieves a list of member scores and ranks centered req a given score.
func (app *App) GetAroundScore(ctx context.Context, req *api.GetAroundScoreRequest) (*api.GetAroundScoreResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetAroundScoreHandler"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	order := getOrder(req.Order)

	pageSize := getPageSize(int(req.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.Errorf(codes.InvalidArgument, msg)
	}

	var members []*lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting players around score.", zap.Int64("score", int64(req.Score)))
		members, err = app.Leaderboards.GetAroundScore(ctx, req.LeaderboardId, pageSize, int64(req.Score), order)
		if err != nil && strings.HasPrefix(err.Error(), notFoundError) {
			lg.Error("Member not found.", zap.Error(err))
			app.AddError()
			return status.Errorf(codes.NotFound, "Member not found.")
		} else if err != nil {
			lg.Error("Getting players around score failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting players around score succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetAroundScoreResponse{
		Success: true,
		Members: newMemberRankResponseList(members),
	}, nil
}

// TotalMembers is the handler responsible for returning the total number of members req a leaderboard.
func (app *App) TotalMembers(ctx context.Context, req *api.TotalMembersRequest) (*api.TotalMembersResponse, error) {
	leaderboardID := req.LeaderboardId
	lg := app.Logger.With(
		zap.String("handler", "TotalMembers"),
		zap.String("leaderboard", leaderboardID),
	)

	var count int
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting total members.")
		count, err = app.Leaderboards.TotalMembers(ctx, leaderboardID)

		if err != nil {
			lg.Error("Getting total members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting total members succeeded.")
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &api.TotalMembersResponse{Success: true, Count: int32(count)}, nil
}

// GetTopMembers retrieves onePage of member score and rank.
func (app *App) GetTopMembers(ctx context.Context, req *api.GetTopMembersRequest) (*api.GetTopMembersResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetTopMembers"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	pageNumber := int(math.Max(float64(req.PageNumber), 1))

	order := getOrder(req.Order)

	pageSize := getPageSize(int(req.PageSize))
	if pageSize > app.Config.GetInt("api.maxReturnedMembers") {
		msg := fmt.Sprintf(
			"Max pageSize allowed: %d. pageSize requested: %d",
			app.Config.GetInt("api.maxReturnedMembers"),
			pageSize,
		)
		return nil, status.Errorf(codes.InvalidArgument, msg)
	}

	var members []*lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting top members.")
		members, err = app.Leaderboards.GetLeaders(ctx, req.LeaderboardId, pageSize, pageNumber, order)

		if err != nil {
			lg.Error("Getting top members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting top members succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetTopMembersResponse{
		Success: true,
		Members: newMemberRankResponseList(members),
	}, nil
}

// GetTopPercentage retrieves top x % members req the leaderboard.
func (app *App) GetTopPercentage(ctx context.Context, req *api.GetTopPercentageRequest) (*api.GetTopPercentageResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetTopPercentage"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	if req.Percentage == 0 {
		app.AddError()
		return nil, status.Errorf(codes.InvalidArgument, "Percentage must be a valid integer between 1 and 100.")
	}

	order := getOrder(req.Order)

	var members []*lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting top percentage.", zap.Int("percentage", int(req.Percentage)))
		members, err = app.Leaderboards.GetTopPercentage(ctx, req.LeaderboardId, defaultPageSize,
			int(req.Percentage), app.Config.GetInt("api.maxReturnedMembers"), order)

		if err != nil {
			lg.Error("Getting top percentage failed.", zap.Error(err))
			if _, ok := err.(*service.PercentageError); ok {
				app.AddError()
				return status.Errorf(codes.InvalidArgument, err.Error())
			}

			app.AddError()
			return err
		}
		lg.Debug("Getting top percentage succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.GetTopPercentageResponse{
		Success: true,
		Members: newMemberRankResponseList(members),
	}, nil
}

func newGetMembersResponseList(members []*lmodel.Member) []*api.GetMembersResponse_Member {
	list := make([]*api.GetMembersResponse_Member, len(members))
	for i, m := range members {
		list[i] = &api.GetMembersResponse_Member{
			PublicID: m.PublicID,
			Score:    float64(m.Score),
			Rank:     int32(m.Rank),
			ExpireAt: int32(m.ExpireAt),
			Position: int32(i),
		}
	}
	return list
}

// GetMembers retrieves several members at once.
func (app *App) GetMembers(ctx context.Context, req *api.GetMembersRequest) (*api.GetMembersResponse, error) {
	lg := app.Logger.With(
		zap.String("handler", "GetMembers"),
		zap.String("leaderboard", req.LeaderboardId),
	)

	order := getOrder(req.Order)

	if req.Ids == "" {
		app.AddError()
		return nil, status.Error(codes.InvalidArgument, "Member IDs are required using the 'ids' querystring parameter")
	}

	memberIDs := strings.Split(req.Ids, ",")

	var members []*lmodel.Member
	err := withSegment("Model", ctx, func() error {
		var err error
		lg.Debug("Getting members.", zap.String("ids", req.Ids))
		members, err = app.Leaderboards.GetMembers(ctx, req.LeaderboardId, memberIDs, order, req.ScoreTTL)

		if err != nil {
			lg.Error("Getting members failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Getting members succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	var notFound []string

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

	return &api.GetMembersResponse{
		Success:  true,
		Members:  newGetMembersResponseList(members),
		NotFound: notFound,
	}, nil
}

func newMemberRankResponseList(members []*lmodel.Member) []*api.Member {
	list := make([]*api.Member, len(members))
	for i, m := range members {
		list[i] = &api.Member{
			PublicID: m.PublicID,
			Score:    float64(m.Score),
			Rank:     int32(m.Rank),
		}
	}
	return list
}

// UpsertScoreMultiLeaderboards sets the member score for all leaderboards.
func (app *App) UpsertScoreMultiLeaderboards(ctx context.Context, req *api.UpsertScoreMultiLeaderboardsRequest) (*api.UpsertScoreMultiLeaderboardsResponse, error) {
	if len(req.ScoreMultiChange.Leaderboards) == 0 {
		return nil, status.Error(codes.InvalidArgument, "leaderboards is required")
	}

	lg := app.Logger.With(
		zap.String("handler", "UpsertScoreAllLeaderboards"),
		zap.String("memberPublicID", req.MemberPublicId),
	)

	serializedScores := make([]*api.UpsertScoreMultiLeaderboardsResponse_Member, len(req.ScoreMultiChange.Leaderboards))

	err := withSegment("Model", ctx, func() error {
		for i, leaderboardID := range req.ScoreMultiChange.Leaderboards {
			lg.Debug("Updating score.",
				zap.String("leaderboardID", leaderboardID),
				zap.Int64("score", int64(req.ScoreMultiChange.Score)))

			member, err := app.Leaderboards.SetMemberScore(ctx, leaderboardID, req.MemberPublicId,
				int64(req.ScoreMultiChange.Score), req.PrevRank, getScoreTTL(req.ScoreTTL))

			if err != nil {
				lg.Error("Update score failed.", zap.Error(err))
				app.AddError()
				return err
			}
			serializedScore := &api.UpsertScoreMultiLeaderboardsResponse_Member{
				PublicID:      member.PublicID,
				Score:         float64(member.Score),
				Rank:          int32(member.Rank),
				PreviousRank:  int32(member.PreviousRank),
				ExpireAt:      int32(member.ExpireAt),
				LeaderboardID: leaderboardID,
			}
			serializedScores[i] = serializedScore
		}
		lg.Debug("Update score succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.UpsertScoreMultiLeaderboardsResponse{
		Success: true,
		Scores:  serializedScores,
	}, nil
}

// RemoveLeaderboard is the handler responsible for removing a leaderboard.
func (app *App) RemoveLeaderboard(ctx context.Context, req *api.RemoveLeaderboardRequest) (*api.RemoveLeaderboardResponse, error) {
	leaderboardID := req.LeaderboardId
	lg := app.Logger.With(
		zap.String("handler", "RemoveLeaderboard"),
		zap.String("leaderboard", leaderboardID),
	)

	err := withSegment("Model", ctx, func() error {
		lg.Debug("Removing leaderboard.")

		if err := app.Leaderboards.RemoveLeaderboard(ctx, leaderboardID); err != nil {
			lg.Error("Remove leaderboard failed.", zap.Error(err))
			app.AddError()
			return err
		}
		lg.Debug("Remove leaderboard succeeded.")
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &api.RemoveLeaderboardResponse{Success: true}, nil
}
