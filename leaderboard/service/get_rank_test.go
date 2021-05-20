package service_test

import (
	"context"
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service GetRank", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "testKey"
	var member string = "member"
	var order = "asc"
	var rank int = 10

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return member position if database returns OK", func() {
		mock.EXPECT().GetRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(order)).Return(rank, nil)

		rankReturned, err := svc.GetRank(context.Background(), leaderboard, member, order)
		Expect(err).NotTo(HaveOccurred())

		Expect(rankReturned).To(Equal(rank + 1))
	})

	It("Should return error MemberNotFoundError if database return in error", func() {
		mock.EXPECT().GetRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(order)).Return(-1, database.NewMemberNotFoundError(leaderboard, member))

		_, err := svc.GetRank(context.Background(), leaderboard, member, order)
		Expect(err).To(Equal(service.NewMemberNotFoundError(leaderboard, member)))
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().GetRank(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member), gomock.Eq(order)).Return(-1, fmt.Errorf("Database error example"))

		_, err := svc.GetRank(context.Background(), leaderboard, member, order)
		Expect(err).To(Equal(service.NewGeneralError("get rank", "Database error example")))
	})
})
