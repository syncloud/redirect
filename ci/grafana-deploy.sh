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

BODY=$(python3 - "${DS_UID}" <<'EOF'
import json, sys
ds_uid = sys.argv[1]
with open('/tmp/redirect-v2-dashboard.json') as f:
    raw = f.read()
raw = raw.replace('${DS_PROMETHEUS}', ds_uid)
d = json.loads(raw)
d.pop('__inputs', None)
d.pop('id', None)
print(json.dumps({'dashboard': d, 'overwrite': True, 'folderId': 0, 'message': 'CI auto-deploy'}))
EOF
)
ok=0
for i in 1 2 3 4 5; do
    if curl -fsS -u "${USER}:${PASS}" -X POST -H 'Content-Type: application/json' --data "$BODY" http://127.0.0.1:3000/api/dashboards/db; then ok=1; break; fi
    echo "grafana dashboard upload failed (attempt $i), retrying..." >&2
    sleep 3
done
[ "$ok" = 1 ] || { echo "grafana dashboard upload failed after retries" >&2; exit 1; }
echo
REMOTE_SCRIPT
