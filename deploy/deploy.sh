#!/bin/bash
set -ex

if [ "$#" -ne 1 ]; then
    echo "usage: $0 <docker-tag>" >&2
    exit 1
fi

TAG=$1
REDIRECT_DIR=/var/www/redirect
IMAGE_NAME=syncloud/redirect
STAGE=/tmp/syncloud-redirect

if ! command -v docker >/dev/null 2>&1; then
    apt-get update
    apt-get install -y docker.io
fi

if ! command -v apache2 >/dev/null 2>&1 || ! command -v openssl >/dev/null 2>&1 || ! command -v mysql >/dev/null 2>&1; then
    apt-get update
    apt-get install -y --no-install-recommends apache2 openssl default-mysql-client
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

STAGED_CONFIG=$STAGE/config
if [ -d "$STAGED_CONFIG" ]; then
    for f in config.cfg secret.cfg; do
        if [ -f "$STAGED_CONFIG/$f" ]; then
            install -o "$REDIRECT_UID" -g "$REDIRECT_GID" -m 0640 "$STAGED_CONFIG/$f" "$REDIRECT_DIR/$f"
        fi
    done
fi

if [ -d "$STAGE/web" ]; then
    WEB_TARGET=$REDIRECT_DIR/current/www
    mkdir -p "$REDIRECT_DIR/current"
    rm -rf "$WEB_TARGET"
    cp -r "$STAGE/web" "$WEB_TARGET"
    chown -R "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR/current"
fi

cfg_get() {
    local section=$1 key=$2
    awk -v s="[$section]" -v k="$key" '
        $0 == s { in_s = 1; next }
        /^\[/ { in_s = 0 }
        in_s && $1 == k { sub(/^[^=]*=[[:space:]]*/, ""); print; exit }
    ' "$REDIRECT_DIR/config.cfg"
}

SYNCLOUD_DOMAIN=$(cfg_get redirect domain)
DB_HOST=$(cfg_get mysql host)
DB_USER=$(cfg_get mysql user)
DB_PASS=$(cfg_get mysql passwd)
DB_NAME=$(cfg_get mysql db)
: "${DB_NAME:=redirect}"

if [ ! -f "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN/fullchain.pem" ]; then
    mkdir -p "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN"
    openssl req -x509 -newkey rsa:4096 \
        -keyout "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN/privkey.pem" \
        -out "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN/fullchain.pem" \
        -nodes -days 365 \
        -subj "/CN=$SYNCLOUD_DOMAIN"
fi

if [ -f "$STAGE/common/apache/redirect.conf" ]; then
    install -m 0644 "$STAGE/common/apache/redirect.conf" /etc/apache2/sites-available/redirect.conf
    if ! grep -q "^export SYNCLOUD_DOMAIN=" /etc/apache2/envvars; then
        echo "export SYNCLOUD_DOMAIN=$SYNCLOUD_DOMAIN" >> /etc/apache2/envvars
    fi
    a2query -s 000-default >/dev/null 2>&1 && a2dissite 000-default
    a2query -s redirect >/dev/null 2>&1 || a2ensite redirect
    a2enmod rewrite ssl proxy proxy_http >/dev/null
    if systemctl is-active --quiet apache2; then
        systemctl restart apache2
    else
        systemctl start apache2 2>/dev/null || apachectl start
    fi
fi

if [ -n "$DB_HOST" ] && [ -d "$STAGE/db" ]; then
    if ! mysql --host="$DB_HOST" --user="$DB_USER" --password="$DB_PASS" -e "use $DB_NAME" 2>/dev/null; then
        mysql --host="$DB_HOST" --user="$DB_USER" --password="$DB_PASS" -e "CREATE DATABASE $DB_NAME"
        mysql --host="$DB_HOST" --user="$DB_USER" --password="$DB_PASS" "$DB_NAME" < "$STAGE/db/init.sql"
        if [ -f "$STAGE/db/update.sql" ]; then
            mysql --host="$DB_HOST" --user="$DB_USER" --password="$DB_PASS" "$DB_NAME" < "$STAGE/db/update.sql"
        fi
    fi
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
