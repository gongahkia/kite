# Multi-stage Dockerfile for Kite v4

# Stage 1: Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o kite-api \
    ./cmd/kite-api

# Stage 2: Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 kite && \
    adduser -D -u 1000 -G kite kite

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/kite-api .

# Copy configuration files
COPY configs/ ./configs/

# Set ownership
RUN chown -R kite:kite /app

# Switch to non-root user
USER kite

# Expose ports
EXPOSE 8080 9091

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
ENTRYPOINT ["./kite-api"]
