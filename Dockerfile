# ── Build stage ──────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /build/normalizer ./cmd/normalizer

# ── Runtime stage ────────────────────────────────────────────
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

RUN adduser -D -u 1000 appuser
USER appuser

WORKDIR /app

COPY --from=builder /build/normalizer /app/normalizer
COPY config.example.yaml /app/config.example.yaml

EXPOSE 8080

ENTRYPOINT ["/app/normalizer"]
CMD ["--config", "/app/config.yaml"]
