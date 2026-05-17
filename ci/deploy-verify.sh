#!/bin/bash
set -ex

if ! command -v curl >/dev/null; then
    apt-get update
    apt-get install -y curl
fi

URL_HOST=$(echo "$DEPLOY_URL" | sed -E 's|https?://([^/:]+).*|\1|')
WWW_URL=$(echo "$DEPLOY_URL" | sed -E 's|//api\.|//www.|')
WWW_HOST=$(echo "$WWW_URL" | sed -E 's|https?://([^/:]+).*|\1|')

resolve_alias() {
    local host=$1
    if getent hosts "$host" >/dev/null 2>&1; then return; fi
    local ip
    ip=$(getent hosts "$DEPLOY_HOST" | awk '{print $1}')
    [ -n "$ip" ] && echo "$ip $host" >> /etc/hosts
}
resolve_alias "$URL_HOST"
resolve_alias "$WWW_HOST"

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

for i in $(seq 1 60); do
    code=$(curl -k -s -o /dev/null -w "%{http_code}" "${DEPLOY_URL}/status" || echo 000)
    if [ "$code" = "200" ]; then
        echo "OK ($code)"
        break
    fi
    sleep 2
done
if [ "$code" != "200" ]; then
    echo "redirect did not come up: last http_code=$code"
    $SSH $REMOTE sudo -n docker ps -a 2>&1 || true
    $SSH $REMOTE sudo -n docker logs redirect-api 2>&1 | tail -40 || true
    $SSH $REMOTE sudo -n docker logs redirect-www 2>&1 | tail -40 || true
    exit 1
fi

web_code=$(curl -k -s -o /dev/null -w "%{http_code}" "${WWW_URL}/")
if [ "$web_code" != "200" ]; then
    echo "web UI did not respond at ${WWW_URL}/: http_code=$web_code"
    exit 1
fi
echo "web UI OK ($web_code)"

if [ -n "${SMOKE_TOKEN:-}" ]; then
    body=$(curl -k -s -X POST "${DEPLOY_URL}/domain/update" \
        -H 'Content-Type: application/json' \
        -d "{\"token\":\"${SMOKE_TOKEN}\",\"ipv4_enabled\":true}")
    echo "smoke /domain/update response: $body"
    if ! echo "$body" | grep -q '"success":true'; then
        echo "smoke /domain/update failed"
        exit 1
    fi
fi
