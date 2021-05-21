package service_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service RemoveLeaderboard", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "testKey"

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return nil if all is OK", func() {
		mock.EXPECT().RemoveLeaderboard(gomock.Any(), gomock.Eq(leaderboard)).Return(nil)

		err := svc.RemoveLeaderboard(context.Background(), leaderboard)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().RemoveLeaderboard(gomock.Any(), gomock.Eq(leaderboard)).Return(database.NewGeneralError("unknown error"))

		err := svc.RemoveLeaderboard(context.Background(), leaderboard)
		Expect(err).To(
			Equal(
				service.NewGeneralError(
					"remove leaderboard",
					database.NewGeneralError("unknown error").Error(),
				),
			),
		)
	})
})
