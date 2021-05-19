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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	api "github.com/topfreegames/podium/proto/podium/api/v1"
	"github.com/topfreegames/podium/testing"
)

var _ = Describe("Healthcheck Handler", func() {
	It("Should respond with default WORKING string (http)", func() {
		a := testing.GetDefaultTestApp()
		status, body := testing.Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("WORKING"))
	})

	It("Should respond with default WORKING string (grpc)", func() {
		a := testing.GetDefaultTestApp()

		testing.SetupGRPC(a, func(cli api.PodiumClient) {
			resp, err := cli.HealthCheck(context.Background(), &api.HealthCheckRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.WorkingString).To(Equal("WORKING"))
		})
	})

	It("Should respond with customized WORKING string (http)", func() {
		a := testing.GetDefaultTestApp()
		a.Config.Set("healthcheck.workingText", "OTHERWORKING")
		status, body := testing.Get(a, "/healthcheck")

		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(Equal("OTHERWORKING"))
	})

	It("Should respond with customized WORKING string (grpc)", func() {
		a := testing.GetDefaultTestApp()
		a.Config.Set("healthcheck.workingText", "OTHERWORKING")

		testing.SetupGRPC(a, func(cli api.PodiumClient) {
			resp, err := cli.HealthCheck(context.Background(), &api.HealthCheckRequest{})

			Expect(err).NotTo(HaveOccurred())
			Expect(resp.WorkingString).To(Equal("OTHERWORKING"))
		})
	})

	It("Should fail if redis failing (http)", func() {
		a := testing.GetDefaultTestAppWithFaultyRedis()

		status, body := testing.Get(a, "/healthcheck")

		Expect(status).To(Equal(500))
		Expect(body).To(ContainSubstring("connection refused"))
	})

	It("Should fail if redis failing (grpc)", func() {
		a := testing.GetDefaultTestAppWithFaultyRedis()

		testing.SetupGRPC(a, func(cli api.PodiumClient) {
			resp, err := cli.HealthCheck(context.Background(), &api.HealthCheckRequest{})

			Expect(err).To(HaveOccurred())
			Expect(status.Code(err)).To(Equal(codes.Internal))
			Expect(err.Error()).To(ContainSubstring("connection refused"))
			Expect(resp).To(BeNil())
		})
	})
})
