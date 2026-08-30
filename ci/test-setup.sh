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

$SSH $REMOTE "sudo -n rm -rf /tmp/syncloud-redirect-setup && sudo -n mkdir -p /tmp/syncloud-redirect-setup/config /tmp/syncloud-redirect-setup/sim"
$SCP "$STAGE_LOCAL/." "${REMOTE}:/tmp/syncloud-redirect-setup/config/"
$SCP "ci/sim/." "${REMOTE}:/tmp/syncloud-redirect-setup/sim/"

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

docker rm -f pebble 2>/dev/null || true
pkill -f /usr/local/bin/dns-faker 2>/dev/null || true
pkill -f /usr/local/bin/ses-faker 2>/dev/null || true
pkill -f /usr/local/bin/payment-faker 2>/dev/null || true
pkill -f /usr/local/bin/device-faker 2>/dev/null || true
pkill -f /usr/local/bin/frpc 2>/dev/null || true

install -m 0755 "$STAGE/sim/dns-faker" /usr/local/bin/dns-faker
( /usr/local/bin/dns-faker </dev/null >/var/log/dns-faker.log 2>&1 & )
for i in $(seq 1 30); do curl -sf http://localhost:4566/health >/dev/null 2>&1 && break; sleep 1; done

install -m 0755 "$STAGE/sim/payment-faker" /usr/local/bin/payment-faker
( /usr/local/bin/payment-faker \
    --paypal 127.0.0.1:4581 \
    --stripe 127.0.0.1:4582 \
    --stripe-url "https://payments.$SYNCLOUD_DOMAIN/stripe" \
    </dev/null >/var/log/payment-faker.log 2>&1 & )
for i in $(seq 1 30); do curl -sf http://127.0.0.1:4581/sdk/js >/dev/null 2>&1 && break; sleep 1; done
curl -sf http://127.0.0.1:4581/sdk/js >/dev/null

install -d /etc/caddy/conf.d
cat > /etc/caddy/conf.d/payment-faker.caddy <<FAKER
payments.$SYNCLOUD_DOMAIN {
	import syncloud_tls
	handle_path /paypal/* {
		reverse_proxy 127.0.0.1:4581
	}
	handle_path /stripe/* {
		reverse_proxy 127.0.0.1:4582
	}
}
FAKER

install -m 0755 "$STAGE/sim/ses-faker" /usr/local/bin/ses-faker
( /usr/local/bin/ses-faker </dev/null >/var/log/ses-faker.log 2>&1 & )
for i in $(seq 1 30); do curl -sf http://localhost:4579/faker/messages >/dev/null 2>&1 && break; sleep 1; done

docker run -d --name pebble --network=host ghcr.io/letsencrypt/pebble:2.6.0 -dnsserver 127.0.0.1:53
docker cp pebble:/test/certs/pebble.minica.pem "$REDIRECT_DIR/pebble.crt"

FRP_VERSION=0.70.0
curl -sfL "https://github.com/fatedier/frp/releases/download/v${FRP_VERSION}/frp_${FRP_VERSION}_linux_amd64.tar.gz" -o /tmp/frp.tgz
tar -xzf /tmp/frp.tgz -C /tmp
install -m 0755 "/tmp/frp_${FRP_VERSION}_linux_amd64/frpc" /usr/local/bin/frpc

rm -rf /var/lib/device-faker
mkdir -p /var/lib/device-faker
install -m 0755 "$STAGE/sim/device-faker" /usr/local/bin/device-faker
( /usr/local/bin/device-faker \
    --smtp 127.0.0.1:2525 \
    --api :4580 \
    --frpc /usr/local/bin/frpc \
    --server-addr "www.$SYNCLOUD_DOMAIN" \
    --server-name "relay.$SYNCLOUD_DOMAIN" \
    --work-dir /var/lib/device-faker </dev/null >/var/log/device-faker.log 2>&1 & )
for i in $(seq 1 30); do curl -sf http://localhost:4580/faker/messages >/dev/null 2>&1 && break; sleep 1; done
REMOTE_SCRIPT
