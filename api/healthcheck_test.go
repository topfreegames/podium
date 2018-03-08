// podium
// https://github.com/topfreegames/podium
// Licensed under the MIT license:
// http://www.opensource.org/licenses/mit-license
// Copyright © 2016 Top Free Games <backend@tfgco.com>
// Forked from
// https://github.com/dayvson/go-leaderboard
// Copyright © 2013 Maxwell Dayvson da Silva

package api_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Healthcheck Handler", func() {
	It("Should respond with default WORKING string", func() {
		a := GetDefaultTestApp()
		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("WORKING"))
	})

	It("Should respond with customized WORKING string", func() {
		a := GetDefaultTestApp()

		a.Config.Set("healthcheck.workingText", "OTHERWORKING")
		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("OTHERWORKING"))
	})

	It("Should fail if redis failing", func() {
		a := GetDefaultTestApp()
		a.RedisClient.Client = GetFaultyRedis()

		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(500))
		Expect(body).To(ContainSubstring("connection refused"))
	})
})
