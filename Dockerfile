FROM gcr.io/distroless/static-debian12
COPY build/bin/api /usr/local/bin/api
COPY build/bin/www /usr/local/bin/www
COPY build/bin/cli /usr/local/bin/cli
COPY emails /app/emails
