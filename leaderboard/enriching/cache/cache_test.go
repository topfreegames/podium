package cache

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v8"
	redismock "github.com/go-redis/redismock/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/model"
)

var _ = Describe("Members array to keys array test", func() {
	tenantID := "tenantID"

	It("should return keys correctly", func() {
		members := []*model.Member{
			{
				PublicID: "member1",
			},
			{
				PublicID: "member2",
			},
		}

		keys := getKeysFromMemberArray(tenantID, members)

		Expect(keys).To(HaveLen(2))
		Expect(keys[0]).To(Equal("leaderboards-enrich-caching:tenantID:member1"))
		Expect(keys[1]).To(Equal("leaderboards-enrich-caching:tenantID:member2"))
	})
})

var _ = Describe("Enricher cacheConfig Get tests", func() {
	tenantID := "tenantID"

	It("should return false and error if redis fails", func() {
		redis, redisMock := redismock.NewClientMock()

		members := []*model.Member{
			{
				PublicID: "member1",
			},
		}

		redisMock.ExpectMGet(
			getKeysFromMemberArray(tenantID, members)...,
		).SetErr(errors.New("some error"))

		cache := NewEnricherRedisCache(redis)
		res, hit, err := cache.Get(context.Background(), tenantID, members)

		Expect(res).To(BeNil())
		Expect(hit).To(BeFalse())
		Expect(err).To(HaveOccurred())

	})

	It("should return false if no members are found", func() {
		redis, redisMock := redismock.NewClientMock()

		members := []*model.Member{
			{
				PublicID: "member1",
			},
			{
				PublicID: "member2",
			},
		}

		redisMock.ExpectMGet(
			getKeysFromMemberArray(tenantID, members)...,
		).SetVal([]interface{}{nil, nil})

		cache := NewEnricherRedisCache(redis)
		res, hit, err := cache.Get(context.Background(), tenantID, members)

		Expect(res).To(BeNil())
		Expect(hit).To(BeFalse())
		Expect(err).To(BeNil())
	})

	It("should return false if one or more members are not found", func() {
		redis, redisMock := redismock.NewClientMock()

		members := []*model.Member{
			{
				PublicID: "member1",
			},
			{
				PublicID: "member2",
			},
		}

		mgetExpectedResult := []interface{}{
			"{\"key1\": \"value1\"}",
			nil,
		}

		redisMock.ExpectMGet(
			getKeysFromMemberArray(tenantID, members)...,
		).SetVal(mgetExpectedResult)

		cache := NewEnricherRedisCache(redis)
		res, hit, err := cache.Get(context.Background(), tenantID, members)

		Expect(res).To(BeNil())
		Expect(hit).To(BeFalse())
		Expect(err).To(BeNil())
	})

	It("should return true and the data if all members are found", func() {
		redis, redisMock := redismock.NewClientMock()

		members := []*model.Member{
			{
				PublicID: "member1",
			},
			{
				PublicID: "member2",
			},
		}

		mgetExpectedResult := []interface{}{
			"{\"key1\": \"value1\"}",
			"{\"key1\": \"value1\"}",
		}

		redisMock.ExpectMGet(
			getKeysFromMemberArray(tenantID, members)...,
		).SetVal(mgetExpectedResult)

		cache := NewEnricherRedisCache(redis)
		res, hit, err := cache.Get(context.Background(), tenantID, members)

		expectedResult := map[string]map[string]string{
			"member1": {
				"key1": "value1",
			},
			"member2": {
				"key1": "value1",
			},
		}

		Expect(res).To(Equal(expectedResult))
		Expect(hit).To(BeTrue())
		Expect(err).To(BeNil())
	})
})

var _ = Describe("Ericher cacheConfig Set tests", func() {
	tenantID := "tenantID"

	It("should set the data in redis", func() {
		redis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

		cache := NewEnricherRedisCache(redis)
		members := []*model.Member{
			{
				PublicID: "member1",
				Metadata: map[string]string{
					"key1": "value1",
				},
			},
			{
				PublicID: "member2",
				Metadata: map[string]string{
					"key2": "value2",
				},
			},
		}

		err := cache.Set(context.Background(), tenantID, members, 0)

		res, err := redis.Get(context.Background(), fmt.Sprintf(cacheKeyFormat, tenantID, "member1")).Result()
		Expect(res).To(Equal("{\"key1\":\"value1\"}"))
		Expect(err).To(BeNil())

		res, err = redis.Get(context.Background(), fmt.Sprintf(cacheKeyFormat, tenantID, "member2")).Result()
		Expect(res).To(Equal("{\"key2\":\"value2\"}"))
		Expect(err).To(BeNil())
	})

	AfterEach(func() {
		redis := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
		redis.Del(context.Background(), fmt.Sprintf(cacheKeyFormat, tenantID, "member1"))
		redis.Del(context.Background(), fmt.Sprintf(cacheKeyFormat, tenantID, "member2"))
	})
})
