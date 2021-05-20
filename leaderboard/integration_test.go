// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package leaderboard_test

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/database/redis"
	"github.com/topfreegames/podium/leaderboard/v2/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"

	uuid "github.com/satori/go.uuid"
)

var _ = Describe("Leaderboard integration tests", func() {

	var redisDatabase *database.Redis
	var leaderboards service.Leaderboard
	var faultyLeaderboards service.Leaderboard
	const testLeaderboardID = "test-leaderboard"

	BeforeEach(func() {
		var err error

		defaultConfig, err := testing.GetDefaultConfig("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())

		redisDatabase, err = GetDefaultRedis()
		Expect(err).NotTo(HaveOccurred())

		leaderboards = service.NewService(redisDatabase)

		faultyLeaderboards = service.NewService(database.NewRedisDatabase(database.RedisOptions{
			ClusterEnabled: defaultConfig.GetBool("faultyRedis.clusterEnabled"),
			Addrs:          defaultConfig.GetStringSlice("faultyRedis.addrs"),
			Host:           defaultConfig.GetString("faultyRedis.host"),
			Port:           defaultConfig.GetInt("faultyRedis.port"),
			Password:       defaultConfig.GetString("faultyRedis.password"),
			DB:             defaultConfig.GetInt("faultyRedis.db"),
		}))

		err = leaderboards.RemoveLeaderboard(NewEmptyCtx(), testLeaderboardID)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterSuite(func() {
		err := leaderboards.RemoveLeaderboard(NewEmptyCtx(), testLeaderboardID)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("setting member scores", func() {
		It("should set scores and return ranks", func() {
			dayvson, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID,
				"dayvson", 481516, false, "")
			Expect(err).NotTo(HaveOccurred())
			arthur, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID,
				"arthur", 1000, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(arthur.Rank).To(Equal(2))
		})

		It("should set score expiration if expiry field is passed", func() {
			ttl := "100"
			_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID,
				"denix", 481516, false, ttl)
			Expect(err).NotTo(HaveOccurred())
			redisLBExpirationKey := fmt.Sprintf("%s:ttl", testLeaderboardID)
			err = redisDatabase.Exists(context.Background(), redisLBExpirationKey)
			Expect(err).NotTo(HaveOccurred())
			redisExpirationSetKey := "expiration-sets"
			err = redisDatabase.Exists(context.Background(), redisExpirationSetKey)
			Expect(err).NotTo(HaveOccurred())
			result2, err := redisDatabase.SMembers(context.Background(), redisExpirationSetKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(ContainElement(redisLBExpirationKey))
			Expect(err).NotTo(HaveOccurred())
			result3, err := redisDatabase.ZScore(context.Background(), redisLBExpirationKey, "denix")
			Expect(err).NotTo(HaveOccurred())
			ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
			Expect(result3).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		})

		It("should set scores and return previous ranks", func() {
			member1, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member1",
				481516, true, "")
			Expect(err).NotTo(HaveOccurred())
			member2, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member2",
				1000, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member1.Rank).To(Equal(1))
			Expect(member1.PreviousRank).To(Equal(-1))
			Expect(member2.Rank).To(Equal(2))
			Expect(member2.PreviousRank).To(Equal(0))
			nmember1, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member1",
				1, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(nmember1.Rank).To(Equal(2))
			Expect(nmember1.PreviousRank).To(Equal(1))
		})

		It("should fail if invalid connection to Redis", func() {
			_, err := faultyLeaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "dayvson",
				481516, false, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("setting members scores via bulk", func() {
		It("should set many members score in one action and return ranks", func() {
			members := []*model.Member{
				{Score: 481516, PublicID: "dayvson"},
				{Score: 1000, PublicID: "arthur"},
			}
			err := leaderboards.SetMembersScore(NewEmptyCtx(), testLeaderboardID, members, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(members[0].PublicID).To(Equal("dayvson"))
			Expect(members[0].Rank).To(Equal(1))
			Expect(members[0].PreviousRank).To(Equal(0))
			Expect(members[1].PublicID).To(Equal("arthur"))
			Expect(members[1].Rank).To(Equal(2))
			Expect(members[1].PreviousRank).To(Equal(0))
		})

		It("should set many score expirations if expiry field is passed", func() {
			ttl := "100"
			members := []*model.Member{
				{Score: 481516, PublicID: "denix1"},
				{Score: 481516, PublicID: "denix2"},
			}
			err := leaderboards.SetMembersScore(NewEmptyCtx(), testLeaderboardID, members, false, ttl)
			Expect(err).NotTo(HaveOccurred())
			redisLBExpirationKey := fmt.Sprintf("%s:ttl", testLeaderboardID)
			err = redisDatabase.Exists(context.Background(), redisLBExpirationKey)
			Expect(err).NotTo(HaveOccurred())
			redisExpirationSetKey := "expiration-sets"
			err = redisDatabase.Exists(context.Background(), redisExpirationSetKey)
			Expect(err).NotTo(HaveOccurred())
			result2, err := redisDatabase.SMembers(context.Background(), redisExpirationSetKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result2).To(ContainElement(redisLBExpirationKey))
			result3, err := redisDatabase.ZScore(context.Background(), redisLBExpirationKey, "denix1")
			Expect(err).NotTo(HaveOccurred())
			ttlInt, _ := strconv.ParseInt(ttl, 10, 64)
			Expect(result3).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
			result4, err := redisDatabase.ZScore(context.Background(), redisLBExpirationKey, "denix2")
			Expect(err).NotTo(HaveOccurred())
			Expect(result4).To(BeNumerically("~", time.Now().Unix()+ttlInt, 1))
		})

		It("should set many scores and return previous ranks", func() {
			members := []*model.Member{
				{Score: 481516, PublicID: "member1"},
				{Score: 1000, PublicID: "member2"},
			}
			err := leaderboards.SetMembersScore(NewEmptyCtx(), testLeaderboardID, members, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(members[0].Rank).To(Equal(1))
			Expect(members[0].PreviousRank).To(Equal(-1))
			Expect(members[1].Rank).To(Equal(2))
			Expect(members[1].PreviousRank).To(Equal(-1))
			members = []*model.Member{
				{Score: 1, PublicID: "member1"},
				{Score: 500, PublicID: "member2"},
			}
			err = leaderboards.SetMembersScore(NewEmptyCtx(), testLeaderboardID, members, true, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(members[0].Rank).To(Equal(2))
			Expect(members[0].PreviousRank).To(Equal(1))
			Expect(members[1].Rank).To(Equal(1))
			Expect(members[1].PreviousRank).To(Equal(2))
		})

		It("should fail if invalid connection to Redis", func() {
			err := faultyLeaderboards.SetMembersScore(NewEmptyCtx(), testLeaderboardID, []*model.Member{}, false, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("increment member scores", func() {
		It("should increment member score and return ranks", func() {
			lbID := uuid.NewV4().String()

			_, err := leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "dayvson", 1000, false, "")
			Expect(err).NotTo(HaveOccurred())

			member, err := leaderboards.IncrementMemberScore(NewEmptyCtx(), lbID, "dayvson", 10, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Score).To(Equal(int64(1010)))
			Expect(member.PublicID).To(Equal("dayvson"))

			score, err := redisDatabase.ZScore(context.Background(), lbID, "dayvson")
			Expect(err).NotTo(HaveOccurred())
			Expect(int(score)).To(Equal(1010))
		})

		It("should increment member score when leaderboard does not exist and return ranks", func() {
			lbID := uuid.NewV4().String()

			member, err := leaderboards.IncrementMemberScore(NewEmptyCtx(), lbID, "dayvson", 10, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(member.Score).To(Equal(int64(10)))
			Expect(member.PublicID).To(Equal("dayvson"))

			score, err := redisDatabase.ZScore(context.Background(), lbID, "dayvson")
			Expect(err).NotTo(HaveOccurred())
			Expect(int(score)).To(Equal(10))
		})

		It("should fail if invalid connection to Redis", func() {
			_, err := faultyLeaderboards.IncrementMemberScore(NewEmptyCtx(), testLeaderboardID, "dayvson", 16, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting number of members", func() {
		It("should retrieve the number of members in a leaderboard", func() {
			for i := 0; i < 10; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			count, err := leaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(10))
		})

		It("should fail if faulty redis client", func() {
			_, err := faultyLeaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("removing members", func() {
		It("should remove member", func() {
			for i := 0; i < 10; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i),
					int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(leaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)).To(Equal(10))
			member := "member_5"
			err := leaderboards.RemoveMember(NewEmptyCtx(), testLeaderboardID, member)
			Expect(err).NotTo(HaveOccurred())
			Expect(leaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)).To(Equal(9))
		})

		It("should remove many members", func() {
			for i := 0; i < 10; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i),
					int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(leaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)).To(Equal(10))
			members := make([]string, 2)
			members[0] = "member_5"
			members[1] = "member_6"
			err := leaderboards.RemoveMembers(NewEmptyCtx(), testLeaderboardID, members)
			Expect(err).NotTo(HaveOccurred())
			Expect(leaderboards.TotalMembers(NewEmptyCtx(), testLeaderboardID)).To(Equal(8))
		})

		It("should fail if faulty redis client", func() {
			members := make([]string, 1)
			members[0] = "invalid member"
			err := faultyLeaderboards.RemoveMembers(NewEmptyCtx(), testLeaderboardID, members)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting number of pages in leaderboard", func() {
		It("should return total number of pages", func() {
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i),
					int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			Expect(leaderboards.TotalPages(NewEmptyCtx(), testLeaderboardID, 25)).To(Equal(5))
		})

		It("should fail if faulty redis client", func() {
			_, err := faultyLeaderboards.TotalPages(NewEmptyCtx(), testLeaderboardID, 10)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting member details for a given leaderboard", func() {
		It("should return member details", func() {
			lbID := uuid.NewV4().String()
			dayvson, err := leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "dayvson", 12345, false, "")
			Expect(err).NotTo(HaveOccurred())
			felipe, err := leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "felipe", 12344, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "felipe", 12346, false, "")
			felipe, err = leaderboards.GetMember(NewEmptyCtx(), lbID, "felipe", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			dayvson, err = leaderboards.GetMember(NewEmptyCtx(), lbID, "dayvson", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
		})

		It("should return member details including score expiration", func() {
			lbID := uuid.NewV4().String()
			dayvson, err := leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "dayvson", 12345, false, "10")
			Expect(err).NotTo(HaveOccurred())
			felipe, err := leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "felipe", 12344, false, "")
			Expect(err).NotTo(HaveOccurred())
			Expect(dayvson.Rank).To(Equal(1))
			Expect(felipe.Rank).To(Equal(2))
			leaderboards.SetMemberScore(NewEmptyCtx(), lbID, "felipe", 12346, false, "")
			felipe, err = leaderboards.GetMember(NewEmptyCtx(), lbID, "felipe", "desc", true)
			Expect(err).NotTo(HaveOccurred())
			dayvson, err = leaderboards.GetMember(NewEmptyCtx(), lbID, "dayvson", "desc", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(felipe.Rank).To(Equal(1))
			Expect(dayvson.Rank).To(Equal(2))
			Expect(felipe.ExpireAt).To(Equal(0))
			Expect(dayvson.ExpireAt).To(BeNumerically("~", time.Now().Unix()+10, 1))
		})

		It("should fail if member does not exist", func() {
			lbID := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			member, err := leaderboards.GetMember(NewEmptyCtx(), lbID, memberID, "desc", false)
			Expect(member).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for member %s in leaderboard %s.", memberID, lbID)),
			)
		})

		It("should fail if member does not exist and should include expiration timestamp", func() {
			lbID := uuid.NewV4().String()
			memberID := uuid.NewV4().String()
			member, err := leaderboards.GetMember(NewEmptyCtx(), lbID, memberID, "desc", true)
			Expect(member).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring(
				fmt.Sprintf("Could not find data for member %s in leaderboard %s.", memberID, lbID)),
			)
		})

		It("should fail if faulty redis client", func() {
			_, err := faultyLeaderboards.GetMember(NewEmptyCtx(), testLeaderboardID, "qwe", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("get members around someone in a leaderboard", func() {
		It("should get members around specific member", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, "member_20", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})

		It("should always return page size members when page size is less than total members", func() {
			pageSize := 3
			for i := 0; i < 5; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			for i := 0; i < 5; i++ {
				members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, fmt.Sprintf("member_%d", i), "desc", false)
				Expect(err).NotTo(HaveOccurred())

				Expect(len(members)).To(Equal(pageSize))
			}

		})

		It("should get members around specific member in reverse order", func() {
			pageSize := 20
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, "member_20", "asc", false)
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_11"))
			Expect(lastAroundMe.PublicID).To(Equal("member_30"))
		})

		It("should get members around specific member if repeated scores", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, "member_20", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(pageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(firstAroundMe.Score).To(Equal(int64(100)))
			Expect(lastAroundMe.Score).To(Equal(int64(100)))
		})

		It("should get PageSize members around specific member even if member in ranking top", func() {
			pageSize := 25
			for i := 1; i <= 100; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, "member_2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(pageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_1"))
			Expect(lastAroundMe.PublicID).To(Equal("member_25"))
		})

		It("should get PageSize members around specific member even if member in ranking bottom", func() {
			pageSize := 25
			for i := 1; i <= 100; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, pageSize, "member_99", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(members)).To(Equal(pageSize))
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(firstAroundMe.PublicID).To(Equal("member_76"))
			Expect(lastAroundMe.PublicID).To(Equal("member_100"))
		})

		It("should get PageSize members when interval larger than total members", func() {
			for i := 1; i <= 10; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, 25, "member_2", "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(10))
			firstAroundMe := members[0]
			lastAroundMe := members[9]
			Expect(firstAroundMe.PublicID).To(Equal("member_1"))
			Expect(lastAroundMe.PublicID).To(Equal("member_10"))
		})

		It("should fail if faulty redis client", func() {
			_, err := faultyLeaderboards.GetAroundMe(NewEmptyCtx(), testLeaderboardID, 10, "qwe", "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("get members around score in a leaderboard", func() {
		It("should get members around specific score", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, pageSize, 1234*20, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_31"))
			Expect(lastAroundMe.PublicID).To(Equal("member_7"))
		})

		It("should always return page size members when page size is less than total members", func() {
			pageSize := 3
			for i := 0; i < 5; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			for i := 0; i < 5; i++ {
				members, err := leaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, pageSize, int64(i), "desc")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(members)).To(Equal(pageSize))
			}

		})

		It("should get members around specific score reverse order", func() {
			pageSize := 20
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, pageSize, 1234*20, "asc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_11"))
			Expect(lastAroundMe.PublicID).To(Equal("member_30"))
		})

		It("should get last members if score <= 0", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, pageSize, -50, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_24"))
			Expect(lastAroundMe.PublicID).To(Equal("member_0"))
		})

		It("should get top members if score > max score in leaderboard", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, pageSize, 1234*200, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.PublicID).To(Equal("member_100"))
			Expect(lastAroundMe.PublicID).To(Equal("member_76"))
		})

		It("should fail if faulty redis client", func() {
			_, err := faultyLeaderboards.GetAroundScore(NewEmptyCtx(), testLeaderboardID, 10, 20, "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting member ranking", func() {
		It("should return specific member ranking", func() {
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_6", 1000, false, "")
			Expect(leaderboards.GetRank(NewEmptyCtx(), testLeaderboardID, "member_6", "desc")).To(Equal(100))
		})

		It("should return specific member ranking if asc order", func() {
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_6", 1000, false, "")
			Expect(leaderboards.GetRank(NewEmptyCtx(), testLeaderboardID, "member_6", "asc")).To(Equal(2))
		})

		It("should fail if member does not exist", func() {
			rank, err := leaderboards.GetRank(NewEmptyCtx(), uuid.NewV4().String(), "invalid-member", "desc")
			Expect(rank).To(Equal(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Could not find data for member invalid-member in leaderboard"))
		})

		It("should fail if invalid redis connection", func() {
			rank, err := faultyLeaderboards.GetRank(NewEmptyCtx(), uuid.NewV4().String(), "invalid-member", "desc")
			Expect(rank).To(Equal(-1))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting leaderboard leaders", func() {
		It("should get specific number of leaders", func() {
			pageSize := 25
			for i := 0; i < 1000; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i+1), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, pageSize, 1, "desc")
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := members[0]
			lastOnPage := members[len(members)-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1000"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_976"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get specific number of leaders in reverse order", func() {
			pageSize := 25
			for i := 0; i < 1000; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i+1), int64(1234*i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, pageSize, 1, "asc")
			Expect(err).NotTo(HaveOccurred())

			firstOnPage := members[0]
			lastOnPage := members[len(members)-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstOnPage.PublicID).To(Equal("member_1"))
			Expect(firstOnPage.Rank).To(Equal(1))
			Expect(lastOnPage.PublicID).To(Equal("member_25"))
			Expect(lastOnPage.Rank).To(Equal(25))
		})

		It("should get leaders if repeated scores", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, pageSize, 1, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.Score).To(Equal(int64(100)))
			Expect(lastAroundMe.Score).To(Equal(int64(100)))
		})

		It("should get leaders for negative pages get page 1", func() {
			pageSize := 25
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, pageSize, -1, "desc")
			Expect(err).NotTo(HaveOccurred())
			firstAroundMe := members[0]
			lastAroundMe := members[pageSize-1]
			Expect(len(members)).To(Equal(pageSize))
			Expect(firstAroundMe.Score).To(Equal(int64(100)))
			Expect(lastAroundMe.Score).To(Equal(int64(100)))
		})

		It("should get empty leaders for pages greater than total pages", func() {
			for i := 0; i < 101; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), testLeaderboardID, "member_"+strconv.Itoa(i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
			}
			members, err := leaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, 25, 99999, "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(0))
		})

		It("should fail if invalid connection to Redis", func() {
			//testLeaderboard := NewClient(getFaultyRedis(), "test-leaderboard", 25)
			_, err := faultyLeaderboards.GetLeaders(NewEmptyCtx(), testLeaderboardID, 25, 1, "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("expiration of leaderboards", func() {
		It("should fail if invalid leaderboard", func() {
			leaderboardID := "leaderboard_from20201039to20201011"
			_, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, "dayvson", 12345, false, "")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("day out of range"))
		})

		It("should add yearly expiration if leaderboard supports it", func() {
			leaderboardID := fmt.Sprintf("test-leaderboard-year%d", time.Now().UTC().Year())
			_, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, "dayvson", 12345, false, "")
			Expect(err).NotTo(HaveOccurred())

			result, err := redisDatabase.TTL(context.Background(), leaderboardID)
			Expect(err).NotTo(HaveOccurred())

			exp := result.Seconds()
			Expect(err).NotTo(HaveOccurred())
			Expect(exp).To(BeNumerically(">", int64(-1)))
		})
	})

	Describe("get top x percent of members in the leaderboard", func() {
		It("should get top 10 percent members in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()
			members := []*model.Member{}
			for i := 0; i < 100; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64((100-i)*100), false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 10, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(int64(10000)))

			Expect(top10[9].PublicID).To(Equal("friend-9"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(int64(9100)))
		})

		It("should not break if order is different from asc and desc, should only default to desc", func() {
			leaderboardID := uuid.NewV4().String()

			members := []*model.Member{}
			for i := 0; i < 100; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64((100-i)*100), false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 10, 2000, "lalala")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(int64(10000)))

			Expect(top10[9].PublicID).To(Equal("friend-9"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(int64(9100)))
		})

		It("should get top 10 percent members in the leaderboard in reverse order", func() {
			leaderboardID := uuid.NewV4().String()

			members := []*model.Member{}
			for i := 0; i < 100; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64((100-i)*100), false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 10, 2000, "asc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].PublicID).To(Equal("friend-99"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(int64(100)))

			Expect(top10[9].PublicID).To(Equal("friend-90"))
			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(int64(1000)))
		})

		It("should get max members if query too broad", func() {
			leaderboardID := uuid.NewV4().String()

			members := []*model.Member{}
			for i := 0; i < 10; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64((100-i)*100), false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top3, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 100, 3, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top3).To(HaveLen(3))

			Expect(top3[0].PublicID).To(Equal("friend-0"))
			Expect(top3[0].Rank).To(Equal(1))
			Expect(top3[0].Score).To(Equal(int64(10000)))

			Expect(top3[2].PublicID).To(Equal("friend-2"))
			Expect(top3[2].Rank).To(Equal(3))
			Expect(top3[2].Score).To(Equal(int64(9800)))
		})

		It("should get top 1 percent return at least 1", func() {
			leaderboardID := uuid.NewV4().String()

			members := []*model.Member{}
			for i := 0; i < 2; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64((100-i)*100), false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 1, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(1))

			Expect(top10[0].PublicID).To(Equal("friend-0"))
			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(int64(10000)))
		})

		It("should get top 10 percent members in the leaderboard if repeated scores", func() {
			leaderboardID := uuid.NewV4().String()

			members := []*model.Member{}
			for i := 0; i < 100; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), 100, false, "")
				Expect(err).NotTo(HaveOccurred())
				members = append(members, member)
			}

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 10, 2000, "desc")
			Expect(err).NotTo(HaveOccurred())

			Expect(top10).To(HaveLen(10))

			Expect(top10[0].Rank).To(Equal(1))
			Expect(top10[0].Score).To(Equal(int64(100)))

			Expect(top10[9].Rank).To(Equal(10))
			Expect(top10[9].Score).To(Equal(int64(100)))
		})

		It("should fail if more than 100 percent", func() {
			leaderboardID := uuid.NewV4().String()

			top10, err := leaderboards.GetTopPercentage(NewEmptyCtx(), leaderboardID, 10, 101, 2000, "desc")
			Expect(top10).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(service.NewPercentageError(101)))
		})

		It("should fail if invalid redis connection", func() {
			members, err := faultyLeaderboards.GetTopPercentage(NewEmptyCtx(), uuid.NewV4().String(), 25, 10, 2000, "desc")
			Expect(members).To(BeEmpty())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("get members by range in a given leaderboard", func() {
		It("should get members in a range in the leaderboard", func() {
			leaderboardID := uuid.NewV4().String()

			expMembers := []*model.Member{}
			for i := 0; i < 100; i++ {
				member, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
				expMembers = append(expMembers, member)
			}

			members, err := leaderboards.GetMembersByRange(NewEmptyCtx(), leaderboardID, 20, 39, "desc")
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(20))

			for i := 20; i < 40; i++ {
				Expect(members[i-20].PublicID).To(Equal(expMembers[i].PublicID))
			}
		})

		It("should fail if invalid connection to Redis", func() {
			leaderboardID := uuid.NewV4().String()
			_, err := faultyLeaderboards.GetMembersByRange(NewEmptyCtx(), leaderboardID, 20, 39, "desc")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("remove leaderboard", func() {
		It("should remove a leaderboard from redis", func() {
			leaderboardID := uuid.NewV4().String()

			for i := 0; i < 10; i++ {
				_, err := leaderboards.SetMemberScore(NewEmptyCtx(), leaderboardID, fmt.Sprintf("friend-%d", i), int64(100-i), false, "")
				Expect(err).NotTo(HaveOccurred())
			}

			err := leaderboards.RemoveLeaderboard(NewEmptyCtx(), leaderboardID)
			Expect(err).NotTo(HaveOccurred())

			err = redisDatabase.Exists(context.Background(), leaderboardID)
			Expect(err).To(MatchError(redis.NewKeyNotFoundError(leaderboardID)))
		})

		It("should fail if invalid connection to Redis", func() {
			leaderboardID := uuid.NewV4().String()
			err := faultyLeaderboards.RemoveLeaderboard(NewEmptyCtx(), leaderboardID)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

	Describe("getting many members at once", func() {
		It("should return all member details", func() {
			lbID := uuid.NewV4().String()
			for i := 0; i < 100; i++ {
				leaderboards.SetMemberScore(NewEmptyCtx(), lbID, fmt.Sprintf("member-%d", i), int64(100-i), false, "")
			}

			members, err := leaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"member-10", "member-30", "member-20"}, "desc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			Expect(members[0].PublicID).To(Equal("member-10"))
			Expect(members[0].Rank).To(Equal(11))
			Expect(members[0].Score).To(Equal(int64(90)))

			Expect(members[1].PublicID).To(Equal("member-20"))
			Expect(members[1].Rank).To(Equal(21))
			Expect(members[1].Score).To(Equal(int64(80)))

			Expect(members[2].PublicID).To(Equal("member-30"))
			Expect(members[2].Rank).To(Equal(31))
			Expect(members[2].Score).To(Equal(int64(70)))
		})

		It("should return all member details using reverse rank", func() {
			lbID := uuid.NewV4().String()
			for i := 0; i < 100; i++ {
				leaderboards.SetMemberScore(NewEmptyCtx(), lbID, fmt.Sprintf("member-%d", i), int64(100-i), false, "")
			}

			members, err := leaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"member-10", "member-30", "member-20"}, "asc", false)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			Expect(members[0].PublicID).To(Equal("member-30"))
			Expect(members[0].Rank).To(Equal(70))
			Expect(members[0].Score).To(Equal(int64(70)))

			Expect(members[1].PublicID).To(Equal("member-20"))
			Expect(members[1].Rank).To(Equal(80))
			Expect(members[1].Score).To(Equal(int64(80)))

			Expect(members[2].PublicID).To(Equal("member-10"))
			Expect(members[2].Rank).To(Equal(90))
			Expect(members[2].Score).To(Equal(int64(90)))
		})

		It("should return all member details including score expiration timestamp", func() {
			lbID := uuid.NewV4().String()
			for i := 1; i <= 100; i++ {
				ttl := ""
				if i%30 == 0 {
					ttl = "15"
				}
				leaderboards.SetMemberScore(NewEmptyCtx(), lbID, fmt.Sprintf("member-%d", i), int64(100-i), false, ttl)
			}

			members, err := leaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"member-10", "member-30", "member-20"}, "desc", true)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(3))

			Expect(members[0].PublicID).To(Equal("member-10"))
			Expect(members[0].Rank).To(Equal(10))
			Expect(members[0].Score).To(Equal(int64(90)))
			Expect(members[0].ExpireAt).To(Equal(0))

			Expect(members[1].PublicID).To(Equal("member-20"))
			Expect(members[1].Rank).To(Equal(20))
			Expect(members[1].Score).To(Equal(int64(80)))
			Expect(members[1].ExpireAt).To(Equal(0))

			Expect(members[2].PublicID).To(Equal("member-30"))
			Expect(members[2].Rank).To(Equal(30))
			Expect(members[2].ExpireAt).To(BeNumerically("~", time.Now().Unix()+15, 1))
			Expect(members[2].Score).To(Equal(int64(70)))
		})

		It("should return empty list if invalid leaderboard id", func() {
			lbID := uuid.NewV4().String()
			members, err := leaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"test"}, "desc", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(HaveLen(0))
		})

		It("should return empty list if invalid members", func() {
			lbID := uuid.NewV4().String()

			for i := 0; i < 10; i++ {
				leaderboards.SetMemberScore(NewEmptyCtx(), lbID, fmt.Sprintf("member-%d", i), int64(100-i), false, "")
			}

			members, err := leaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"member-0", "invalid-member"}, "desc", false)
			Expect(err).NotTo(HaveOccurred())

			Expect(members).To(HaveLen(1))
			Expect(members[0].PublicID).To(Equal("member-0"))
			Expect(members[0].Rank).To(Equal(1))
			Expect(members[0].Score).To(Equal(int64(100)))
		})

		It("should fail with faulty redis", func() {
			lbID := uuid.NewV4().String()
			_, err := faultyLeaderboards.GetMembers(NewEmptyCtx(), lbID, []string{"member-example"}, "desc", false)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
		})
	})

})
