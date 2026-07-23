FROM caddy:2.10-builder AS builder
ENV GOTOOLCHAIN=auto
RUN xcaddy build --with github.com/caddy-dns/route53

FROM caddy:2.10
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
ADD https://raw.githubusercontent.com/letsencrypt/pebble/v2.6.0/test/certs/pebble.minica.pem /usr/local/share/ca-certificates/pebble.crt
RUN cat /usr/local/share/ca-certificates/pebble.crt >> /etc/ssl/certs/ca-certificates.crt
