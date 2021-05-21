package service_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/leaderboard/v2/database"
	"github.com/topfreegames/podium/leaderboard/v2/model"
	"github.com/topfreegames/podium/leaderboard/v2/service"
)

var _ = Describe("Service GetTopPercentage", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var pageSize int = 10
	var amount int = 30
	var maxMembers int = 3
	var order string = "asc"

	membersReturnedByDatabase := []*database.Member{
		{
			Member: "member1",
			Score:  float64(1),
			Rank:   int64(0),
		},
		{
			Member: "member2",
			Score:  float64(2),
			Rank:   int64(1),
		},
		{
			Member: "member3",
			Score:  float64(3),
			Rank:   int64(2),
		},
	}

	expectedMembersToReturn := []*model.Member{
		{
			PublicID: "member1",
			Score:    1,
			Rank:     1,
		},
		{
			PublicID: "member2",
			Score:    2,
			Rank:     2,
		},
		{
			PublicID: "member3",
			Score:    3,
			Rank:     3,
		},
	}

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("With order = desc", func() {
		order = "desc"

		It("Should return top percentage members if everything is OK", func() {
			amount = 3
			maxMembers = 10

			mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(100, nil)
			mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Eq(2), gomock.Eq(order)).Return(membersReturnedByDatabase, nil)

			members, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(members).To(HaveLen(3))
			Expect(members).To(Equal(expectedMembersToReturn))
		})
	})

	Describe("With order = asc", func() {
		order = "asc"

		It("Should return percentage members if everything is OK", func() {
			amount = 3
			maxMembers = 10

			membersDatabaseWillReturn := []*database.Member{
				membersReturnedByDatabase[2],
				membersReturnedByDatabase[1],
				membersReturnedByDatabase[0],
			}

			expectedReturn := []*model.Member{
				expectedMembersToReturn[2],
				expectedMembersToReturn[1],
				expectedMembersToReturn[0],
			}

			mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(100, nil)
			mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Eq(2), gomock.Eq(order)).Return(membersDatabaseWillReturn, nil)

			members, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(members).To(HaveLen(3))
			Expect(members).To(Equal(expectedReturn))
		})
	})

	It("Should use desc when order is invalid", func() {
		order = "not_valid_order"

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(100, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Eq(2), gomock.Eq("desc")).Return(membersReturnedByDatabase, nil)

		members, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(members).To(HaveLen(3))
		Expect(members).To(Equal(expectedMembersToReturn))
	})

	It("Should return PercentageError if percentage is grater than 100 or small than 1", func() {
		_, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, 0, maxMembers, order)
		Expect(err).To(MatchError(service.NewPercentageError(0)))

		_, err = svc.GetTopPercentage(context.Background(), leaderboard, pageSize, 101, maxMembers, order)
		Expect(err).To(MatchError(service.NewPercentageError(101)))
	})

	It("Should return no members when percentage is small", func() {
		amount = 1
		maxMembers = 10
		order = "desc"

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(10, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Eq(0), gomock.Eq(order)).Return([]*database.Member{}, nil)

		members, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(members).To(HaveLen(0))
	})

	It("Should return maxMembers if more members are returned by the database", func() {
		amount = 20
		maxMembers = 3
		order = "desc"

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(1000, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Any(), gomock.Eq(order)).Return(membersReturnedByDatabase, nil)

		members, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(members).To(HaveLen(maxMembers))
	})

	It("Should return error if amount is smaller than one", func() {
		amount = 0
		order = "desc"

		_, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).Should(HaveOccurred())
	})

	It("Should return error if amount is greater than 100", func() {
		amount = 101
		order = "desc"

		_, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).Should(HaveOccurred())
	})

	It("Should return error when GetTotalMembers return error", func() {
		amount = 10
		maxMembers = 3
		order = "desc"

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(0, errors.New("Database error example"))
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

		_, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).Should(HaveOccurred())
	})

	It("Should return error when GetOrderedMembers return error", func() {
		amount = 10
		maxMembers = 3
		order = "desc"

		mock.EXPECT().GetTotalMembers(gomock.Any(), gomock.Eq(leaderboard)).Return(100, nil)
		mock.EXPECT().GetOrderedMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(0), gomock.Eq(2), gomock.Eq(order)).Return([]*database.Member{}, errors.New("Database error example"))

		_, err := svc.GetTopPercentage(context.Background(), leaderboard, pageSize, amount, maxMembers, order)
		Expect(err).Should(HaveOccurred())
	})
})
