# podium
# https://github.com/topfreegames/podium
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2016 Top Free Games <backend@tfgco.com>
# Forked from
# https://github.com/dayvson/go-leaderboard
# Copyright © 2013 Maxwell Dayvson da Silva

GODIRS = $(shell go list ./... | grep -v /vendor/ | sed s@github.com/topfreegames/podium@.@g | egrep -v "^[.]$$")
MYIP = $(shell ifconfig | egrep inet | egrep -v inet6 | egrep -v 127.0.0.1 | awk ' { print $$2 } ')
OS = "$(shell uname | awk '{ print tolower($$0) }')"
PROTOTOOL := go run github.com/uber/prototool/cmd/prototool
LOCAL_GO_MODCACHE = $(shell go env | grep GOMODCACHE | cut -d "=" -f 2 | sed 's/"//g')

help: Makefile ## Show list of commands
	@echo "Choose a command run in "$(PROJECT_NAME)":"
	@echo ""
	@awk 'BEGIN {FS = ":.*?## "} /[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-40s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST) | sort

.PHONY: build proto

setup-hooks: ## Create pre-commit git hooks
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

clear-hooks: ## Remove pre-commit git hooks
	@cd .git/hooks && rm pre-commit

setup: setup-hooks ## Install local dependencies and tidy go mods
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get github.com/gordonklaus/ineffassign
	@go get github.com/uber/prototool/cmd/prototool
	@go mod download

setup-docs: ## Install dependencies necessary for building docs
	@pip2.7 install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

build: ## Build the project
	@go build -o ./bin/podium ./main.go

run: ## Execute the project
	@go run main.go start

test: test-podium test-leaderboard test-client ## Execute all tests

test-podium: ## Execute all API tests
	@ginkgo --cover -r -nodes=1 -skipPackage=leaderboard,client ./

test-leaderboard: ## Execute all leaderboard tests
	@cd leaderboard && ginkgo --cover -r -nodes=1 ./

test-client: ## Execute all client tests
	@cd client && ginkgo --cover -r -nodes=1 ./

coverage: ## Generate code coverage file
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f | sed -e "s#v2##g" >> _build/test-coverage-all.out; done'

test-coverage-html: test coverage ## Generate html page with code coverage information
	@go tool cover -html=_build/test-coverage-all.out

docker-build: ## Build docker-compose services
	@docker build -f ./build/Dockerfile -t podium .

docker-run: ## Run podium inside Docker
	@docker run -i -t --rm -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=6379 -p 8080:80 podium

docker-run-redis: ## Run a redis instance in Docker
	@docker run --name=redis -d -p 6379:6379 redis:6.0.9-alpine

docker-run-basic-auth: ## Run podium inside Docker and setup basic auth (admin:12345)
	@docker run -i -t --rm -e BASICAUTH_USERNAME=admin -e BASICAUTH_PASSWORD=12345 -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=6379 -p 8080:80 podium

deployments/docker-compose.yaml: deployments/docker-compose-model.yaml
	@sed "s%<<LOCAL_GO_MODCACHE>>%${LOCAL_GO_MODCACHE}%g" $< > $@

compose-up-dependencies: deployments/docker-compose.yaml ## Run all dependencies using docker-compose
	@docker-compose -f $< up -d redis-node-0 redis-node-1 redis-node-2 redis-standalone initialize-cluster

compose-up-api: deployments/docker-compose.yaml ## Initialize api on composer environment
	@docker-compose -f $< up -d --build podium-api podium-api

compose-test: deployments/docker-compose.yaml compose-up-dependencies ## Execute podium tests using docker-compose
	@docker-compose -f $< up podium-test

compose-down: deployments/docker-compose.yaml ## Stop all dependency containers
	@docker-compose -f $< down

bench-podium-app: build bench-podium-app-run ## Execute benchmark app

bench-podium-app-run: bench-podium-app-kill ## Execute benchmark app
	@rm -rf /tmp/podium-bench.log
	@./bin/podium start -p 8888 -g 8889 -q -c ./config/perf.yaml 2>&1 > /tmp/podium-bench.log &
	@echo "Podium started at http://localhost:8888. GRPC at 8889."

bench-podium-app-kill: ## Stop benchmark app
	@-ps aux | egrep 'podium.+perf.yaml' | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9

rtfd: ## Build and open podium documentation
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html

mock-lib: ## Generate mocks
	@mockgen github.com/topfreegames/podium/lib PodiumInterface | sed 's/mock_lib/mocks/' > lib/mocks/podium.go

proto: ## Generate protobuf files
	@rm proto/podium/api/v1/*.go > /dev/null 2>&1 || true
	@${PROTOTOOL} generate
