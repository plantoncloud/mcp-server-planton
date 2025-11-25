# Multi-stage Dockerfile for mcp-server-planton
# Follows GitHub's MCP server Docker distribution approach

# Build stage
FROM golang:1.23-alpine AS builder

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

# Install ca-certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mcp-server-planton .

# Copy README and LICENSE for reference
COPY README.md LICENSE ./

# Set the entrypoint
ENTRYPOINT ["./mcp-server-planton"]

