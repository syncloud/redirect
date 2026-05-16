#!/bin/bash
set -ex

if ! command -v curl >/dev/null; then
    apt-get update
    apt-get install -y curl
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
