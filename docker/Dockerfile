# Multi-stage build for API Direct CLI
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY cli/go.mod cli/go.sum ./
RUN go mod download

# Copy source code
COPY cli/ .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o apidirect main.go

# Final stage - minimal image
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    git \
    docker-cli \
    curl \
    bash \
    jq

# Create non-root user
RUN adduser -D -g '' apidirect

# Copy binary from builder
COPY --from=builder /build/apidirect /usr/local/bin/
RUN chmod +x /usr/local/bin/apidirect

# Create working directory
RUN mkdir -p /workspace && chown apidirect:apidirect /workspace
WORKDIR /workspace

# Switch to non-root user
USER apidirect

# Set environment variables
ENV APIDIRECT_HOME=/home/apidirect/.apidirect
ENV PATH="/usr/local/bin:${PATH}"

# Verify installation
RUN apidirect --version

# Default command
ENTRYPOINT ["apidirect"]
CMD ["--help"]