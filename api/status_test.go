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
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/golang/protobuf/ptypes/empty"
	api "github.com/topfreegames/podium/proto/podium/api/v1"
)

var _ = Describe("Status Handler", func() {
	It("Should respond with status (http)", func() {
		a := GetDefaultTestApp()
		status, body := Get(a, "/status")

		Expect(status).To(Equal(http.StatusOK))

		var result map[string]interface{}
		json.Unmarshal([]byte(body), &result)

		Expect(result["app"]).NotTo(BeNil())

		app := result["app"].(map[string]interface{})
		Expect(app["errorRate"]).To(BeEquivalentTo(0.0))
	})

	It("Should respond with status (grpc)", func() {
		a := GetDefaultTestApp()

		SetupGRPC(a, func(cli api.PodiumClient) {
			resp, err := cli.Status(context.Background(), &empty.Empty{})
			Expect(err).NotTo(HaveOccurred())
			Expect(resp).NotTo(BeNil())
			Expect(resp.ErrorRate).To(BeEquivalentTo(0.0))
		})
	})
})
