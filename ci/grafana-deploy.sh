#!/bin/bash
set -ex

KEYFILE=/tmp/_deploy_key
SSH="ssh -i $KEYFILE -o StrictHostKeyChecking=no"
SCP="scp -i $KEYFILE -o StrictHostKeyChecking=no"
REMOTE="${DEPLOY_USER}@${DEPLOY_HOST}"

$SCP monitoring/grafana/redirect-v2.json "${REMOTE}:/tmp/redirect-v2-dashboard.json"

$SSH $REMOTE 'sudo bash -s' <<'REMOTE_SCRIPT'
set -e
USER=$(awk -F= '/^[[:space:]]*admin_user[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' /etc/grafana/grafana.ini | head -1)
PASS=$(awk -F= '/^[[:space:]]*admin_password[[:space:]]*=/{gsub(/^[[:space:]]+|[[:space:]]+$/, "", $2); print $2}' /etc/grafana/grafana.ini | head -1)
DS_UID=$(curl -s -u "${USER}:${PASS}" http://127.0.0.1:3000/api/datasources \
  | python3 -c "import json,sys; print(next(d['uid'] for d in json.load(sys.stdin) if d['type']=='prometheus'))")

python3 <<EOF
import json, os
with open('/tmp/redirect-v2-dashboard.json') as f:
    raw = f.read()
raw = raw.replace('\${DS_PROMETHEUS}', '${DS_UID}')
d = json.loads(raw)
d.pop('__inputs', None)
d.pop('id', None)
payload = {'dashboard': d, 'overwrite': True, 'folderId': 0, 'message': 'CI auto-deploy'}
with open('/tmp/import.json', 'w') as f:
    json.dump(payload, f)
EOF

curl -fsS -u "${USER}:${PASS}" -X POST -H 'Content-Type: application/json' \
  --data @/tmp/import.json http://127.0.0.1:3000/api/dashboards/db
echo
REMOTE_SCRIPT
