#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
REDIRECT_DIR=/var/www/redirect
STAGE=/tmp/syncloud-redirect

PKGS="docker.io apache2 nginx libnginx-mod-stream default-mysql-client python3 openssl"
if ! dpkg -s $PKGS >/dev/null 2>&1; then
    printf '#!/bin/sh\nexit 101\n' > /usr/sbin/policy-rc.d
    chmod +x /usr/sbin/policy-rc.d
    apt-get update
    apt-get install -y --no-install-recommends $PKGS
    rm -f /usr/sbin/policy-rc.d
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
install -m 0644 "$STAGE/common/apache/redirect.conf" /etc/apache2/sites-available/redirect.conf
if ! grep -q "^export SYNCLOUD_DOMAIN=" /etc/apache2/envvars; then
    echo "export SYNCLOUD_DOMAIN=$SYNCLOUD_DOMAIN" >> /etc/apache2/envvars
fi

cat > /etc/apache2/ports.conf <<'PORTS'
Listen 80
<IfModule ssl_module>
    Listen 127.0.0.1:8443
</IfModule>
<IfModule mod_gnutls.c>
    Listen 127.0.0.1:8443
</IfModule>
PORTS

a2dissite 000-default 2>/dev/null || true
a2ensite redirect
a2enmod rewrite ssl proxy proxy_http
systemctl restart apache2 2>/dev/null || apachectl restart

rm -f /etc/nginx/sites-enabled/default
install -d /etc/nginx/stream-enabled
sed "s/__SYNCLOUD_DOMAIN__/$SYNCLOUD_DOMAIN/g" "$STAGE/common/nginx/stream-relay.conf" > /etc/nginx/stream-enabled/relay.conf
if ! grep -q "stream-enabled" /etc/nginx/nginx.conf; then
    printf '\nstream {\n    include /etc/nginx/stream-enabled/*.conf;\n}\n' >> /etc/nginx/nginx.conf
fi
nginx -t
systemctl reload nginx 2>/dev/null || systemctl restart nginx 2>/dev/null || nginx

FRPS_ADMIN_FILE="$REDIRECT_DIR/frps.admin"
[ -f "$FRPS_ADMIN_FILE" ] || openssl rand -hex 16 > "$FRPS_ADMIN_FILE"
chmod 600 "$FRPS_ADMIN_FILE"
sed -e "s/__FRPS_ADMIN_PASSWORD__/$(cat "$FRPS_ADMIN_FILE")/g" \
    "$STAGE/common/frp/frps.toml" > "$REDIRECT_DIR/frps.toml"
chmod 600 "$REDIRECT_DIR/frps.toml"

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

FRPS_IMAGE=snowdreamtech/frps:0.70.0-alpine
docker pull "$FRPS_IMAGE"
docker rm -f frps 2>/dev/null || true
docker run -d \
    --name frps \
    --restart=unless-stopped \
    --network=host \
    -v "$REDIRECT_DIR/frps.toml:/etc/frp/frps.toml:ro" \
    "$FRPS_IMAGE"

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

for name in redirect-api redirect-www node-exporter frps; do
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
