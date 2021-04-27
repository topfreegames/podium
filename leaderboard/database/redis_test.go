package database_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/database/redis"
)

var _ = Describe("Redis Database", func() {
	var ctrl *gomock.Controller
	var mock *redis.MockRedis
	var redisDatabase database.Database
	var leaderboard string = "leaderboardTest"
	var leaderboardTTL string = "leaderboardTest:ttl"
	var member string = "memberTest"

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = redis.NewMockRedis(ctrl)

		redisDatabase = &database.Redis{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetMembers", func() {
		var members = []string{"member1", "member2"}
		Describe("When order is asc", func() {
			var order = "asc"

			Describe("When includeTTL is true", func() {
				var includeTTL = true

				It("Should return member list if redis return ok", func() {
					expectedMembers := []*database.Member{
						&database.Member{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    float64(10000),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    float64(10000),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(10000),
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
						&database.Member{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    float64(0),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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

				It("Should return member list with TTL equals zero if redis return ok", func() {
					expectedMembers := []*database.Member{
						&database.Member{
							Member: "member1",
							Score:  float64(1),
							Rank:   int64(0),
							TTL:    float64(0),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member1",
							Score:  float64(2),
							Rank:   int64(0),
							TTL:    float64(10000),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(1),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(10000),
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

				It("Should return member list if redis return ok", func() {
					expectedMembers := []*database.Member{
						&database.Member{
							Member: "member1",
							Score:  float64(2),
							Rank:   int64(0),
							TTL:    float64(0),
						},
						&database.Member{
							Member: "member2",
							Score:  float64(1),
							Rank:   int64(1),
							TTL:    float64(0),
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
						&database.Member{
							Member: "member2",
							Score:  float64(2),
							Rank:   int64(1),
							TTL:    float64(0),
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
					&redis.Member{
						Member: "member1",
						Score:  float64(1),
					},
					&redis.Member{
						Member: "member2",
						Score:  float64(2),
					},
					&redis.Member{
						Member: "member3",
						Score:  float64(3),
					},
				}

				membersToReturn := []*database.Member{
					&database.Member{
						Member: "member1",
						Score:  float64(1),
					},
					&database.Member{
						Member: "member2",
						Score:  float64(2),
					},
					&database.Member{
						Member: "member3",
						Score:  float64(3),
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
					&redis.Member{
						Member: "member3",
						Score:  float64(3),
					},
					&redis.Member{
						Member: "member2",
						Score:  float64(2),
					},
					&redis.Member{
						Member: "member1",
						Score:  float64(1),
					},
				}

				membersToReturn := []*database.Member{
					&database.Member{
						Member: "member3",
						Score:  float64(3),
					},
					&database.Member{
						Member: "member2",
						Score:  float64(2),
					},
					&database.Member{
						Member: "member1",
						Score:  float64(1),
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
		It("Should return nil if no error occur", func() {
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
})
