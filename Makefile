PACKAGES = $(shell glide novendor)
GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/go-leaderboard@.@g | egrep -v "^[.]$$")

setup:
	@go get -u github.com/Masterminds/glide/...
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get github.com/gordonklaus/ineffassign
	@glide install

test: test-redis
	@ginkgo --cover $(GODIRS)
	@make test-redis-kill

test-coverage: test
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-html: test-coverage
	@go tool cover -html=_build/test-coverage-all.out

# get a redis instance up (localhost:1234)
test-redis:
	@redis-server --port 1234 --daemonize yes; sleep 1
	@redis-cli -p 1234 info > /dev/null

# kill this redis instance (localhost:1234)
test-redis-kill:
	@-redis-cli -p 1234 shutdown
