#!/bin/bash
set -e

REDIRECT_DIR=/var/www/redirect
CADDY_DATA=$(docker volume inspect caddy_data --format '{{.Mountpoint}}' 2>/dev/null)
DOMAIN=$1

if [ -z "$DOMAIN" ]; then
    echo "usage: $0 <redirect-domain>" >&2
    exit 1
fi

if [ -z "$CADDY_DATA" ] || [ ! -d "$CADDY_DATA" ]; then
    echo "caddy_data volume is not available" >&2
    exit 1
fi

# the issuer directory differs between the staging and production acme
# endpoints, so find the certificate rather than assume where it lives
CERT=$(find "$CADDY_DATA/caddy/certificates" \
    -name "wildcard_.mx.$DOMAIN.crt" -type f 2>/dev/null | head -1)
KEY=${CERT%.crt}.key

if [ -z "$CERT" ] || [ ! -f "$KEY" ]; then
    echo "no certificate for *.mx.$DOMAIN yet" >&2
    exit 0
fi

DEST_CERT=$REDIRECT_DIR/mx.crt
DEST_KEY=$REDIRECT_DIR/mx.key

if cmp -s "$CERT" "$DEST_CERT" && cmp -s "$KEY" "$DEST_KEY"; then
    exit 0
fi

install -m 0644 -o redirect -g redirect "$CERT" "$DEST_CERT"
install -m 0640 -o redirect -g redirect "$KEY" "$DEST_KEY"
echo "copied the *.mx.$DOMAIN certificate"
