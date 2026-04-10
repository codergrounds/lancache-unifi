# ---- Build stage ----
FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/lancache-unifi .

# ---- Runtime stage ----
FROM alpine:3.21

LABEL org.opencontainers.image.source="https://github.com/codergrounds/lancache-unifi"
LABEL org.opencontainers.image.description="Lancache UniFi DNS Sync"

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /build/lancache-unifi /app/lancache-unifi

ENTRYPOINT ["/app/lancache-unifi"]
