#!/bin/bash
set -ex

# Provision the test host so deploy.sh runs against the same shape as UAT/prod:
# apache2 + SSL cert + redirect.conf + mysql schema + config.cfg/secret.cfg
# already in place. Real envs have all of this from prior installs; this script
# is the test-env equivalent and must not run against UAT/prod.

if [ -z "$DEPLOY_ENV" ] || [ ! -d "config/env/$DEPLOY_ENV" ]; then
    echo "DEPLOY_ENV must be set to a dir under config/env/ (e.g. integration)" >&2
    exit 1
fi

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no -r"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

STAGE_LOCAL=$(mktemp -d)
trap 'rm -rf "$STAGE_LOCAL"' EXIT
cp -r "config/env/$DEPLOY_ENV/." "$STAGE_LOCAL/"
for v in access_key_id secret_access_key hosted_zone_id; do
    val=$(eval echo "\$$v")
    if [ -n "$val" ] && [ -f "$STAGE_LOCAL/secret.cfg" ]; then
        sed -i "s#@$v@#$val#g" "$STAGE_LOCAL/secret.cfg"
    fi
done

$SSH $REMOTE "sudo -n mkdir -p /tmp/syncloud-redirect-setup"
$SCP "$STAGE_LOCAL/." "${REMOTE}:/tmp/syncloud-redirect-setup/config/"
$SCP config/common "${REMOTE}:/tmp/syncloud-redirect-setup/common"
$SCP db "${REMOTE}:/tmp/syncloud-redirect-setup/db"

DOMAIN=$(awk -v s='[redirect]' -v k=domain '
    $0 == s { in_s = 1; next }
    /^\[/ { in_s = 0 }
    in_s && $1 == k { sub(/^[^=]*=[[:space:]]*/, ""); print; exit }
' "$STAGE_LOCAL/config.cfg")
DB_HOST=$(awk -v s='[mysql]' -v k=host '
    $0 == s { in_s = 1; next }
    /^\[/ { in_s = 0 }
    in_s && $1 == k { sub(/^[^=]*=[[:space:]]*/, ""); print; exit }
' "$STAGE_LOCAL/config.cfg")

$SSH $REMOTE sudo -n DOMAIN="$DOMAIN" DB_HOST="$DB_HOST" bash -s <<'REMOTE_SCRIPT'
set -ex
REDIRECT_DIR=/var/www/redirect
STAGE=/tmp/syncloud-redirect-setup

apt-get update
apt-get install -y --no-install-recommends apache2 openssl default-mysql-client

if ! id -u redirect >/dev/null 2>&1; then
    adduser --disabled-password --gecos "" redirect
fi
REDIRECT_UID=$(id -u redirect)
REDIRECT_GID=$(id -g redirect)

mkdir -p "$REDIRECT_DIR"
chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR"
install -o "$REDIRECT_UID" -g "$REDIRECT_GID" -m 0640 "$STAGE/config/config.cfg" "$REDIRECT_DIR/config.cfg"
install -o "$REDIRECT_UID" -g "$REDIRECT_GID" -m 0640 "$STAGE/config/secret.cfg" "$REDIRECT_DIR/secret.cfg"

if [ ! -f "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" ]; then
    mkdir -p "/etc/letsencrypt/live/$DOMAIN"
    openssl req -x509 -newkey rsa:4096 \
        -keyout "/etc/letsencrypt/live/$DOMAIN/privkey.pem" \
        -out "/etc/letsencrypt/live/$DOMAIN/fullchain.pem" \
        -nodes -days 365 \
        -subj "/CN=$DOMAIN"
fi

install -m 0644 "$STAGE/common/apache/redirect.conf" /etc/apache2/sites-available/redirect.conf
if ! grep -q "^export SYNCLOUD_DOMAIN=" /etc/apache2/envvars; then
    echo "export SYNCLOUD_DOMAIN=$DOMAIN" >> /etc/apache2/envvars
fi
a2query -s 000-default >/dev/null 2>&1 && a2dissite 000-default
a2query -s redirect >/dev/null 2>&1 || a2ensite redirect
a2enmod rewrite ssl proxy proxy_http >/dev/null
if systemctl is-active --quiet apache2; then
    systemctl restart apache2
else
    systemctl start apache2 2>/dev/null || apachectl start
fi

if ! mysql --host="$DB_HOST" --user=root --password=root -e "use redirect" 2>/dev/null; then
    mysql --host="$DB_HOST" --user=root --password=root -e "CREATE DATABASE redirect"
    mysql --host="$DB_HOST" --user=root --password=root redirect < "$STAGE/db/init.sql"
    if [ -f "$STAGE/db/update.sql" ]; then
        mysql --host="$DB_HOST" --user=root --password=root redirect < "$STAGE/db/update.sql"
    fi
fi
REMOTE_SCRIPT
