# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.26-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Download dependencies first (cached layer)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/main.go

# ── Stage 2: Production ───────────────────────────────────────────────────────
FROM alpine:3.20

WORKDIR /app

# CA certs for outbound HTTPS (Paystack, S3, SMTP)
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/server      ./server
COPY --from=builder /app/migrations  ./migrations
COPY --from=builder /app/templates   ./templates

EXPOSE 8080

ENTRYPOINT ["./server"]
