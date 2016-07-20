// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package util_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/topfreegames/podium/util"
)

var _ = Describe("Expires Helper", func() {

	It("should get expiration for year 2016", func() {
		exp, err := util.GetExpireAt("league_year2016")
		Expect(err).NotTo(HaveOccurred())
		Expect(exp).To(BeNumerically(">", -1))
	})
})
