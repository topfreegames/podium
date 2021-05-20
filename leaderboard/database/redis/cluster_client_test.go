package redis_test

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
	"github.com/topfreegames/podium/leaderboard/v2/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Cluster Client", func() {
	const testKey string = "testKey"
	const member string = "member"

	var clusterClient redis.Client
	var goRedis *goredis.ClusterClient

	BeforeEach(func() {
		config, err := testing.GetDefaultConfig("../../../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())

		clusterClient = redis.NewClusterClient(redis.ClusterOptions{
			Addrs:    config.GetStringSlice("redis.addrs"),
			Password: config.GetString("redis.password"),
		})

		goRedis = goredis.NewClusterClient(&goredis.ClusterOptions{
			Addrs:    config.GetStringSlice("redis.addrs"),
			Password: config.GetString("redis.password"),
		})
	})

	AfterEach(func() {
		err := goRedis.Del(context.Background(), testKey).Err()
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Del", func() {
		It("Should return nil if key is removed", func() {
			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: 1.0}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.Del(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())

			keys, err := goRedis.Keys(context.Background(), testKey).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(keys).To(BeEmpty())
		})

		It("Should return nil if set doesnt exists", func() {
			err := clusterClient.Del(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Exists", func() {
		It("Should return nil if key exists", func() {
			err := goRedis.Set(context.Background(), testKey, "testValue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.Exists(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Should return KeyNotFoundError if key doesn't exists", func() {
			err := clusterClient.Exists(context.Background(), testKey)
			Expect(err).To(MatchError(redis.NewKeyNotFoundError(testKey)))
		})
	})

	Describe("ExpireAt", func() {
		It("Should return nil if timeout is set", func() {
			expirationTime := time.Now().Add(10 * time.Minute)

			err := goRedis.Set(context.Background(), testKey, "testValue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.ExpireAt(context.Background(), testKey, expirationTime)
			Expect(err).NotTo(HaveOccurred())

			ttl, err := goRedis.TTL(context.Background(), testKey).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(ttl).NotTo(Equal(redis.TTLKeyNotFound))
			Expect(ttl).NotTo(Equal(redis.KeyWithoutTTL))

			Expect(ttl).Should(BeNumerically("~", 10*time.Minute, time.Minute))
		})

		It("Should return KeyNotFound if key doesn't exists", func() {
			expirationTime := time.Now().Add(10 * time.Minute)

			err := clusterClient.ExpireAt(context.Background(), testKey, expirationTime)
			Expect(err).To(Equal(redis.NewKeyNotFoundError(testKey)))
		})
	})

	Describe("Ping", func() {
		It("Should return nil if redis is OK", func() {
			result, err := clusterClient.Ping(context.Background())
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(Equal("PONG"))
		})
	})

	Describe("SAdd", func() {
		It("Should return nil if member is add to set", func() {
			err := clusterClient.SAdd(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())

			isMember, err := goRedis.SIsMember(context.Background(), testKey, member).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(isMember).To(Equal(true))
		})
	})

	Describe("SMembers", func() {
		It("Should return all members in a set", func() {
			member2 := "member2"
			err := goRedis.SAdd(context.Background(), testKey, member).Err()
			Expect(err).NotTo(HaveOccurred())

			err = goRedis.SAdd(context.Background(), testKey, member2).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := clusterClient.SMembers(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())

			Expect(result).To(ContainElements(member, member2))
		})
	})

	Describe("SRem", func() {
		It("Should return nil if member is removed from set", func() {
			err := goRedis.SAdd(context.Background(), testKey, member).Err()
			Expect(err).NotTo(HaveOccurred())

			err = goRedis.SAdd(context.Background(), testKey, "member2").Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.SRem(context.Background(), testKey, member, "member2")
			Expect(err).NotTo(HaveOccurred())

			isMember, err := goRedis.SIsMember(context.Background(), testKey, member).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(isMember).To(Equal(false))
		})

		It("Should return nil if set doesnt exists", func() {
			err := clusterClient.SRem(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("TTL", func() {
		It("Should return time.Duration if key has TTL set", func() {
			err := goRedis.Set(context.Background(), testKey, "testValue", 10*time.Minute).Err()
			Expect(err).NotTo(HaveOccurred())

			ttl, err := clusterClient.TTL(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())

			Expect(ttl).NotTo(Equal(redis.TTLKeyNotFound))
			Expect(ttl).NotTo(Equal(redis.KeyWithoutTTL))

			Expect(ttl).Should(BeNumerically("~", 10*time.Minute, time.Minute))
		})

		It("Should return KeyNotFound if key doesn't exists", func() {
			_, err := clusterClient.TTL(context.Background(), testKey)
			Expect(err).To(Equal(redis.NewKeyNotFoundError(testKey)))
		})

		It("Should return TTLNotFound if ttl was not set", func() {
			err := goRedis.Set(context.Background(), testKey, "testValue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.TTL(context.Background(), testKey)
			Expect(err).To(Equal(redis.NewTTLNotFoundError(testKey)))
		})
	})

	Describe("ZAdd", func() {
		It("Should return nil if member is add to set", func() {
			score := 1.0
			members := []*redis.Member{
				{
					Member: member,
					Score:  score,
				},
				{
					Member: "member2",
					Score:  2.0,
				},
			}
			err := clusterClient.ZAdd(context.Background(), testKey, members...)
			Expect(err).NotTo(HaveOccurred())

			returnedScore, err := goRedis.ZScore(context.Background(), testKey, member).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedScore).To(Equal(score))
		})
	})

	Describe("ZCard", func() {
		It("Should return nil if member is add to set", func() {
			member2 := "member2"

			score := 1.0
			score2 := 2.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: member2, Score: score2}).Err()
			Expect(err).NotTo(HaveOccurred())

			count, err := clusterClient.ZCard(context.Background(), testKey)
			Expect(err).NotTo(HaveOccurred())

			Expect(count).To(BeEquivalentTo(2))
		})
	})

	Describe("ZIncrBy", func() {
		It("Should return nil if member is updated", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.ZIncrBy(context.Background(), testKey, member, score)
			Expect(err).NotTo(HaveOccurred())

			returnedScore, err := goRedis.ZScore(context.Background(), testKey, member).Result()
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedScore).To(Equal(score + score))
		})
	})

	Describe("ZRange", func() {
		It("Should return members ordered by score, with respective scores", func() {
			member2 := "member2"

			score := 1.0
			score2 := 2.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: member2, Score: score2}).Err()
			Expect(err).NotTo(HaveOccurred())

			members, err := clusterClient.ZRange(context.Background(), testKey, 0, -1)
			Expect(err).NotTo(HaveOccurred())

			Expect(members[0].Member).To(Equal(member))
			Expect(members[0].Score).To(Equal(score))

			Expect(members[1].Member).To(Equal(member2))
			Expect(members[1].Score).To(Equal(score2))
		})
	})

	Describe("ZRangeByScore", func() {
		It("Should return members closest members ordered by score", func() {
			member2 := "member2"

			score := 1.0
			score2 := 2.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: member2, Score: score2}).Err()
			Expect(err).NotTo(HaveOccurred())

			members, err := clusterClient.ZRangeByScore(context.Background(), testKey, "-inf", "1", 0, 1)
			Expect(err).NotTo(HaveOccurred())

			Expect(members[0]).To(Equal(member))
		})
	})

	Describe("ZRank", func() {
		It("Should return member rank and nil if no error ocurr", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			rank, err := clusterClient.ZRank(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())

			Expect(rank).To(BeEquivalentTo(0))
		})

		It("Should return error MemberNotFounderror if sorted set is empty", func() {
			_, err := clusterClient.ZRank(context.Background(), testKey, member)
			Expect(err).To(Equal(redis.NewMemberNotFoundError(testKey, member)))
		})

		It("Should return error MemberNotFounderror if sorted set doesn't have member", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.ZRank(context.Background(), testKey, "member not found")
			Expect(err).To(Equal(redis.NewMemberNotFoundError(testKey, "member not found")))
		})
	})

	Describe("ZRem", func() {
		It("Should return nil if member is removed from set", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.ZRem(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())

			_, err = goRedis.ZRank(context.Background(), testKey, member).Result()
			Expect(err).To(HaveOccurred())
		})

		It("Should return nil if multiple members is removed from set", func() {
			score := 1.0
			secondMember := "member2"

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: secondMember, Score: score * 2.0}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = clusterClient.ZRem(context.Background(), testKey, member, secondMember)
			Expect(err).NotTo(HaveOccurred())

			_, err = goRedis.ZRank(context.Background(), testKey, member).Result()
			Expect(err).To(HaveOccurred())

			_, err = goRedis.ZRank(context.Background(), testKey, secondMember).Result()
			Expect(err).To(HaveOccurred())
		})

		It("Should return nil if set doesn't exists", func() {
			err := clusterClient.ZRem(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ZRevRange", func() {
		It("Should return members ordered by score, with respective scores", func() {
			member2 := "member2"

			score := 1.0
			score2 := 2.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: member2, Score: score2}).Err()
			Expect(err).NotTo(HaveOccurred())

			members, err := clusterClient.ZRevRange(context.Background(), testKey, 0, -1)
			Expect(err).NotTo(HaveOccurred())

			Expect(members[0].Member).To(Equal(member2))
			Expect(members[0].Score).To(Equal(score2))

			Expect(members[1].Member).To(Equal(member))
			Expect(members[1].Score).To(Equal(score))
		})
	})

	Describe("ZRevRangeByScore", func() {
		It("Should return members closest members ordered by score", func() {
			member2 := "member2"

			score := 1.0
			score2 := 2.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}, &goredis.Z{Member: member2, Score: score2}).Err()
			Expect(err).NotTo(HaveOccurred())

			members, err := clusterClient.ZRevRangeByScore(context.Background(), testKey, "-inf", "1", 0, 1)
			Expect(err).NotTo(HaveOccurred())

			Expect(members[0]).To(Equal(member))
		})
	})

	Describe("ZRevRank", func() {
		It("Should return rank position if member is in set", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: "another-member", Score: score * 2.0}).Err()
			Expect(err).NotTo(HaveOccurred())

			returnedRank, err := clusterClient.ZRevRank(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedRank).To(BeEquivalentTo(1))
		})

		It("Should return MemberNotFound if key doesn't have member", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			err = goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: "another-member", Score: score * 2.0}).Err()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.ZRevRank(context.Background(), testKey, "wrongKey")
			Expect(err).To(Equal(redis.NewMemberNotFoundError(testKey, "wrongKey")))
		})
	})

	Describe("ZScore", func() {
		It("Should return score if member is in set", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			returnedScore, err := clusterClient.ZScore(context.Background(), testKey, member)
			Expect(err).NotTo(HaveOccurred())

			Expect(returnedScore).To(Equal(score))
		})

		It("Should return MemberNotFound if key doesn't have member", func() {
			score := 1.0

			err := goRedis.ZAdd(context.Background(), testKey, &goredis.Z{Member: member, Score: score}).Err()
			Expect(err).NotTo(HaveOccurred())

			_, err = clusterClient.ZScore(context.Background(), testKey, "wrongKey")
			Expect(err).To(Equal(redis.NewMemberNotFoundError(testKey, "wrongKey")))
		})
	})
})
