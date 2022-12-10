#!/bin/sh
ARCH=$1

TAG=latest
if [ -n "$DRONE_TAG" ]; then
    TAG=$DRONE_TAG
fi
IMAGE="syncloud/redirect-test-${ARCH}:$TAG"

while ! docker login -u $DOCKER_USERNAME -p $DOCKER_PASSWORD; do
  echo "retry login"
  sleep 10
done

while ! docker build -f Dockerfile.redirect-test -t ${IMAGE} .; do
  echo "retry build"
  sleep 10
done

while ! docker push ${IMAGE}; do
  echo "retry push"
  sleep 10
done
