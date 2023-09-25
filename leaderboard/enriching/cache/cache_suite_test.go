package cache

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestEnrichingCache(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Enriching Cache Suite")
}
