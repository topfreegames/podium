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

var _ = Describe("Service IncrementMemberScore", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboard"
	var member string = "member1"
	var score int = 1.0
	var scoreTTL string = ""

	databaseMembersReturned := []*database.Member{
		{
			Member: "member1",
			Score:  2.0,
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

	It("Should increment member score if all is ok", func() {
		expectedMember := &model.Member{
			PublicID:     "member1",
			Score:        2,
			PreviousRank: 0,
			Rank:         2,
		}

		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)

		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)

		member, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(member).To(Equal(expectedMember))
	})

	Describe("When scoreTTL is empty", func() {
		It("Should IncrementMember without filling expire ordered set", func() {
			expectedMember := &model.Member{
				PublicID:     "member1",
				Score:        2,
				PreviousRank: 0,
				Rank:         2,
			}

			mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(member),
			).Return(databaseMembersReturned, nil)

			member, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member).To(Equal(expectedMember))
		})
	})

	Describe("When scoreTTL is set", func() {
		scoreTTL := "100"

		It("Should IncrementMember filling expire ordered set", func() {
			mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(member),
			).Return(databaseMembersReturned, nil)

			mock.EXPECT().SetMembersTTL(gomock.Any(), gomock.Eq(leaderboard), gomock.Any()).Times(1).Return(nil)

			member, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(member.ExpireAt).To(BeNumerically("~", time.Now().Add(100*time.Second).Unix(), 100))
		})
	})

	Describe("When scoreTTL is invalid", func() {
		scoreTTL := "invalid"

		It("Should return error", func() {
			mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(member),
			).Return(databaseMembersReturned, nil)

			_, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
			Expect(err).To(MatchError(service.NewGeneralError("increment member score", "strconv.ParseInt: parsing \"invalid\": invalid syntax")))

		})
	})

	It("Should return error if database SetMembers return in error", func() {
		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(fmt.Errorf("New database error"))

		_, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("increment member score", "New database error")))
	})

	It("Should return error if database GetMembers return in error", func() {
		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(nil, fmt.Errorf("New database error"))

		_, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("increment member score", "New database error")))

	})

	It("Should not call database GetLeaderboardExpiration and SetLeaderboardExpiration if leaderboard isn't formatted to have an expiration", func() {
		expectedMember := &model.Member{
			PublicID:     "member1",
			Score:        2,
			PreviousRank: 0,
			Rank:         2,
		}

		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(float64(score))).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)

		member, err := svc.IncrementMemberScore(context.Background(), leaderboard, member, score, scoreTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(member).To(Equal(expectedMember))
	})

	It("Should set leaderboard expiration if GetLeaderboardExpiration return TTLNotFoundError", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(member), gomock.Eq(float64(score))).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(nil)

		_, err = svc.IncrementMemberScore(context.Background(), leaderboardExpiration, member, score, scoreTTL)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error LeaderboardExpiredError if leaderboard key is formatted and have an expired time", func() {
		leaderboardExpiration := fmt.Sprintf(
			"testkey-from%dto%d",
			time.Now().UTC().Add(time.Duration(-2)*time.Second).Unix(),
			time.Now().UTC().Add(time.Duration(-1)*time.Second).Unix(),
		)
		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(member), gomock.Eq(float64(score))).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)

		_, err := svc.IncrementMemberScore(context.Background(), leaderboardExpiration, member, score, scoreTTL)
		Expect(err).To(MatchError(service.NewLeaderboardExpiredError(leaderboardExpiration)))
	})

	It("Should return error if database GetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())

		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(member), gomock.Eq(float64(score))).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), fmt.Errorf("New database error"))

		_, err := svc.IncrementMemberScore(context.Background(), leaderboardExpiration, member, score, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("increment member score", "New database error")))
	})

	It("Should return error if database SetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		mock.EXPECT().IncrementMemberScore(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(member), gomock.Eq(float64(score))).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(member),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(fmt.Errorf("New database error"))

		_, err = svc.IncrementMemberScore(context.Background(), leaderboardExpiration, member, score, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("increment member score", "New database error")))
	})

})
