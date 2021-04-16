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

.PHONY: build proto

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

clear-hooks:
	@cd .git/hooks && rm pre-commit

setup: setup-hooks
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get github.com/gordonklaus/ineffassign
	@go get github.com/uber/prototool/cmd/prototool
	@go mod tidy

setup-docs:
	@pip2.7 install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

build:
	@go build -o ./bin/podium ./main.go

run:
	@go run main.go start

test: test-podium test-leaderboard

test-podium:
	@ginkgo --cover -r -nodes=1 -skipPackage=leaderboard ./

test-leaderboard:
	@cd leaderboard && ginkgo --cover -r -nodes=1 ./

coverage:
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-html: test coverage
	@go tool cover -html=_build/test-coverage-all.out

docker-build:
	@docker build -f ./build/Dockerfile -t podium .

docker-run:
	@docker run -i -t --rm -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=6379 -p 8080:80 podium

docker-run-redis:
	@docker run --name=redis -d -p 6379:6379 redis:6.0.9-alpine

docker-run-basic-auth:
	@docker run -i -t --rm -e BASICAUTH_USERNAME=admin -e BASICAUTH_PASSWORD=12345 -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=6379 -p 8080:80 podium

compose-up-dependencies:
	@sed "s%<<LOCAL_GO_MODCACHE>>%${LOCAL_GO_MODCACHE}%g" deployments/docker-compose-model.yaml > deployments/docker-compose.yaml
	@docker-compose -f deployments/docker-compose.yaml up -d redis-node-0 redis-node-1 redis-node-2 redis-standalone initialize-cluster

compose-test:
	@sed "s%<<LOCAL_GO_MODCACHE>>%${LOCAL_GO_MODCACHE}%g" deployments/docker-compose-model.yaml > deployments/docker-compose.yaml
	@docker-compose -f deployments/docker-compose.yaml up --build podium-test

compose-down:
	@docker-compose -f deployments/docker-compose.yaml down

bench-podium-app: build bench-podium-app-run

bench-podium-app-run: bench-podium-app-kill
	@rm -rf /tmp/podium-bench.log
	@./bin/podium start -p 8888 -g 8889 -q -c ./config/perf.yaml 2>&1 > /tmp/podium-bench.log &
	@echo "Podium started at http://localhost:8888. GRPC at 8889."

bench-podium-app-kill:
	@-ps aux | egrep 'podium.+perf.yaml' | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9

rtfd:
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html

mock-lib:
	@mockgen github.com/topfreegames/podium/lib PodiumInterface | sed 's/mock_lib/mocks/' > lib/mocks/podium.go

proto:
	@rm proto/podium/api/v1/*.go > /dev/null 2>&1 || true
	@${PROTOTOOL} generate
