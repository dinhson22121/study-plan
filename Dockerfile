# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.23-alpine AS builder

WORKDIR /src
RUN apk add --no-cache git ca-certificates

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build the server and the migration runner as static binaries.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/server  ./cmd/server \
 && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

# ---- Runtime stage ----
FROM alpine:3.20

# ca-certificates: outbound TLS (FCM). tzdata: timezone-aware scheduler
# (Asia/Ho_Chi_Minh). wget (busybox) is used by the compose healthcheck.
RUN apk add --no-cache ca-certificates tzdata \
 && adduser -D -u 10001 app

WORKDIR /app
COPY --from=builder /out/server  /app/server
COPY --from=builder /out/migrate /app/migrate
COPY config/config.yaml /app/config/config.yaml
COPY migrations /app/migrations

USER app
EXPOSE 8080

# Override config via EDU_* env vars (see docker-compose.yml / README).
ENTRYPOINT ["/app/server"]
