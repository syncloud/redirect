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
$SCP config/common "${REMOTE}:/tmp/syncloud-redirect/common"
$SCP build/www "${REMOTE}:/tmp/syncloud-redirect/web"

$SCP "config/env/${DEPLOY_ENV:?DEPLOY_ENV is required}/config.cfg" "${REMOTE}:/tmp/syncloud-redirect/config.cfg"

set +x
PAYMENTS="[paypal]
plan_monthly_id = ${PAYPAL_PLAN_MONTHLY_ID:?PAYPAL_PLAN_MONTHLY_ID is required}
plan_annual_id = ${PAYPAL_PLAN_ANNUAL_ID:?PAYPAL_PLAN_ANNUAL_ID is required}
plan_max_monthly_id = ${PAYPAL_PLAN_MAX_MONTHLY_ID:-}
plan_max_annual_id = ${PAYPAL_PLAN_MAX_ANNUAL_ID:-}
client_id = ${PAYPAL_CLIENT_ID:?PAYPAL_CLIENT_ID is required}
secret_id = ${PAYPAL_SECRET_ID:?PAYPAL_SECRET_ID is required}
url = ${PAYPAL_URL:?PAYPAL_URL is required}
sdk_url = ${PAYPAL_SDK_URL:-}

[stripe]
secret_key = ${STRIPE_SECRET_KEY:?STRIPE_SECRET_KEY is required}
url = ${STRIPE_URL:-}
price_monthly_id = ${STRIPE_PRICE_MONTHLY_ID:?STRIPE_PRICE_MONTHLY_ID is required}
price_annual_id = ${STRIPE_PRICE_ANNUAL_ID:?STRIPE_PRICE_ANNUAL_ID is required}
price_max_monthly_id = ${STRIPE_PRICE_MAX_MONTHLY_ID:-}
price_max_annual_id = ${STRIPE_PRICE_MAX_ANNUAL_ID:-}
"
printf '%s' "$PAYMENTS" | $SSH $REMOTE "sudo -n tee /var/www/redirect/payments.cfg >/dev/null"
set -x
$SSH $REMOTE "sudo -n chmod 600 /var/www/redirect/payments.cfg"
