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

var _ = Describe("Service GetMember", func() {
	var ctrl *gomock.Controller
	var mock *database.MockDatabase
	var svc *service.Service

	var leaderboard string = "leaderboardTest"
	var order string = "asc"
	var member string = "member1"
	var includeTTL bool = true

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mock = database.NewMockDatabase(ctrl)

		svc = &service.Service{mock}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("Should return member if all is OK", func() {
		membersDatabaseReturn := []*database.Member{
			&database.Member{
				Member: "member1",
				Score:  float64(1),
				Rank:   int64(0),
				TTL:    time.Unix(10000, 0),
			},
		}

		membersReturn := &model.Member{
			PublicID: "member1",
			Score:    1,
			Rank:     1,
			ExpireAt: 10000,
		}

		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq(member)).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetMember(context.Background(), leaderboard, member, order, includeTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should return member with Expire zero if database return empty time TTL", func() {
		membersDatabaseReturn := []*database.Member{
			&database.Member{
				Member: "member1",
				Score:  float64(1),
				Rank:   int64(0),
				TTL:    time.Time{},
			},
		}

		membersReturn := &model.Member{
			PublicID: "member1",
			Score:    1,
			Rank:     1,
			ExpireAt: 0,
		}

		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq(member)).Return(membersDatabaseReturn, nil)

		membersFromService, err := svc.GetMember(context.Background(), leaderboard, member, order, includeTTL)
		Expect(err).NotTo(HaveOccurred())

		Expect(membersFromService).To(Equal(membersReturn))
	})

	It("Should return member not found database return nil member", func() {
		membersDatabaseReturn := []*database.Member{
			nil,
		}

		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq(member)).Return(membersDatabaseReturn, nil)

		_, err := svc.GetMember(context.Background(), leaderboard, member, order, includeTTL)
		Expect(err).To(Equal(service.NewMemberNotFoundError(leaderboard, member)))

	})

	It("Should return error if database return in error", func() {
		mock.EXPECT().GetMembers(gomock.Any(), gomock.Eq(leaderboard), gomock.Eq(order), gomock.Eq(includeTTL), gomock.Eq(member)).Return(nil, fmt.Errorf("Database error example"))

		_, err := svc.GetMember(context.Background(), leaderboard, member, order, includeTTL)
		Expect(err).To(Equal(service.NewGeneralError("get member", "Database error example")))
	})
})
