FROM caddy:2.11-builder AS builder
ENV GOTOOLCHAIN=auto
COPY caddy-smtp /src/caddy-smtp
RUN xcaddy build \
    --with github.com/caddy-dns/route53 \
    --with github.com/mholt/caddy-l4 \
    --with github.com/syncloud/caddy-smtp=/src/caddy-smtp

FROM caddy:2.11
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
