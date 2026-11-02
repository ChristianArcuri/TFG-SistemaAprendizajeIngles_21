# syntax=docker/dockerfile:1.6

# ===== Build stage =====
FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/api ./cmd/api

# ===== Runtime stage =====
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app

WORKDIR /app
COPY --from=builder /out/api /app/api

USER app

EXPOSE 8080

ENTRYPOINT ["/app/api"]
