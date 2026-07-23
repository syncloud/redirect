#!/bin/bash
set -ex

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
apt-get install -y --no-install-recommends curl docker.io

mkdir -p "$REDIRECT_DIR"
install -m 0640 "$STAGE/config/config.cfg" "$REDIRECT_DIR/config.cfg"
install -m 0640 "$STAGE/config/secret.cfg" "$REDIRECT_DIR/secret.cfg"

if ! docker info >/dev/null 2>&1; then
    ( dockerd --storage-driver=vfs </dev/null >/var/log/dockerd.log 2>&1 & )
fi
for i in $(seq 1 30); do docker info >/dev/null 2>&1 && break; sleep 1; done

docker rm -f localstack pebble coredns 2>/dev/null || true

install -d /tmp/simdns
cat > /tmp/simdns/Corefile <<'COREFILE'
test.:53 {
    file /zones/test.zone {
        reload 1s
    }
    errors
}
COREFILE
cat > /tmp/simdns/test.zone <<'ZONE'
$ORIGIN test.
$TTL 60
@ IN SOA ns.test. admin.test. 1 7200 3600 1209600 60
@ IN NS ns.test.
ns IN A 127.0.0.1
ZONE

docker run -d --name coredns --network=host -v /tmp/simdns:/zones coredns/coredns:1.11.1 -conf /zones/Corefile
docker run -d --name localstack --network=host -v /tmp/simdns:/zones -e SERVICES=route53 localstack/localstack:3
docker run -d --name pebble --network=host ghcr.io/letsencrypt/pebble:2.6.0 -dnsserver 127.0.0.1:53

for i in $(seq 1 60); do curl -sf http://localhost:4566/_localstack/health >/dev/null 2>&1 && break; sleep 2; done
docker run --rm --network=host -e AWS_ACCESS_KEY_ID=test -e AWS_SECRET_ACCESS_KEY=test -e AWS_DEFAULT_REGION=us-east-1 \
    amazon/aws-cli --endpoint-url http://localhost:4566 route53 create-hosted-zone --name test --caller-reference ci

cat > /tmp/simdns/poller.sh <<'POLLER'
#!/bin/sh
while true; do
  ZID=$(awslocal route53 list-hosted-zones --query "HostedZones[0].Id" --output text 2>/dev/null | sed "s#/hostedzone/##")
  {
    echo '$ORIGIN test.'
    echo '$TTL 60'
    echo "@ IN SOA ns.test. admin.test. $(date +%s) 7200 3600 1209600 60"
    echo '@ IN NS ns.test.'
    echo 'ns IN A 127.0.0.1'
    [ -n "$ZID" ] && awslocal route53 list-resource-record-sets --hosted-zone-id "$ZID" --output json 2>/dev/null | python3 -c '
import json,sys
try: rs=json.load(sys.stdin).get("ResourceRecordSets",[])
except Exception: rs=[]
for r in rs:
    if r.get("Type")!="TXT": continue
    for v in r.get("ResourceRecords",[]):
        print(r["Name"], "IN TXT", v["Value"])
'
  } > /zones/test.zone.tmp && mv /zones/test.zone.tmp /zones/test.zone
  sleep 2
done
POLLER
docker exec -d localstack sh /zones/poller.sh
REMOTE_SCRIPT
