package service_test

import (
	"context"
	"fmt"
	"math"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/database"
	"github.com/topfreegames/podium/leaderboard/model"
	"github.com/topfreegames/podium/leaderboard/service"
)

var _ = Describe("Service GetLeaders", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var totalMembers int = 10
	var pageSize int = 3
	var page int = 1
	var order string = "asc"

	var start int = (page - 1) * pageSize
	var stop int = start + pageSize - 1

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
			},
			&database.Member{
				Member: "member2",
				Score:  float64(2),
			},
			&database.Member{
				Member: "member3",
				Score:  float64(3),
			},
		}

		membersReturn := []*model.Member{
			&model.Member{
				PublicID: "member1",
				Score:    1,
				Rank:     0 + 0 + 1,
			},
			&model.Member{
				PublicID: "member2",
				Score:    2,
				Rank:     0 + 1 + 1,
			},
			&model.Member{
				PublicID: "member3",
				Score:    3,
				Rank:     0 + 2 + 1,
			},
		}

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(totalMembers, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(start), gomock.Eq(stop), gomock.Eq(order)).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetLeaders(context.Background(), leaderboard, pageSize, page, order)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should return error if page is less than 1", func() {
		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(totalMembers, nil)

		totalPages := int(math.Ceil(float64(totalMembers) / float64(pageSize)))

		_, err := svc.GetLeaders(context.Background(), leaderboard, pageSize, -1, order)
		Expect(err).To(Equal(service.NewPageOutOfRangeError(-1, totalPages)))
	})

	It("Should return error if page is more than totalPages", func() {
		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(totalMembers, nil)

		totalPages := int(math.Ceil(float64(totalMembers) / float64(pageSize)))

		_, err := svc.GetLeaders(context.Background(), leaderboard, pageSize, totalPages+1, order)
		Expect(err).To(Equal(service.NewPageOutOfRangeError(totalPages+1, totalPages)))
	})

	It("Should return error if database return in error on GetTotalPages", func() {
		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(-1, fmt.Errorf("Database error example"))

		_, err := svc.GetLeaders(context.Background(), leaderboard, pageSize, page, order)
		Expect(err).To(Equal(service.NewGeneralError("get leaders", "Database error example")))
	})

	It("Should return error if database return in error on GetOrderedMembers", func() {
		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(totalMembers, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(start), gomock.Eq(stop), gomock.Eq(order)).Return(nil, fmt.Errorf("Database error example"))

		_, err := svc.GetLeaders(context.Background(), leaderboard, pageSize, page, order)
		Expect(err).To(Equal(service.NewGeneralError("get leaders", "Database error example")))
	})
})
