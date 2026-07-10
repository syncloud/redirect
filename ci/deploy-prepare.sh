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
$SCP bin "${REMOTE}:/tmp/syncloud-redirect/"
$SCP db "${REMOTE}:/tmp/syncloud-redirect/"
$SCP config/common "${REMOTE}:/tmp/syncloud-redirect/common"
$SCP build/www "${REMOTE}:/tmp/syncloud-redirect/web"

if [ -n "${PAYPAL_CLIENT_ID:-}" ]; then
    set +x
    PAYMENTS="[paypal]
plan_monthly_id = ${PAYPAL_PLAN_MONTHLY_ID}
plan_annual_id = ${PAYPAL_PLAN_ANNUAL_ID}
client_id = ${PAYPAL_CLIENT_ID}
secret_id = ${PAYPAL_SECRET_ID}
url = ${PAYPAL_URL}
"
    if [ -n "${STRIPE_SECRET_KEY:-}" ]; then
        PAYMENTS="${PAYMENTS}
[stripe]
secret_key = ${STRIPE_SECRET_KEY}
price_monthly_id = ${STRIPE_PRICE_MONTHLY_ID}
price_annual_id = ${STRIPE_PRICE_ANNUAL_ID}
"
    fi
    printf '%s' "$PAYMENTS" | $SSH $REMOTE "sudo -n tee /var/www/redirect/payments.cfg >/dev/null"
    set -x
    $SSH $REMOTE "sudo -n chmod 600 /var/www/redirect/payments.cfg"
fi
