#!/bin/bash -ex

DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
VERSION=${1:-0}
BUILD_DIR=${DIR}/../build/bin
mkdir -p "${BUILD_DIR}"
cd "$DIR"

GIT_SHA=${DRONE_COMMIT_SHA:-unknown}
BUILD_TIME=$(date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS="-s -w \
    -X github.com/syncloud/redirect/version.GitSha=${GIT_SHA} \
    -X github.com/syncloud/redirect/version.BuildNumber=${VERSION} \
    -X github.com/syncloud/redirect/version.BuildTime=${BUILD_TIME}"

export CGO_ENABLED=0

go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/api" ./cmd/api
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/www" ./cmd/www
go build -ldflags "${LDFLAGS}" -o "${BUILD_DIR}/cli" ./cmd/cli
