# 1) Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /build/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --prefer-offline --no-audit

# 2) Build backend
FROM golang:1.26-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /build/server
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.3
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build \
    mkdir -p /out && CGO_ENABLED=0 go build -o /out/server ./cmd/server

# 3) Runtime
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata bash
RUN wget -q -O /tmp/caddy.tar.gz \
    https://github.com/caddyserver/caddy/releases/download/v2.11.4/caddy_2.11.4_linux_amd64.tar.gz \
    && tar -xzf /tmp/caddy.tar.gz -C /usr/bin caddy \
    && chmod +x /usr/bin/caddy \
    && rm /tmp/caddy.tar.gz
WORKDIR /app
COPY --from=backend-builder /out/server /app/bin/server
COPY --from=frontend-builder \
    /build/frontend/node_modules/@hexlet/project-url-shortener-frontend/dist \
    /app/public
COPY --from=backend-builder /build/server/db/migrations /app/db/migrations
COPY --from=backend-builder /go/bin/goose /usr/local/bin/goose
COPY bin/run.sh /app/bin/run.sh
RUN chmod +x /app/bin/run.sh
COPY Caddyfile /etc/caddy/Caddyfile
EXPOSE 80
CMD ["/app/bin/run.sh"]
