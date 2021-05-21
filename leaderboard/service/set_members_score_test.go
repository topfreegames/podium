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

var _ = Describe("Service SetMembersScore", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboard"
	var previousRank bool = false
	var scoreTTL string = ""

	databaseMembersToInsert := []*database.Member{
		{
			Member: "member1",
			Score:  1.0,
		},
		{
			Member: "member2",
			Score:  2.0,
		},
	}

	databaseMembersToGetRank := []string{
		"member1",
		"member2",
	}

	databaseMembersPreviousRankReturned := []*database.Member{
		{
			Member: "member1",
			Score:  2.0,
			Rank:   int64(0),
		},
		{
			Member: "member2",
			Score:  1.0,
			Rank:   int64(1),
		},
	}

	databaseMembersReturned := []*database.Member{
		{
			Member: "member1",
			Score:  1.0,
			Rank:   int64(1),
		},
		{
			Member: "member2",
			Score:  2.0,
			Rank:   int64(0),
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
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			expectedMembers := []*model.Member{
				{
					PublicID:     "member1",
					Score:        1,
					PreviousRank: 0,
					Rank:         2,
				},
				{
					PublicID:     "member2",
					Score:        2,
					PreviousRank: 0,
					Rank:         1,
				},
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal(expectedMembers))
		})
	})

	Describe("When previousRank is true", func() {
		previousRank := true

		It("Should set Members with previousRank", func() {
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			expectedMembers := []*model.Member{
				{
					PublicID:     "member1",
					Score:        1,
					PreviousRank: 1,
					Rank:         2,
				},
				{
					PublicID:     "member2",
					Score:        2,
					PreviousRank: 2,
					Rank:         1,
				},
			}

			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Times(1).Return(databaseMembersPreviousRankReturned, nil)

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal(expectedMembers))
		})

		It("Should set a non existent member as rank equals to -1", func() {
			databaseMembersPreviousRankReturned := []*database.Member{
				nil,
				{
					Member: "member2",
					Score:  1.0,
					Rank:   int64(1),
				},
			}
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			expectedMembers := []*model.Member{
				{
					PublicID:     "member1",
					Score:        1,
					PreviousRank: -1,
					Rank:         2,
				},
				{
					PublicID:     "member2",
					Score:        2,
					PreviousRank: 2,
					Rank:         1,
				},
			}

			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Times(1).Return(databaseMembersPreviousRankReturned, nil)

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal(expectedMembers))
		})

		It("Should return error if GetMembers return in error", func() {
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Times(1).Return(nil, fmt.Errorf("New database error"))
			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).To(Equal(service.NewGeneralError("set members score", "New database error")))
		})
	})

	Describe("When scoreTTL is empty", func() {
		It("Should SetMembers without filling expire ordered set", func() {
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			expectedMembers := []*model.Member{
				{
					PublicID:     "member1",
					Score:        1,
					PreviousRank: 0,
					Rank:         2,
				},
				{
					PublicID:     "member2",
					Score:        2,
					PreviousRank: 0,
					Rank:         1,
				},
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal(expectedMembers))
		})
	})

	Describe("When scoreTTL is set", func() {
		scoreTTL := "100"

		It("Should SetMembers filling expire ordered set", func() {
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			mock.EXPECT().SetMembersTTL(gomock.Any(), gomock.Eq(leaderboard), gomock.Any()).Times(1).Return(nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).NotTo(HaveOccurred())

			Expect(members[0].ExpireAt).To(BeNumerically("~", time.Now().Add(100*time.Second).Unix(), 100))
			Expect(members[1].ExpireAt).To(BeNumerically("~", time.Now().Add(100*time.Second).Unix(), 100))
		})
	})

	Describe("When scoreTTL is invalid", func() {
		scoreTTL := "invalid"

		It("Should return error", func() {
			members := []*model.Member{
				{
					PublicID: "member1",
					Score:    1,
				},
				{
					PublicID: "member2",
					Score:    2,
				},
			}

			mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
			mock.EXPECT().GetMembers(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq("desc"),
				gomock.Eq(true),
				gomock.Eq(databaseMembersToGetRank[0]),
				gomock.Eq(databaseMembersToGetRank[1]),
			).Return(databaseMembersReturned, nil)

			err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
			Expect(err).To(MatchError(service.NewGeneralError("set members score", "strconv.ParseInt: parsing \"invalid\": invalid syntax")))

		})
	})

	It("Should return error if database SetMembers return in error", func() {
		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(fmt.Errorf("New database error"))

		err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set members score", "New database error")))
	})

	It("Should return error if database GetMembers return in error", func() {
		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(nil, fmt.Errorf("New database error"))

		err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set members score", "New database error")))

	})

	It("Should not call database GetLeaderboardExpiration and SetLeaderboardExpiration if leaderboard isn't formatted to have an expiration", func() {
		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		expectedMembers := []*model.Member{
			{
				PublicID:     "member1",
				Score:        1,
				PreviousRank: 0,
				Rank:         2,
			},
			{
				PublicID:     "member2",
				Score:        2,
				PreviousRank: 0,
				Rank:         1,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboard),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(databaseMembersReturned, nil)

		err := svc.SetMembersScore(context.Background(), leaderboard, members, previousRank, scoreTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(members).To(Equal(expectedMembers))
	})

	It("Should set leaderboard expiration if GetLeaderboardExpiration return TTLNotFoundError", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(nil)

		err = svc.SetMembersScore(context.Background(), leaderboardExpiration, members, previousRank, scoreTTL)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error LeaderboardExpiredError if leaderboard is expired", func() {
		leaderboardExpiration := fmt.Sprintf(
			"testkey-from%dto%d",
			time.Now().UTC().Add(time.Duration(-2)*time.Second).Unix(),
			time.Now().UTC().Add(time.Duration(-1)*time.Second).Unix(),
		)

		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(databaseMembersReturned, nil)

		err := svc.SetMembersScore(context.Background(), leaderboardExpiration, members, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewLeaderboardExpiredError(leaderboardExpiration)))
	})

	It("Should return error if database GetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())

		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), fmt.Errorf("New database error"))

		err := svc.SetMembersScore(context.Background(), leaderboardExpiration, members, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set members score", "New database error")))
	})

	It("Should return error if database SetLeaderboardExpiration return in error", func() {
		leaderboardExpiration := fmt.Sprintf("year%d", time.Now().UTC().Year())
		expireAt, err := expiration.GetExpireAt(leaderboardExpiration)
		Expect(err).NotTo(HaveOccurred())

		members := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
			},
			{
				PublicID: "member2",
				Score:    2,
			},
		}

		mock.EXPECT().SetMembers(gomock.Any(), gomock.Eq(leaderboardExpiration), gomock.Eq(databaseMembersToInsert)).Times(1).Return(nil)
		mock.EXPECT().GetMembers(
			gomock.Any(),
			gomock.Eq(leaderboardExpiration),
			gomock.Eq("desc"),
			gomock.Eq(true),
			gomock.Eq(databaseMembersToGetRank[0]),
			gomock.Eq(databaseMembersToGetRank[1]),
		).Return(databaseMembersReturned, nil)
		mock.EXPECT().GetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration)).Return(int64(-1), database.NewTTLNotFoundError(leaderboard))

		mock.EXPECT().SetLeaderboardExpiration(gomock.Any(), gomock.Eq(leaderboardExpiration), time.Unix(expireAt, 0)).Return(fmt.Errorf("New database error"))

		err = svc.SetMembersScore(context.Background(), leaderboardExpiration, members, previousRank, scoreTTL)
		Expect(err).To(MatchError(service.NewGeneralError("set members score", "New database error")))
	})

})
