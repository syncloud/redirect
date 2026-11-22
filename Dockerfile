FROM gcr.io/distroless/static-debian12
COPY build/bin/api /usr/local/bin/api
COPY build/bin/www /usr/local/bin/www
COPY build/bin/cli /usr/local/bin/cli
COPY build/bin/migrate /usr/local/bin/migrate
COPY build/bin/stun /usr/local/bin/stun
COPY emails /app/emails
