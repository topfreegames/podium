package service_test

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/expiration"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service SetMemberScore", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboard"
	var member string = "member1"
	var score int64 = 1.0
	var previousRank bool = false
	var scoreTTL string = ""

	databaseMembersToInsert := []*database.Member{
		{
			Member: "member1",
			Score:  1.0,
		},
	}

	databaseMembersToGetRank := []string{
		"member1",
	}

	databaseMembersPreviousRankReturned := []*database.Member{
		{
			Member: "member1",
			Score:  2.0,
			Rank:   int64(0),
		},
	}

	databaseMembersReturned := []*database.Member{
		{
			Member: "member1",
			Score:  1.0,
			Rank:   int64(1),
		},
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("When previousRank is false", func() {
		It("Should set Members with previousRank equals zero", func() {
			expectedMember := &model.Member{
				PublicID:     "member1",
				Score:        1,
				PreviousRank: 0,
				Rank:         2,
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member).To(Equal(expectedMember))
		})
	})

	Describe("When previousRank is true", func() {
		previousRank := true

		It("Should set Members with previousRank", func() {
			expectedMember := &model.Member{
				PublicID:     "member1",
				Score:        1,
				PreviousRank: 1,
				Rank:         2,
			}

			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Times(1).Return(databaseMembersPreviousRankReturned, nil)

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member).To(Equal(expectedMember))
		})

		It("Should set a non existent member as rank equals to -1", func() {
			databaseMembersPreviousRankReturned := []*database.Member{nil}
			expectedMember := &model.Member{
				PublicID:     "member1",
				Score:        1,
				PreviousRank: -1,
				Rank:         2,
			}

			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Times(1).Return(databaseMembersPreviousRankReturned, nil)

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member).To(Equal(expectedMember))
		})

		It("Should return error if GetMembers return in error", func() {
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Times(1).Return(nil, fmt.Errorf("New database error"))

			_, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).To(Equal(service.NewGeneralError("set member score", "New database error")))
		})
	})

	Describe("When scoreTTL is empty", func() {
		It("Should SetMembers without filling expire ordered set", func() {
			expectedMember := &model.Member{
				PublicID:     "member1",
				Score:        1,
				PreviousRank: 0,
				Rank:         2,
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member).To(Equal(expectedMember))
		})
	})

	Describe("When scoreTTL is set", func() {
		scoreTTL := "100"

		It("Should SetMembers filling expire ordered set", func() {
			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			mock.EXPECT().SetMembersTTL(gomock.Any(), gomock.Eq(leaderboard), gomock.Any()).Times(1).Return(nil)

			member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member.ExpireAt).To(BeNumerically("~", time.Now().Add(100*time.Second).Unix(), 100))
		})
	})

	Describe("When scoreTTL is invalid", func() {
		scoreTTL := "invalid"

		It("Should return error", func() {
			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
			).Return(databaseMembersReturned, nil)

			_, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
			Expect(err).To(MatchError(service.NewGeneralError("set member score", "strconv.ParseInt: parsing \"invalid\": invalid syntax")))

		})
	})

	It("Should return error if database SetMembers return in error", func() {
		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(fmt.Errorf("New database error"))

		_, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set member score", "New database error")))
	})

	It("Should return error if database GetMembers return in error", func() {
		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(nil, fmt.Errorf("New database error"))

		_, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set member score", "New database error")))

	})

	It("Should not call database GetLeaderboardExpiration and SetLeaderboardExpiration if leaderboard isn't formatted to have an expiration", func() {
		expectedMember := &model.Member{
			PublicID:     "member1",
			Score:        1,
			PreviousRank: 0,
			Rank:         2,
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(databaseMembersReturned, nil)

		member, err := svc.SetMemberScore(context.Background(), leaderboard, member, score, previousRank, scoreTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(member).To(Equal(expectedMember))
	})

	It("Should set leaderboard expiration if GetLeaderboardExpiration return TTLNotFoundError", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(nil)

		_, err = svc.SetMemberScore(context.Background(), leaderboardExpiration, member, score, previousRank, scoreTTL)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error LeaderboardExpiredError if leaderboard was already expired", func() {
		leaderboardExpiration := fmt.Sprintf(
			"testkey-from%dto%d",
			time.Now().UTC().Add(time.Duration(-2)*time.Second).Unix(),
			time.Now().UTC().Add(time.Duration(-1)*time.Second).Unix(),
		)

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(databaseMembersReturned, nil)

		_, err := svc.SetMemberScore(context.Background(), leaderboardExpiration, member, score, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewLeaderboardExpiredError(leaderboardExpiration)))

	})

	It("Should return error if database GetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), fmt.Errorf("New database error"))

		_, err := svc.SetMemberScore(context.Background(), leaderboardExpiration, member, score, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set member score", "New database error")))
	})

	It("Should return error if database SetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(fmt.Errorf("New database error"))

		_, err = svc.SetMemberScore(context.Background(), leaderboardExpiration, member, score, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set member score", "New database error")))
	})

})
