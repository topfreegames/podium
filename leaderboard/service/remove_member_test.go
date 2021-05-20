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

var _ = Describe("Service RemoveMember", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var member string = "member"

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return nil if all is OK", func() {
		mock.EXPECT().RemoveMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member)).Return(nil)

		err := svc.RemoveMember(context.Background(), leaderboard, member)
		Expect(err).NotTo(HaveOccurred())
	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().RemoveMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(member)).Return(fmt.Errorf("unknown error"))

		err := svc.RemoveMember(context.Background(), leaderboard, member)
		Expect(err).To(Equal(service.NewGeneralError("remove member", "unknown error")))
	})
})
