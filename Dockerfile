# podium - webhook dispatching service
# https://github.com/topfreegames/podium
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2019 Top Free Games <backend@tfgco.com>

FROM golang:1.16.3-alpine as build

MAINTAINER TFG Co <backend@tfgco.com>

COPY . /podium

WORKDIR /podium

RUN apk update && apk add make && make setup && make build

FROM alpine:3.12

COPY --from=build /podium/bin/podium /podium
COPY --from=build /podium/config/default.yaml /default.yaml

RUN chmod +x /podium

ENTRYPOINT ["/podium", "-c", "/default.yaml"]
