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
	"context"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/golang/protobuf/ptypes/empty"
	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

var _ = Describe("Healthcheck Handler", func() {
	It("Should respond with default WORKING string (http)", func() {
		a := GetDefaultTestApp()
		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("WORKING"))
	})

	It("Should respond with default WORKING string (grpc)", func() {
		a := GetDefaultTestApp()

		SetupGRPC(a, func(cli api.PodiumAPIClient) {
			resp, err := cli.HealthCheck(context.Background(), &empty.Empty{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.WorkingString).To(Equal("WORKING"))
		})
	})

	It("Should respond with customized WORKING string (http)", func() {
		a := GetDefaultTestApp()
		a.Config.Set("healthcheck.workingText", "OTHERWORKING")
		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("OTHERWORKING"))
	})

	It("Should respond with customized WORKING string (grpc)", func() {
		a := GetDefaultTestApp()
		a.Config.Set("healthcheck.workingText", "OTHERWORKING")

		SetupGRPC(a, func(cli api.PodiumAPIClient) {
			resp, err := cli.HealthCheck(context.Background(), &empty.Empty{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.WorkingString).To(Equal("OTHERWORKING"))
		})
	})

	It("Should fail if redis failing (http)", func() {
		a := GetDefaultTestAppWithFaultyRedis()

		status, body := Get(a, "/healthcheck")

		Expect(status).To(Equal(500))
		Expect(body).To(ContainSubstring("connection refused"))
	})

	It("Should fail if redis failing (grpc)", func() {
		a := GetDefaultTestAppWithFaultyRedis()

		SetupGRPC(a, func(cli api.PodiumAPIClient) {
			resp, err := cli.HealthCheck(context.Background(), &empty.Empty{})

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("connection refused"))
			Expect(resp).To(BeNil())
		})
	})
})
