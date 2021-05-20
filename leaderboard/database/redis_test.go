package database_test

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
)

var _ = Describe("Redis Database", func() {
	var ctrl *gomock.Controller
	var mock *redis.MockRedis
	var redisDatabase database.Database
	var leaderboard string = "leaderboardTest"
	var leaderboardTTL string = "leaderboardTest:ttl"
	var member string = "memberTest"
	var score float64 = 1.0

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = redis.NewMockRedis(ctrl)

		redisDatabase = &database.Redis{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetLeaderboardExpiration", func() {
		It("Should return leaderboard expiration time if all is OK", func() {
			expiration, err := time.ParseDuration("10h")
			Expect(err).NotTo(HaveOccurred())

			mock.EXPECT().TTL(gomock.Any(), gomock.Eq(leaderboard)).Return(expiration, nil)

			ttl, err := redisDatabase.GetLeaderboardExpiration(context.Background(), leaderboard)
			Expect(err).NotTo(HaveOccurred())

			Expect(ttl).To(Equal(int64(expiration)))
		})

		It("Should return TTLNotFoundError if redis redis return TTLNotFoundError", func() {
			mock.EXPECT().TTL(gomock.Any(), gomock.Eq(leaderboard)).Return(time.Duration(-1), redis.NewTTLNotFoundError(leaderboard))

			_, err := redisDatabase.GetLeaderboardExpiration(context.Background(), leaderboard)
			Expect(err).To(Equal(database.NewTTLNotFoundError(leaderboard)))

		})

		It("Should return GeneralError if redis redis return any other error", func() {
			mock.EXPECT().TTL(gomock.Any(), gomock.Eq(leaderboard)).Return(time.Duration(-1), fmt.Errorf("redis error"))

			_, err := redisDatabase.GetLeaderboardExpiration(context.Background(), leaderboard)
			Expect(err).To(Equal(database.NewGeneralError("redis error")))

		})
	})

	Describe("GetMembers", func() {
		var members = []string{"member1", "member2"}
		Describe("When order is asc", func() {
			var order = "asc"

			Describe("When includeTTL is true", func() {
				var includeTTL = true

				It("Should return member list if redis return ok", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    time.Unix(10000, 0),
						},
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(1), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member1")).Return(float64(10000), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member2")).Return(float64(0), redis.NewMemberNotFoundError(leaderboard, "member2"))

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return member list with TTL equals zero if redis return ok", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    time.Unix(10000, 0),
						},
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(1), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member1")).Return(float64(10000), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member2")).Return(float64(0), redis.NewMemberNotFoundError(leaderboard, "member2"))

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return nil member if it doesnt exists", func() {
					expectedMembers := []*database.Member{
						nil,
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Unix(10000, 0),
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), redis.NewMemberNotFoundError(leaderboard, "member1"))
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member2")).Return(float64(10000), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))

				})

				It("Should return General Error if redis return in error", func() {
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), fmt.Errorf("General error"))

					_, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).To(Equal(database.NewGeneralError("General error")))
				})
			})

			Describe("When includeTTL is false", func() {
				var includeTTL = false

				It("Should return member list if redis return ok with TTL = 0", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    time.Time{},
						},
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(1), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return member list with empty TTL if redis return ok", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    time.Time{},
						},
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(1), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return nil member if it doesnt exists", func() {
					expectedMembers := []*database.Member{
						nil,
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), redis.NewMemberNotFoundError(leaderboard, "member1"))
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))

				})

				It("Should return General Error if redis return in error", func() {
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), fmt.Errorf("General error"))

					_, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).To(Equal(database.NewGeneralError("General error")))
				})
			})
		})

		Describe("When order is desc", func() {
			var order = "desc"

			Describe("When includeTTL is true", func() {
				var includeTTL = true

				It("Should return member list if redis return ok", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(2),
							Rank:   int64(0),
							TTL:    time.Unix(10000, 0),
						},
						{
							Member: "member2",
							Score:  float64(1),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(2), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(1), nil)

					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member1")).Return(float64(10000), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member2")).Return(float64(0), redis.NewMemberNotFoundError(leaderboard, "member2"))

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return nil member if it doesnt exists", func() {
					expectedMembers := []*database.Member{
						nil,
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Unix(10000, 0),
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), redis.NewMemberNotFoundError(leaderboard, "member1"))
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq("member2")).Return(float64(10000), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))

				})

				It("Should return General Error if redis return in error", func() {
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), fmt.Errorf("General error"))

					_, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).To(Equal(database.NewGeneralError("General error")))
				})
			})

			Describe("When includeTTL is false", func() {
				var includeTTL = false

				It("Should return member list if redis return ok with TTL equal zero", func() {
					expectedMembers := []*database.Member{
						{
							Member: "member1",
							Score:  float64(2),
							Rank:   int64(0),
							TTL:    time.Time{},
						},
						{
							Member: "member2",
							Score:  float64(1),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(2), nil)
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(1), nil)

					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(int64(0), nil)
					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))
				})

				It("Should return nil member if it doesnt exists", func() {
					expectedMembers := []*database.Member{
						nil,
						{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    time.Time{},
						},
					}

					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), redis.NewMemberNotFoundError(leaderboard, "member1"))
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(float64(2), nil)

					mock.EXPECT().ZRevRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member2")).Return(int64(1), nil)

					members, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).NotTo(HaveOccurred())

					Expect(members).To(Equal(expectedMembers))

				})

				It("Should return General Error if redis return in error", func() {
					mock.EXPECT().ZScore(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq("member1")).Return(float64(-1), fmt.Errorf("General error"))

					_, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
					Expect(err).To(Equal(database.NewGeneralError("General error")))
				})
			})
		})

		Describe("When order is neither asc or desc", func() {
			var order = "invalid"

			It("Should return error InvalidOrder", func() {
				var includeTTL = true

				_, err := redisDatabase.GetMembers(context.Background(), leaderboard, order, includeTTL, members...)
				Expect(err).To(Equal(database.NewInvalidOrderError(order)))
			})
		})
	})

	Describe("GetMemberIDsWithScoreInsideRange", func() {
		var min string = "-inf"
		var max string = "10"
		var offset int = 0
		var count int = 3
		It("Should return member list if redis return ok", func() {
			membersRedisReturn := []string{"member1", "member2", "member3"}

			mock.EXPECT().ZRevRangeByScore(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq(min),
				gomock.Eq(max),
				gomock.Eq(int64(offset)),
				gomock.Eq(int64(count)),
			).Return(membersRedisReturn, nil)

			members, err := redisDatabase.GetMemberIDsWithScoreInsideRange(context.Background(), leaderboard, min, max, offset, count)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(Equal(membersRedisReturn))
		})

		It("Should return General Error if redis return in error", func() {
			mock.EXPECT().ZRevRangeByScore(
				gomock.Any(),
				gomock.Eq(leaderboard),
				gomock.Eq(min),
				gomock.Eq(max),
				gomock.Eq(int64(offset)),
				gomock.Eq(int64(count)),
			).Return(nil, fmt.Errorf("General error"))

			_, err := redisDatabase.GetMemberIDsWithScoreInsideRange(context.Background(), leaderboard, min, max, offset, count)
			Expect(err).To(Equal(database.NewGeneralError("General error")))
		})
	})

	Describe("GetOrderedMembers", func() {
		var start int = 0
		var stop int = 10
		Describe("When order is asc", func() {
			var order = "asc"
			It("Should return member list if redis return ok", func() {
				membersRedisReturn := []*redis.Member{
					{
						Member: "member1",
						Score:  float64(1),
					},
					{
						Member: "member2",
						Score:  float64(2),
					},
					{
						Member: "member3",
						Score:  float64(3),
					},
				}

				membersToReturn := []*database.Member{
					{
						Member: "member1",
						Score:  float64(1),
						Rank:   int64(start),
					},
					{
						Member: "member2",
						Score:  float64(2),
						Rank:   int64(start + 1),
					},
					{
						Member: "member3",
						Score:  float64(3),
						Rank:   int64(start + 2),
					},
				}

				mock.EXPECT().ZRange(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(int64(start)), gomock.Eq(int64(stop))).Return(membersRedisReturn, nil)

				members, err := redisDatabase.GetOrderedMembers(context.Background(), leaderboard, start, stop, order)
				Expect(err).NotTo(HaveOccurred())

				Expect(members).To(Equal(membersToReturn))
			})

			It("Should return General Error if redis return in error", func() {
				mock.EXPECT().ZRange(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(int64(start)), gomock.Eq(int64(stop))).Return(nil, fmt.Errorf("General error"))

				_, err := redisDatabase.GetOrderedMembers(context.Background(), leaderboard, start, stop, order)
				Expect(err).To(Equal(database.NewGeneralError("General error")))
			})
		})

		Describe("When order is desc", func() {
			var order = "desc"
			It("Should return member list if redis return ok", func() {
				membersRedisReturn := []*redis.Member{
					{
						Member: "member3",
						Score:  float64(3),
					},
					{
						Member: "member2",
						Score:  float64(2),
					},
					{
						Member: "member1",
						Score:  float64(1),
					},
				}

				membersToReturn := []*database.Member{
					{
						Member: "member3",
						Score:  float64(3),
						Rank:   int64(start + 0),
					},
					{
						Member: "member2",
						Score:  float64(2),
						Rank:   int64(start + 1),
					},
					{
						Member: "member1",
						Score:  float64(1),
						Rank:   int64(start + 2),
					},
				}

				mock.EXPECT().ZRevRange(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(int64(start)), gomock.Eq(int64(stop))).Return(membersRedisReturn, nil)

				members, err := redisDatabase.GetOrderedMembers(context.Background(), leaderboard, start, stop, order)
				Expect(err).NotTo(HaveOccurred())

				Expect(members).To(Equal(membersToReturn))
			})

			It("Should return General Error if redis return in error", func() {
				mock.EXPECT().ZRevRange(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(int64(start)), gomock.Eq(int64(stop))).Return(nil, fmt.Errorf("General error"))

				_, err := redisDatabase.GetOrderedMembers(context.Background(), leaderboard, start, stop, order)
				Expect(err).To(Equal(database.NewGeneralError("General error")))
			})
		})

		Describe("When order is neither asc or desc", func() {
			var order = "invalid"
			It("Should return error InvalidOrder", func() {
				_, err := redisDatabase.GetOrderedMembers(context.Background(), leaderboard, start, stop, order)
				Expect(err).To(Equal(database.NewInvalidOrderError(order)))
			})
		})
	})

	Describe("GetRank", func() {
		var rank int = 7

		Describe("When order is asc", func() {
			var order string = "asc"

			It("Should return rank of a member in leaderboard if redis return OK", func() {
				mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member)).Return(int64(rank), nil)

				memberRank, err := redisDatabase.GetRank(context.Background(), leaderboard, member, order)
				Expect(err).NotTo(HaveOccurred())

				Expect(memberRank).To(Equal(rank))
			})

			It("Should return error MemberNotFound if redis returns in error", func() {
				mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member)).Return(int64(-1), redis.NewMemberNotFoundError(leaderboard, member))

				_, err := redisDatabase.GetRank(context.Background(), leaderboard, member, order)
				Expect(err).To(Equal(database.NewMemberNotFoundError(leaderboard, member)))
			})

			It("Should return error if redis returns in error", func() {
				mock.EXPECT().ZRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member)).Return(int64(-1), fmt.Errorf("General error"))

				_, err := redisDatabase.GetRank(context.Background(), leaderboard, member, order)
				Expect(err).To(Equal(database.NewGeneralError("General error")))
			})
		})
	})

	Describe("GetTotalMembers", func() {
		var countMembers int = 10

		It("Should return total members of a leaderboard if redis return OK", func() {
			mock.EXPECT().ZCard(gomock.Any(), gomock.Eq(leaderboard)).Return(int64(countMembers), nil)

			totalMembers, err := redisDatabase.GetTotalMembers(context.Background(), leaderboard)
			Expect(err).NotTo(HaveOccurred())

			Expect(totalMembers).To(Equal(countMembers))
		})

		It("Should return total members as zero if redis return KeyNotFoundError", func() {
			mock.EXPECT().ZCard(gomock.Any(), gomock.Eq(leaderboard)).Return(int64(-1), redis.NewKeyNotFoundError(leaderboard))

			totalMembers, err := redisDatabase.GetTotalMembers(context.Background(), leaderboard)
			Expect(err).NotTo(HaveOccurred())

			Expect(totalMembers).To(Equal(0))
		})

		It("Should return error if redis returns in error", func() {
			mock.EXPECT().ZCard(gomock.Any(), gomock.Eq(leaderboard)).Return(int64(-1), fmt.Errorf("General error"))

			_, err := redisDatabase.GetTotalMembers(context.Background(), leaderboard)
			Expect(err).To(Equal(database.NewGeneralError("General error")))
		})
	})

	Describe("Healthcheck", func() {
		It("Should return nil if no error occur", func() {
			mock.EXPECT().Ping(gomock.Any()).Return("PONG", nil)

			err := redisDatabase.Healthcheck(context.Background())
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return error if an error happened", func() {
			mock.EXPECT().Ping(gomock.Any()).Return("", redis.NewGeneralError("New redis error"))

			err := redisDatabase.Healthcheck(context.Background())
			Expect(err).To(Equal(database.NewGeneralError(redis.NewGeneralError("New redis error").Error())))
		})
	})

	Describe("RemoveMembers", func() {
		It("Should return nil if no error occur", func() {
			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq("member2")).Return(nil)

			err := redisDatabase.RemoveMembers(context.Background(), leaderboard, member, "member2")
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return error if an error happened", func() {
			mock.EXPECT().ZRem(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq("member2")).Return(redis.NewGeneralError("New redis error"))

			err := redisDatabase.RemoveMembers(context.Background(), leaderboard, member, "member2")
			Expect(err).To(Equal(database.NewGeneralError(redis.NewGeneralError("New redis error").Error())))
		})
	})

	Describe("RemoveLeaderboard", func() {
		It("Should return nil if no error happended", func() {
			mock.EXPECT().Del(gomock.Any(), gomock.Eq(leaderboard)).Return(nil)

			err := redisDatabase.RemoveLeaderboard(context.Background(), leaderboard)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return error if an error happened", func() {
			mock.EXPECT().Del(gomock.Any(), gomock.Eq(leaderboard)).Return(redis.NewGeneralError("New redis error"))

			err := redisDatabase.RemoveLeaderboard(context.Background(), leaderboard)
			Expect(err).To(Equal(database.NewGeneralError(redis.NewGeneralError("New redis error").Error())))
		})
	})

	Describe("SetLeaderboardExpiration", func() {
		It("Should return nil if all is ok", func() {
			expireTime := time.Unix(123456, 0)
			mock.EXPECT().ExpireAt(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(expireTime)).Return(nil)

			err := redisDatabase.SetLeaderboardExpiration(context.Background(), leaderboard, expireTime)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return GeneralError if redis return in error", func() {
			expireTime := time.Unix(123456, 0)
			mock.EXPECT().ExpireAt(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(expireTime)).Return(fmt.Errorf("New redis error"))

			err := redisDatabase.SetLeaderboardExpiration(context.Background(), leaderboard, expireTime)
			Expect(err).To(Equal(database.NewGeneralError("New redis error")))
		})
	})

	Describe("SetMembersScore", func() {
		redisMembers := []*redis.Member{
			{
				Member: member,
				Score:  score,
			},
			{
				Member: "member2",
				Score:  2.0,
			},
		}

		databaseMembers := []*database.Member{
			{
				Member: member,
				Score:  score,
			},
			{
				Member: "member2",
				Score:  2.0,
			},
		}
		It("Should return nil if all is ok", func() {
			mock.EXPECT().ZAdd(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(redisMembers[0]), gomock.Eq(redisMembers[1])).Return(nil)

			err := redisDatabase.SetMembers(context.Background(), leaderboard, databaseMembers)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return GeneralError if redis return in error", func() {
			mock.EXPECT().ZAdd(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(redisMembers[0]), gomock.Eq(redisMembers[1])).Return(fmt.Errorf("New redis error"))

			err := redisDatabase.SetMembers(context.Background(), leaderboard, databaseMembers)
			Expect(err).To(Equal(database.NewGeneralError("New redis error")))
		})
	})

	Describe("SetMembersTTL", func() {
		time1 := time.Now().Add(-2 * time.Hour)
		time2 := time.Now().Add(-12 * time.Hour)
		leaderboardTTL := fmt.Sprintf("%s:ttl", leaderboard)
		redisMembers := []*redis.Member{
			{
				Member: member,
				Score:  float64(time1.Unix()),
			},
			{
				Member: "member2",
				Score:  float64(time2.Unix()),
			},
		}

		databaseMembers := []*database.Member{
			{
				Member: member,
				TTL:    time1,
			},
			{
				Member: "member2",
				TTL:    time2,
			},
		}
		It("Should return nil if all is ok", func() {
			mock.EXPECT().ZAdd(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq(redisMembers[0]), gomock.Eq(redisMembers[1])).Return(nil)
			mock.EXPECT().SAdd(gomock.Any(), gomock.Eq(database.ExpirationSet), gomock.Eq(leaderboardTTL)).Return(nil)

			err := redisDatabase.SetMembersTTL(context.Background(), leaderboard, databaseMembers)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return GeneralError if redis ZAdd return in error", func() {
			mock.EXPECT().ZAdd(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq(redisMembers[0]), gomock.Eq(redisMembers[1])).Return(fmt.Errorf("New redis error"))

			err := redisDatabase.SetMembersTTL(context.Background(), leaderboard, databaseMembers)
			Expect(err).To(Equal(database.NewGeneralError("New redis error")))
		})

		It("Should return GeneralError if redis SAdd return in error", func() {
			mock.EXPECT().ZAdd(gomock.Any(), gomock.Eq(leaderboardTTL), gomock.Eq(redisMembers[0]), gomock.Eq(redisMembers[1])).Return(nil)
			mock.EXPECT().SAdd(gomock.Any(), gomock.Eq(database.ExpirationSet), gomock.Eq(leaderboardTTL)).Return(fmt.Errorf("New redis error"))

			err := redisDatabase.SetMembersTTL(context.Background(), leaderboard, databaseMembers)
			Expect(err).To(Equal(database.NewGeneralError("New redis error")))
		})
	})
})
