FROM caddy:2.11-builder AS builder
ENV GOTOOLCHAIN=auto
RUN xcaddy build \
    --with github.com/caddy-dns/route53 \
    --with github.com/mholt/caddy-l4

FROM caddy:2.11
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
