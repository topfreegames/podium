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
REDIS_CONF_PATH=./scripts/redis.conf
LOCAL_REDIS_PORT=1212
LOCAL_TEST_REDIS_PORT=1234

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

clear-hooks:
	@cd .git/hooks && rm pre-commit

setup: setup-hooks
	@go get github.com/mailru/easyjson/...
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get github.com/gordonklaus/ineffassign
	@dep ensure -v

setup-docs:
	@pip install -q --log /tmp/pip.log --no-cache-dir sphinx recommonmark sphinx_rtd_theme

build:
	@go build $(GODIRS)
	@go build -o ./bin/podium ./main.go

# run app
run: schema-update redis
	@go run main.go start

# run app
run-prod: schema-update redis build
	@./bin/podium start -q -f -c ./config/local.yaml

# get a redis instance up (localhost:1212)
redis: redis-shutdown
	@if [ -z "$$REDIS_PORT" ]; then \
		redis-server $(REDIS_CONF_PATH) && sleep 1 &&  \
		redis-cli -p $(LOCAL_REDIS_PORT) info > /dev/null && \
		echo "REDIS running locally at localhost:$(LOCAL_REDIS_PORT)."; \
	else \
		echo "REDIS running at $$REDIS_PORT"; \
	fi

# kill this redis instance (localhost:1212)
redis-shutdown:
	@-redis-cli -p 1212 shutdown

redis-clear:
	@redis-cli -p 1212 FLUSHDB

test: schema-update test-redis
	@ginkgo --cover -r .
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
	@redis-server --port ${LOCAL_TEST_REDIS_PORT} --daemonize yes; sleep 1
	@redis-cli -p ${LOCAL_TEST_REDIS_PORT} info > /dev/null

# kill this redis instance (localhost:1234)
test-redis-kill:
	@-redis-cli -p ${LOCAL_TEST_REDIS_PORT} shutdown

cross: cross-linux cross-darwin

cross-linux:
	@mkdir -p ./bin
	@echo "Building for linux-i386..."
	@env GOOS=linux GOARCH=386 go build -o ./bin/podium-linux-i386 ./main.go
	@echo "Building for linux-x86_64..."
	@env GOOS=linux GOARCH=amd64 go build -o ./bin/podium-linux-x86_64 ./main.go
	@$(MAKE) cross-exec

cross-darwin:
	@mkdir -p ./bin
	@echo "Building for darwin-i386..."
	@env GOOS=darwin GOARCH=386 go build -o ./bin/podium-darwin-i386 ./main.go
	@echo "Building for darwin-x86_64..."
	@env GOOS=darwin GOARCH=amd64 go build -o ./bin/podium-darwin-x86_64 ./main.go
	@$(MAKE) cross-exec

cross-exec:
	@chmod +x bin/*

docker-build:
	@docker build -t podium .

docker-run:
	@docker run -i -t --rm -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=$(LOCAL_REDIS_PORT) -p 8080:80 podium

docker-run-basic-auth:
	@docker run -i -t --rm -e BASICAUTH_USERNAME=admin -e BASICAUTH_PASSWORD=12345 -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=$(LOCAL_REDIS_PORT) -p 8080:80 podium

docker-shell:
	@docker run -it --rm -e PODIUM_REDIS_HOST=$(MYIP) -e PODIUM_REDIS_PORT=$(LOCAL_REDIS_PORT) --entrypoint "/bin/bash" podium

docker-dev-build:
	@docker build -t podium-dev -f ./DevDockerfile .

docker-dev-run:
	@docker run -i -t --rm -p 8080:8080 podium-dev

bench-podium-app: build bench-podium-app-run

bench-podium-app-run: bench-podium-app-kill
	@rm -rf /tmp/podium-bench.log
	@./bin/podium start -p 8888 -f -q -c ./config/perf.yaml 2>&1 > /tmp/podium-bench.log &
	@echo "Podium started at http://localhost:8888."

bench-podium-app-kill:
	@-ps aux | egrep 'podium.+perf.yaml' | egrep -v egrep | awk ' { print $$2 } ' | xargs kill -9

# get a redis instance up (localhost:1224)
bench-redis: bench-redis-kill
	@redis-server --port 1224 --daemonize yes; sleep 1
	@redis-cli -p 1224 info > /dev/null

# kill this redis instance (localhost:1224)
bench-redis-kill:
	@-redis-cli -p 1224 shutdown

bench-run:
	@go test -benchmem -bench . -benchtime 5s ./bench/...

bench-seed:
	@go run bench/seed/main.go

ci-bench-run:
	@mkdir -p ./bench-data
	@if [ -f "./bench-data/new.txt" ]; then \
		mv ./bench-data/new.txt ./bench-data/old.txt; \
	fi
	@go test -benchmem -bench . -benchtime 5s ./bench/... > ./bench-data/new.txt
	@echo "Benchmark Results:"
	@cat ./bench-data/new.txt
	@echo
	@-if [ -f "./bench-data/old.txt" ]; then \
		echo "Comparison to previous build:" && \
		benchcmp ./bench-data/old.txt ./bench-data/new.txt; \
	fi

rtfd:
	@rm -rf docs/_build
	@sphinx-build -b html -d ./docs/_build/doctrees ./docs/ docs/_build/html
	@open docs/_build/html/index.html

schema-update:
	@go generate ./api/payload.go
