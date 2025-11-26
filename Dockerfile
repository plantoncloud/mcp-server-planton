# Multi-stage Dockerfile for mcp-server-planton
# Follows GitHub's MCP server Docker distribution approach

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
# GOOS=linux for Linux container
RUN CGO_ENABLED=0 GOOS=linux go build -o mcp-server-planton ./cmd/mcp-server-planton

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections and wget for health checks
RUN apk --no-cache add ca-certificates wget

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mcp-server-planton .

# Copy README and LICENSE for reference
COPY README.md LICENSE ./

# Expose HTTP port (only used when PLANTON_MCP_TRANSPORT=http or both)
EXPOSE 8080

# Health check for HTTP mode
# This will only work when the server is running in HTTP or both mode
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Set the entrypoint
ENTRYPOINT ["./mcp-server-planton"]

