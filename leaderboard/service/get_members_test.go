package service_test

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service GetMembers", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var order string = "asc"
	var members []string = []string{"member1", "member2", "member3"}
	var includeTTL bool = true

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return members slice if all is OK", func() {
		membersDatabaseReturn := []*database.Member{
			{
				Member: "member1",
				Score:  float64(1),
				Rank:   int64(0),
				TTL:    time.Unix(10000, 0),
			},
			nil,
			{
				Member: "member3",
				Score:  float64(3),
				Rank:   int64(1),
				TTL:    time.Time{},
			},
		}

		membersReturn := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
				Rank:     1,
				ExpireAt: 10000,
			},
			{
				PublicID: "member3",
				Score:    3,
				Rank:     2,
				ExpireAt: 0,
			},
		}

		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq("member1"), gomock.Eq("member2"), gomock.Eq("member3")).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetMembers(context.Background(), leaderboard, members, order, includeTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should order members", func() {
		membersDatabaseReturn := []*database.Member{
			{
				Member: "member1",
				Score:  float64(1),
				Rank:   int64(0),
				TTL:    time.Unix(10000, 0),
			},
			{
				Member: "member3",
				Score:  float64(3),
				Rank:   int64(2),
				TTL:    time.Unix(10000, 0),
			},
			nil,
			{
				Member: "member2",
				Score:  float64(2),
				Rank:   int64(1),
				TTL:    time.Time{},
			},
		}

		membersReturn := []*model.Member{
			{
				PublicID: "member1",
				Score:    1,
				Rank:     1,
				ExpireAt: 10000,
			},
			{
				PublicID: "member2",
				Score:    2,
				Rank:     2,
				ExpireAt: 0,
			},
			{
				PublicID: "member3",
				Score:    3,
				Rank:     3,
				ExpireAt: 10000,
			},
		}

		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq("member1"), gomock.Eq("member2"), gomock.Eq("member3")).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetMembers(context.Background(), leaderboard, members, order, includeTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq("member1"), gomock.Eq("member2"), gomock.Eq("member3")).Return(nil, fmt.Errorf("Database error example"))

		_, err := svc.GetMembers(context.Background(), leaderboard, members, order, includeTTL)
		Expect(err).To(Equal(service.NewGeneralError("get members", "Database error example")))
	})
})
