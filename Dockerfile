# ---------- frontend ----------
FROM node:22-alpine AS web
WORKDIR /web
COPY frontend/package*.json ./
RUN npm install --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# ---------- backend ----------
FROM golang:1.27-alpine AS build
WORKDIR /src
RUN apk add --no-cache git ca-certificates
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# ---------- runtime ----------
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S app && adduser -S -G app app
WORKDIR /app
COPY --from=build /out/api /app/api
COPY --from=web /web/dist /app/web

ENV APP_ENV=production \
    HTTP_ADDR=:8080 \
    WEB_DIR=/app/web \
    FILES_DIR=/data/files \
    RUN_MIGRATIONS=true

RUN mkdir -p /data/files && chown -R app:app /data/files
USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=10s \
    CMD wget -qO- http://127.0.0.1:8080/api/health || exit 1

ENTRYPOINT ["/app/api"]
