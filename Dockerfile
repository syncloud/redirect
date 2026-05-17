FROM golang:1.20.4-buster AS builder
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/api ./cmd/api \
 && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/www ./cmd/www \
 && CGO_ENABLED=0 go build -ldflags="-s -w" -o /out/cli ./cmd/cli

FROM gcr.io/distroless/static-debian12
COPY --from=builder /out/api /usr/local/bin/api
COPY --from=builder /out/www /usr/local/bin/www
COPY --from=builder /out/cli /usr/local/bin/cli
COPY emails /app/emails
