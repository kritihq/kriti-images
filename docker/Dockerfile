# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies including libwebp-dev for CGO
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    libwebp-dev

# Set working directory
WORKDIR /app

# Copy go mod files
COPY ../go.mod ../go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY .. .

# Build the binary with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o kriti-images main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies including libwebp
RUN apk --no-cache add \
    ca-certificates \
    tzdata \
    libwebp

# Create non-root user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/kriti-images .

# Copy web assets and templates
COPY --from=builder /app/web ./web

# Copy configuration file
COPY --from=builder /app/config.yaml .

# Create assets directory
RUN mkdir -p web/static/assets

# Change ownership to appuser
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/ready || exit 1

# Run the binary
CMD ["./kriti-images"]
