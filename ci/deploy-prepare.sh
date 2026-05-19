#!/bin/bash
set -ex

if ! command -v ssh >/dev/null; then
    apt-get update
    apt-get install -y openssh-client
fi

KEYFILE=/tmp/_deploy_key
if [ ! -f "$KEYFILE" ]; then
    set +x
    printf '%s\n' "$DEPLOY_KEY" > "$KEYFILE"
    set -x
    chmod 600 "$KEYFILE"
fi

SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no -r"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SSH $REMOTE "sudo -n rm -rf /tmp/syncloud-redirect && mkdir -p /tmp/syncloud-redirect"
$SCP deploy "${REMOTE}:/tmp/syncloud-redirect/"
$SCP db "${REMOTE}:/tmp/syncloud-redirect/"
$SCP config/common "${REMOTE}:/tmp/syncloud-redirect/common"
if [ -d "build/www" ]; then
    $SCP build/www "${REMOTE}:/tmp/syncloud-redirect/web"
fi

if [ -n "$DEPLOY_ENV" ] && [ -d "config/env/$DEPLOY_ENV" ]; then
    STAGE=$(mktemp -d)
    cp -r "config/env/$DEPLOY_ENV/." "$STAGE/"
    if [ -f "$STAGE/secret.cfg" ]; then
        for v in access_key_id secret_access_key hosted_zone_id; do
            val=$(eval echo "\$$v")
            if [ -n "$val" ]; then
                sed -i "s#@$v@#$val#g" "$STAGE/secret.cfg"
            fi
        done
    fi
    $SCP "$STAGE" "${REMOTE}:/tmp/syncloud-redirect/config"
    rm -rf "$STAGE"
fi
