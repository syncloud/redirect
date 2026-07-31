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
    apt-get -o DPkg::Lock::Timeout=300 update
    apt-get -o DPkg::Lock::Timeout=300 install -y --no-install-recommends $PKGS
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

cp "$STAGE/config.cfg" "$REDIRECT_DIR/config.cfg"

chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR" "$REDIRECT_DIR/config.cfg" "$REDIRECT_DIR/secret.cfg" "$REDIRECT_DIR/payments.cfg"

mkdir -p "$REDIRECT_DIR/current"

rm -rf "$REDIRECT_DIR/current/www"
cp -r "$STAGE/web" "$REDIRECT_DIR/current/www"

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

# transitional: drop this apache removal after prod is migrated to caddy
if dpkg -s apache2 >/dev/null 2>&1; then
    systemctl stop apache2 2>/dev/null || true
    systemctl disable apache2 2>/dev/null || true
    DEBIAN_FRONTEND=noninteractive apt-get -o DPkg::Lock::Timeout=300 purge -y 'apache2*' 2>/dev/null || true
    DEBIAN_FRONTEND=noninteractive apt-get -o DPkg::Lock::Timeout=300 autoremove -y 2>/dev/null || true
    rm -rf /etc/apache2
fi

if crontab -l 2>/dev/null | grep -q certbot; then
    crontab -l 2>/dev/null | grep -v certbot | crontab -
fi

# php-fpm backs the opencart shop; pending migration of the shop to a redirect page
if ! dpkg -s php-fpm >/dev/null 2>&1; then
    apt-get -o DPkg::Lock::Timeout=300 install -y --no-install-recommends php-fpm php-cli php-mysql php-gd php-curl php-mbstring php-xml php-zip
fi
POOL=$(ls /etc/php/*/fpm/pool.d/www.conf 2>/dev/null | head -1)
if [ -n "$POOL" ]; then
    sed -i 's#^listen = .*#listen = 127.0.0.1:9000#' "$POOL"
    systemctl restart "php$(php -r 'echo PHP_MAJOR_VERSION.".".PHP_MINOR_VERSION;')-fpm" 2>/dev/null \
        || systemctl restart 'php*-fpm.service' 2>/dev/null || true
fi

install -d /etc/caddy /etc/caddy/conf.d
install -m 0644 "$STAGE/common/caddy/Caddyfile" /etc/caddy/Caddyfile
for f in "$STAGE"/common/caddy/conf.d/*.caddy; do
    [ -e "$f" ] && install -m 0644 "$f" /etc/caddy/conf.d/
done

. "$STAGE/common/caddy/env/$SYNCLOUD_DOMAIN.env"

CADDY_IMAGE=syncloud/caddy:${TAG##*:}
docker pull "$CADDY_IMAGE"
docker rm -f caddy 2>/dev/null || true
docker run -d \
    --name caddy \
    --restart=unless-stopped \
    --network=host \
    -e AWS_ACCESS_KEY_ID="$(cfg aws access_key_id)" \
    -e AWS_SECRET_ACCESS_KEY="$(cfg aws secret_access_key)" \
    -e AWS_ENDPOINT_URL="$AWS_ENDPOINT_URL" \
    -e ACME_CA="$ACME_CA" \
    -e DNS_RESOLVER="$DNS_RESOLVER" \
    -e SSL_CERT_FILE="$SSL_CERT_FILE" \
    -e REDIRECT_DOMAIN="$SYNCLOUD_DOMAIN" \
    -e STORE_DOMAIN="$STORE_DOMAIN" \
    -e STORE_API_DOMAIN="$STORE_API_DOMAIN" \
    -e SHOP_DOMAIN="$SHOP_DOMAIN" \
    -e SITE_DOMAIN="$SITE_DOMAIN" \
    -e GRAFANA_DOMAIN="$GRAFANA_DOMAIN" \
    -e RELAY_SITE_SNI="$RELAY_SITE_SNI" \
    -v /etc/caddy:/etc/caddy:ro \
    -v /var/www:/var/www:ro \
    -v caddy_data:/data \
    "$CADDY_IMAGE"

crontab -u redirect "$STAGE/common/cron/crontab"

DB_HOST=$(cfg mysql host)
DB_USER=$(cfg mysql user)
DB_PASS=$(cfg mysql passwd)
DB_NAME=$(cfg mysql db)
MYSQL="mysql --host=$DB_HOST --user=$DB_USER --password=$DB_PASS"
if ! $MYSQL -e "use $DB_NAME" 2>/dev/null; then
    $MYSQL -e "create database $DB_NAME"
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

install -m 0644 -o "$REDIRECT_UID" -g "$REDIRECT_GID" "$STAGE/common/frp/frps.toml" "$REDIRECT_DIR/frps.toml"
FRPS_IMAGE=snowdreamtech/frps:0.61.1
docker pull "$FRPS_IMAGE"
docker rm -f frps 2>/dev/null || true
docker run -d \
    --name frps \
    --restart=unless-stopped \
    --network=host \
    -v "$REDIRECT_DIR/frps.toml:/etc/frp/frps.toml:ro" \
    "$FRPS_IMAGE"

RSPAMD_IMAGE=rspamd/rspamd:3.11
docker pull "$RSPAMD_IMAGE"
docker rm -f rspamd 2>/dev/null || true
rm -rf "$REDIRECT_DIR/rspamd"
mkdir -p "$REDIRECT_DIR/rspamd"
cp -r "$STAGE/common/rspamd/local.d" "$REDIRECT_DIR/rspamd/"
chown -R "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR/rspamd"
docker run -d \
    --name rspamd \
    --restart=unless-stopped \
    --network=host \
    -v "$REDIRECT_DIR/rspamd/local.d:/etc/rspamd/local.d:ro" \
    "$RSPAMD_IMAGE"

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

for name in redirect-api redirect-www node-exporter caddy frps rspamd; do
    for i in $(seq 1 30); do
        if docker ps -q --filter name="$name" --filter status=running | grep -q .; then
            break
        fi
        sleep 2
    done
    if ! docker ps -q --filter name="$name" --filter status=running | grep -q .; then
        echo "container $name is not running:"
        docker ps -a --filter name="$name"
        docker logs "$name" 2>&1 | head -60; docker logs "$name" 2>&1 | tail -40
        exit 1
    fi
done

docker image prune -af
