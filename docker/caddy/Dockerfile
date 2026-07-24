FROM caddy:2.10-builder AS builder
ENV GOTOOLCHAIN=auto
RUN xcaddy build --with github.com/caddy-dns/route53

FROM caddy:2.10
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
