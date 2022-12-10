#!/bin/sh -xe
cd docker
ARCH=$1

TAG=latest
if [ -n "$DRONE_TAG" ]; then
    TAG=$DRONE_TAG
fi
IMAGE="syncloud/redirect-test-${ARCH}:$TAG"

set +ex
while ! docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD; do
  echo "retry login"
  sleep 10
done
set -ex

docker build -f Dockerfile.redirect-test -t ${IMAGE} .

set -ex
while ! docker push ${IMAGE}; do
  echo "retry push"
  sleep 10
done
set +ex
