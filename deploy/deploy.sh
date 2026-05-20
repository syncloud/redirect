#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
REDIRECT_DIR=/var/www/redirect
STAGE=/tmp/syncloud-redirect

PKGS="docker.io apache2 default-mysql-client confget"
if ! dpkg -s $PKGS >/dev/null 2>&1; then
    apt-get update
    apt-get install -y --no-install-recommends $PKGS
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

if ! id -u redirect >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" redirect
fi
REDIRECT_UID=$(id -u redirect)
REDIRECT_GID=$(id -g redirect)

chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR" "$REDIRECT_DIR/config.cfg" "$REDIRECT_DIR/secret.cfg"

mkdir -p "$REDIRECT_DIR/current"

rm -rf "$REDIRECT_DIR/current/www"
cp -r "$STAGE/web" "$REDIRECT_DIR/current/www"

rm -rf "$REDIRECT_DIR/current/bin"
cp -r "$STAGE/bin" "$REDIRECT_DIR/current/bin"
chmod -R +x "$REDIRECT_DIR/current/bin"

rm -rf "$REDIRECT_DIR/current/db"
cp -r "$STAGE/db" "$REDIRECT_DIR/current/db"

chown -R "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR/current"

cfg() {
    confget -f "$REDIRECT_DIR/secret.cfg" -s "$1" "$2" 2>/dev/null \
    || confget -f "$REDIRECT_DIR/config.cfg" -s "$1" "$2"
}

SYNCLOUD_DOMAIN=$(cfg redirect domain)
install -m 0644 "$STAGE/common/apache/redirect.conf" /etc/apache2/sites-available/redirect.conf
if ! grep -q "^export SYNCLOUD_DOMAIN=" /etc/apache2/envvars; then
    echo "export SYNCLOUD_DOMAIN=$SYNCLOUD_DOMAIN" >> /etc/apache2/envvars
fi
a2dissite 000-default 2>/dev/null || true
a2ensite redirect
a2enmod rewrite ssl proxy proxy_http
systemctl restart apache2 2>/dev/null || apachectl restart

crontab -u redirect "$STAGE/common/cron/crontab"

DB_HOST=$(cfg mysql host)
DB_USER=$(cfg mysql user)
DB_PASS=$(cfg mysql passwd)
DB_NAME=$(cfg mysql db)
DB_TARGET_VERSION=$(awk -F"'" '/insert into db_version/ {v=$2} END{print v}' "$STAGE/db/update.sql")
MYSQL="mysql --host=$DB_HOST --user=$DB_USER --password=$DB_PASS"
if ! $MYSQL -e "use $DB_NAME" 2>/dev/null; then
    $MYSQL -e "create database $DB_NAME"
    $MYSQL "$DB_NAME" < "$STAGE/db/init.sql"
fi
DB_CURRENT_VERSION=$($MYSQL -N -B "$DB_NAME" -e "select version from db_version order by timestamp desc limit 1" 2>/dev/null || true)
if [ "$DB_CURRENT_VERSION" != "$DB_TARGET_VERSION" ]; then
    $MYSQL "$DB_NAME" < "$STAGE/db/update.sql"
fi

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

NODE_EXPORTER_IMAGE=prom/node-exporter:v1.8.2
docker pull "$NODE_EXPORTER_IMAGE"
docker rm -f node-exporter 2>/dev/null || true
docker run -d \
    --name node-exporter \
    --restart=unless-stopped \
    --net=host \
    --pid=host \
    -v /:/host:ro \
    "$NODE_EXPORTER_IMAGE" \
    --path.rootfs=/host

for name in redirect-api redirect-www node-exporter; do
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
