#!/bin/bash

VERSION=$(cat ./api/version.go | grep "var VERSION" | awk ' { print $4 } ' | sed s/\"//g)

cp ./config/default.yaml ./dev

docker build -t podium .
docker build -t podium-dev -f ./DevDockerfile .

docker login -u="$DOCKER_USERNAME" -p="$DOCKER_PASSWORD"

docker tag podium:latest tfgco/podium:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/podium:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag podium:latest tfgco/podium:latest
docker push tfgco/podium

docker tag podium-dev:latest tfgco/podium-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker push tfgco/podium-dev:$VERSION.$TRAVIS_BUILD_NUMBER
docker tag podium-dev:latest tfgco/podium-dev:latest
docker push tfgco/podium-dev

DOCKERHUB_LATEST=$(python ./scripts/get_latest_tag.py)

if [ "$DOCKERHUB_LATEST" != "$VERSION.$TRAVIS_BUILD_NUMBER" ]; then
    echo "Last version is not in docker hub!"
    echo "docker hub: $DOCKERHUB_LATEST, expected: $VERSION.$TRAVIS_BUILD_NUMBER"
    exit 1
fi
