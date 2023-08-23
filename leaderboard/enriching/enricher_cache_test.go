//go:build unit

package enriching

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("Enricher cache get tests", func() {
	It("should return false and error if redis fails", func() {

	})

	It("should return false if one or more members are not found", func() {
	})

	It("should return true and the data if all members are found", func() {
	})
})

var _ = Describe("Ericher cache set tests", func() {
	It("should set the data in redis", func() {
	})

	It("should return error if redis fails", func() {
	})
})
