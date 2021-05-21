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

var _ = Describe("Service RemoveMembers", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var members []string = []string{"member", "member2"}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return nil if all is OK", func() {
		mock.EXPECT().RemoveMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(members)).Return(nil)

		err := svc.RemoveMembers(context.Background(), leaderboard, members)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().RemoveMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(members)).Return(fmt.Errorf("unknown error"))

		err := svc.RemoveMembers(context.Background(), leaderboard, members)
		Expect(err).To(Equal(service.NewGeneralError("remove members", "unknown error")))
	})
})
