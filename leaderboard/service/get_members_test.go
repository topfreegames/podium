package service_test

import (
	"context"
	"fmt"
	"time"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/model"
	"github.com/topfreegames/podium/leaderboard/service"
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
			&database.Member{
				Member: "member1",
				Score:  float64(1),
				Rank:   int64(0),
				TTL:    time.Unix(10000, 0),
			},
			nil,
			&database.Member{
				Member: "member3",
				Score:  float64(3),
				Rank:   int64(1),
				TTL:    time.Unix(10001, 0),
			},
		}

		membersReturn := []*model.Member{
			&model.Member{
				PublicID: "member1",
				Score:    1,
				Rank:     0 + 0 + 1,
				ExpireAt: 10000,
			},
			nil,
			&model.Member{
				PublicID: "member3",
				Score:    3,
				Rank:     0 + 1 + 1,
				ExpireAt: 10001,
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
