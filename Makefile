PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/go-leaderboard@.@g | egrep -v "^[.]$$")

setup:
	@go get -u github.com/Masterminds/glide/...
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get github.com/gordonklaus/ineffassign
	@glide install

test:
	@ginkgo --cover $(GODIRS)
