#!/bin/bash
set -ex

# Test-env-only bits that deploy.sh can't cover on a fresh host:
# 1. /var/www/redirect/{config,secret}.cfg from the in-repo template
#    (with @secret@ placeholders substituted from drone secrets).
#    UAT/prod have these on-disk from one-time provisioning long ago.
# 2. A self-signed SSL cert at /etc/letsencrypt/live/$SYNCLOUD_DOMAIN/.
#    UAT/prod use letsencrypt.

if [ -z "$DEPLOY_ENV" ] || [ ! -d "config/env/$DEPLOY_ENV" ]; then
    echo "DEPLOY_ENV must be set to a dir under config/env/" >&2
    exit 1
fi

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no -r"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

STAGE_LOCAL=$(mktemp -d)
trap 'rm -rf "$STAGE_LOCAL"' EXIT
cp -r "config/env/$DEPLOY_ENV/." "$STAGE_LOCAL/"
sed -i "s#@access_key_id@#$access_key_id#g"         "$STAGE_LOCAL/secret.cfg"
sed -i "s#@secret_access_key@#$secret_access_key#g" "$STAGE_LOCAL/secret.cfg"
sed -i "s#@hosted_zone_id@#$hosted_zone_id#g"       "$STAGE_LOCAL/secret.cfg"

$SSH $REMOTE "sudo -n rm -rf /tmp/syncloud-redirect-setup && sudo -n mkdir -p /tmp/syncloud-redirect-setup/config"
$SCP "$STAGE_LOCAL/." "${REMOTE}:/tmp/syncloud-redirect-setup/config/"

$SSH $REMOTE sudo -n SYNCLOUD_DOMAIN="$SYNCLOUD_DOMAIN" bash -s <<'REMOTE_SCRIPT'
set -ex
REDIRECT_DIR=/var/www/redirect
STAGE=/tmp/syncloud-redirect-setup

apt-get update
apt-get install -y --no-install-recommends openssl

adduser --disabled-password --gecos "" redirect
REDIRECT_UID=$(id -u redirect)
REDIRECT_GID=$(id -g redirect)

mkdir -p "$REDIRECT_DIR"
chown "$REDIRECT_UID:$REDIRECT_GID" "$REDIRECT_DIR"
install -o "$REDIRECT_UID" -g "$REDIRECT_GID" -m 0640 "$STAGE/config/config.cfg" "$REDIRECT_DIR/config.cfg"
install -o "$REDIRECT_UID" -g "$REDIRECT_GID" -m 0640 "$STAGE/config/secret.cfg" "$REDIRECT_DIR/secret.cfg"

mkdir -p "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN"
openssl req -x509 -newkey rsa:4096 \
    -keyout "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN/privkey.pem" \
    -out "/etc/letsencrypt/live/$SYNCLOUD_DOMAIN/fullchain.pem" \
    -nodes -days 365 \
    -subj "/CN=$SYNCLOUD_DOMAIN"
REMOTE_SCRIPT
