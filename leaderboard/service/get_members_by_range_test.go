package service_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service GetMembersByRange", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var start int = 0
	var stop int = 10
	var order string = "asc"

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
				Rank:   0,
			},
			&database.Member{
				Member: "member2",
				Score:  float64(2),
				Rank:   1,
			},
		}

		membersReturn := []*model.Member{
			&model.Member{
				PublicID: "member1",
				Score:    1,
				Rank:     1,
			},
			&model.Member{
				PublicID: "member2",
				Score:    2,
				Rank:     2,
			},
		}

		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(start), gomock.Eq(stop), gomock.Eq(order)).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetMembersByRange(context.Background(), leaderboard, start, stop, order)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(start), gomock.Eq(stop), gomock.Eq(order)).Return(nil, fmt.Errorf("Database error example"))

		_, err := svc.GetMembersByRange(context.Background(), leaderboard, int(start), int(stop), order)
		Expect(err).To(Equal(service.NewGeneralError("get members by range", "Database error example")))
	})
})
