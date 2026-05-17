#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
REDIRECT_DIR=/var/www/redirect
IMAGE_NAME=syncloud/redirect

if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

if ! docker info >/dev/null 2>&1; then
    systemctl start docker 2>/dev/null || true
fi

if ! docker info >/dev/null 2>&1; then
    nohup dockerd --storage-driver=vfs >/var/log/dockerd.log 2>&1 &
    for i in $(seq 1 30); do
        if docker info >/dev/null 2>&1; then break; fi
        sleep 1
    done
fi

if ! docker info >/dev/null 2>&1; then
    echo "docker daemon failed to start"
    tail -60 /var/log/dockerd.log 2>/dev/null || true
    exit 1
fi

for svc in redirect.api redirect.www; do
    if systemctl is-active --quiet "$svc"; then
        systemctl stop "$svc"
    fi
    if systemctl is-enabled --quiet "$svc" 2>/dev/null; then
        systemctl disable "$svc"
    fi
done

if ! id -u redirect >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" redirect
fi
REDIRECT_UID=$(id -u redirect)
REDIRECT_GID=$(id -g redirect)

mkdir -p "$REDIRECT_DIR"
chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR"

rm -f "$REDIRECT_DIR/redirect.api.socket" "$REDIRECT_DIR/redirect.www.socket"

docker pull "$TAG"

run_container() {
    local name=$1
    local bin=$2
    docker rm -f "$name" 2>/dev/null || true
    docker run -d \
        --name "$name" \
        --restart=unless-stopped \
        --network=host \
        --user "$REDIRECT_UID:$REDIRECT_GID" \
        -v "$REDIRECT_DIR:$REDIRECT_DIR" \
        "$TAG" "/usr/local/bin/$bin" --mail-dir /app/emails
}

run_container redirect-api api
run_container redirect-www www

for name in redirect-api redirect-www; do
    for i in $(seq 1 30); do
        if docker ps -q --filter name="$name" --filter status=running | grep -q .; then
            break
        fi
        sleep 2
    done
    if ! docker ps -q --filter name="$name" --filter status=running | grep -q .; then
        echo "container $name is not running:"
        docker ps -a --filter name="$name"
        docker logs "$name" 2>&1 | tail -40
        exit 1
    fi
done

docker image prune -f
