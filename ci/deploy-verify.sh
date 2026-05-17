#!/bin/bash
set -ex

if ! command -v curl >/dev/null; then
    apt-get update
    apt-get install -y curl
fi

URL_HOST=$(echo "$DEPLOY_URL" | sed -E 's|https?://([^/:]+).*|\1|')
if ! getent hosts "$URL_HOST" >/dev/null 2>&1; then
    IP=$(getent hosts "$DEPLOY_HOST" | awk '{print $1}')
    if [ -n "$IP" ]; then
        echo "$IP $URL_HOST" >> /etc/hosts
    fi
fi

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

echo "=== post-migration host state ==="
$SSH $REMOTE "ls -la /var/www/redirect/ /var/www/redirect/current/ /var/www/redirect/current/bin/ 2>&1 | head -40" || true
$SSH $REMOTE "readlink /var/www/redirect/current; which mysqldump" || true
