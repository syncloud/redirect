#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
REDIRECT_DIR=/var/www/redirect
STAGE=/tmp/syncloud-redirect

PKGS="docker.io default-mysql-client python3 openssl"
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

chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR" "$REDIRECT_DIR/config.cfg" "$REDIRECT_DIR/secret.cfg" "$REDIRECT_DIR/payments.cfg"

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
    python3 -c "
import configparser
c = configparser.ConfigParser()
c.read(['$REDIRECT_DIR/config.cfg', '$REDIRECT_DIR/secret.cfg'])
print(c['$1']['$2'])
"
}

SYNCLOUD_DOMAIN=$(cfg redirect domain)

if dpkg -s apache2 >/dev/null 2>&1; then
    systemctl stop apache2 2>/dev/null || true
    systemctl disable apache2 2>/dev/null || true
    DEBIAN_FRONTEND=noninteractive apt-get purge -y 'apache2*' 2>/dev/null || true
    DEBIAN_FRONTEND=noninteractive apt-get autoremove -y 2>/dev/null || true
    rm -rf /etc/apache2
fi

if dpkg -s nginx >/dev/null 2>&1; then
    systemctl stop nginx 2>/dev/null || true
    systemctl disable nginx 2>/dev/null || true
    DEBIAN_FRONTEND=noninteractive apt-get purge -y 'nginx*' 'libnginx*' 2>/dev/null || true
    rm -rf /etc/nginx
fi
docker rm -f frps 2>/dev/null || true

if crontab -l 2>/dev/null | grep -q certbot; then
    crontab -l 2>/dev/null | grep -v certbot | crontab -
fi

if [ "$SYNCLOUD_DOMAIN" != "syncloud.test" ]; then
    if ! dpkg -s php-fpm >/dev/null 2>&1; then
        apt-get install -y --no-install-recommends php-fpm php-cli php-mysql php-gd php-curl php-mbstring php-xml php-zip
    fi
    POOL=$(ls /etc/php/*/fpm/pool.d/www.conf 2>/dev/null | head -1)
    if [ -n "$POOL" ]; then
        sed -i 's#^listen = .*#listen = 127.0.0.1:9000#' "$POOL"
        systemctl restart "php$(php -r 'echo PHP_MAJOR_VERSION.".".PHP_MINOR_VERSION;')-fpm" 2>/dev/null \
            || systemctl restart 'php*-fpm.service' 2>/dev/null || true
    fi
fi

install -d /etc/caddy
install -m 0644 "$STAGE/common/caddy/Caddyfile" /etc/caddy/Caddyfile
if [ "$SYNCLOUD_DOMAIN" = "syncloud.test" ]; then
    echo "tls internal" > /etc/caddy/tls.caddy
else
    printf 'tls {\n\tdns route53\n}\n' > /etc/caddy/tls.caddy
fi

REDIRECT_DOMAIN=$SYNCLOUD_DOMAIN
AWS_ENDPOINT_URL=
. "$STAGE/common/caddy/env/$SYNCLOUD_DOMAIN.env"

CADDY_EXTRA=()
[ -n "$AWS_ENDPOINT_URL" ] && CADDY_EXTRA+=(-e "AWS_ENDPOINT_URL=$AWS_ENDPOINT_URL")

CADDY_IMAGE=syncloud/caddy:${TAG##*:}
docker pull "$CADDY_IMAGE"
docker rm -f caddy 2>/dev/null || true
docker run -d \
    --name caddy \
    --restart=unless-stopped \
    --network=host \
    -e AWS_ACCESS_KEY_ID="$(cfg aws access_key_id)" \
    -e AWS_SECRET_ACCESS_KEY="$(cfg aws secret_access_key)" \
    -e ACME_CA="$ACME_CA" \
    -e REDIRECT_DOMAIN="$SYNCLOUD_DOMAIN" \
    -e STORE_DOMAIN="$STORE_DOMAIN" \
    -e STORE_API_DOMAIN="$STORE_API_DOMAIN" \
    -e SHOP_DOMAIN="$SHOP_DOMAIN" \
    -e SITE_DOMAIN="$SITE_DOMAIN" \
    -e GRAFANA_DOMAIN="$GRAFANA_DOMAIN" \
    "${CADDY_EXTRA[@]}" \
    -v /etc/caddy:/etc/caddy:ro \
    -v /var/www:/var/www:ro \
    -v caddy_data:/data \
    "$CADDY_IMAGE"

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

for name in redirect-api redirect-www node-exporter caddy; do
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
