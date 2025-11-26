# HTTP Transport Support

The MCP server supports HTTP transport using Server-Sent Events (SSE) for remote access, in addition to the default STDIO transport for local use.

## Transport Modes

The server can run in three modes:

1. **stdio** (default) - Standard input/output for local AI clients (Cursor, Claude Desktop)
2. **http** - HTTP/SSE transport for remote access via URL endpoint
3. **both** - Run both transports simultaneously

## Remote Access (Recommended)

### Using the Hosted Endpoint

The easiest way to use the MCP server is via the hosted endpoint at `https://mcp.planton.ai/`.

**Cursor Configuration:**

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "https://mcp.planton.ai/",
      "headers": {
        "Authorization": "Bearer YOUR_PLANTON_API_KEY"
      }
    }
  }
}
```

Replace `YOUR_PLANTON_API_KEY` with your actual API key from the Planton Cloud console.

**Benefits:**
- No local installation required
- Always up-to-date with latest features
- Managed and monitored by Planton Cloud
- High availability and performance

## Running Locally

For development, testing, or private deployments, you can run the MCP server locally.

### Configuration

Configure HTTP transport using environment variables:

**Required:**
- `PLANTON_API_KEY` - Your Planton Cloud API key

**HTTP Transport:**
- `PLANTON_MCP_TRANSPORT` - Set to `http` or `both` (default: `stdio`)
- `PLANTON_MCP_HTTP_PORT` - HTTP server port (default: `8080`)
- `PLANTON_MCP_HTTP_AUTH_ENABLED` - Enable bearer token auth (default: `true`)

**Note:** When authentication is enabled, `PLANTON_API_KEY` is used as the bearer token.

### Local Setup with Docker

**1. Run the Docker container:**

```bash
docker run -p 8080:8080 \
  -e PLANTON_API_KEY="YOUR_PLANTON_API_KEY" \
  -e PLANTON_MCP_TRANSPORT="http" \
  -e PLANTON_MCP_HTTP_AUTH_ENABLED="true" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

**2. Configure Cursor:**

Add to `~/.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "http://localhost:8080/",
      "headers": {
        "Authorization": "Bearer YOUR_PLANTON_API_KEY"
      }
    }
  }
}
```

**3. Test the connection:**

```bash
# Health check
curl http://localhost:8080/health

# SSE endpoint (with authentication)
curl -H "Authorization: Bearer YOUR_PLANTON_API_KEY" http://localhost:8080/sse
```

### Local Setup with Binary

**1. Install the binary:**

```bash
# macOS (ARM64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Darwin_arm64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Darwin_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Linux_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
```

**2. Start the server:**

```bash
export PLANTON_API_KEY="YOUR_PLANTON_API_KEY"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="true"

mcp-server-planton
```

**3. Configure Cursor:** (same as Docker setup above)

**4. Test the connection:** (same as Docker setup above)

## HTTP Endpoints

When running in HTTP mode, the following endpoints are available:

- `GET /health` - Health check endpoint (returns `{"status":"ok"}`)
- `GET /sse` - SSE connection endpoint for MCP protocol
- `POST /message` - Message endpoint for MCP protocol

All endpoints except `/health` require authentication when `PLANTON_MCP_HTTP_AUTH_ENABLED` is `true`.

## Testing

### Health Check

Test if the server is running:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"ok"}
```

### SSE Connection Test

Test the SSE endpoint with authentication:

```bash
curl -H "Authorization: Bearer YOUR_PLANTON_API_KEY" http://localhost:8080/sse
```

The connection will stay open and stream MCP protocol messages.

### Without Authentication (Local Testing Only)

For local testing without authentication:

```bash
export PLANTON_API_KEY="YOUR_PLANTON_API_KEY"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"
mcp-server-planton
```

Then test without the Authorization header:
```bash
curl http://localhost:8080/sse
```

## Use Cases

### STDIO Transport

- Local AI clients (Claude Desktop, Cursor)
- Development and testing
- Direct process spawning by LangGraph

### HTTP Transport

- Remote access to MCP server
- Webhook integrations
- Cloud deployments (Docker, Kubernetes)
- Serverless functions (AWS Lambda, Cloud Run, Cloudflare Workers)
- Team shared services

### Both Transports

- Development environments needing both local and remote access
- Testing remote clients while maintaining local workflow

## Architecture

The HTTP transport implementation uses the `mark3labs/mcp-go` library's `SSEServer`:

- **Stateless** - No session storage, scales horizontally
- **SSE-based** - Real-time bidirectional communication over HTTP
- **Standard MCP protocol** - Compatible with all MCP clients

## Implementation Details

### Architecture

The HTTP transport uses a reverse proxy architecture to add custom functionality on top of the mcp-go library's SSEServer:

1. **Internal SSE Server**: Runs on localhost:18080, handling the core MCP protocol
2. **Proxy Layer**: Runs on the configured port (default 8080), adds:
   - Health check endpoint at `/health`
   - Optional bearer token authentication
   - Request logging
   - Custom routing

This architecture allows us to enhance the SSEServer without modifying the library, while maintaining full compatibility with the MCP protocol.

### Completed Features

- ✅ Bearer token authentication middleware
- ✅ Health check endpoint (`/health`)
- ✅ Custom HTTP server wrapper
- ✅ Request logging
- ✅ Proper SSE streaming with flushing

### Future Enhancements

- [ ] Metrics endpoint (`/metrics`)
- [ ] Rate limiting middleware
- [ ] TLS/HTTPS support (use reverse proxy like nginx/caddy for now)
- [ ] Configurable CORS policies
- [ ] Connection pooling and timeout configuration

## Security Considerations

### Production Deployments

When deploying HTTP transport in production:

1. **Enable Authentication** - Set `PLANTON_MCP_HTTP_AUTH_ENABLED="true"` (default)
2. **Secure API Key** - Your `PLANTON_API_KEY` serves as the bearer token
3. **TLS/HTTPS** - Always use HTTPS (reverse proxy, load balancer, or hosted endpoint)
4. **Network Isolation** - Use VPCs, security groups, firewalls when self-hosting
5. **API Key Management** - Store `PLANTON_API_KEY` securely (secrets management)

### Security Model

```
Client → PLANTON_API_KEY (Bearer) → MCP Server → PLANTON_API_KEY → Planton APIs
         (Authentication)                          (Authorization & FGA)
```

- **PLANTON_API_KEY as Bearer Token** - Authenticates access to your MCP server instance
- **PLANTON_API_KEY to Planton APIs** - Enforces your actual permissions (Fine-Grained Authorization)

This unified approach simplifies authentication while maintaining security through your user permissions.

## Deployment Examples

### Docker Compose

```yaml
version: '3.8'
services:
  mcp-server:
    image: ghcr.io/plantoncloud-inc/mcp-server-planton:latest
    ports:
      - "8080:8080"
    environment:
      - PLANTON_API_KEY=${PLANTON_API_KEY}
      - PLANTON_MCP_TRANSPORT=http
      - PLANTON_MCP_HTTP_PORT=8080
      - PLANTON_MCP_HTTP_AUTH_ENABLED=true
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mcp-server-planton
spec:
  replicas: 3
  selector:
    matchLabels:
      app: mcp-server-planton
  template:
    metadata:
      labels:
        app: mcp-server-planton
    spec:
      containers:
      - name: mcp-server
        image: ghcr.io/plantoncloud-inc/mcp-server-planton:latest
        ports:
        - containerPort: 8080
        env:
        - name: PLANTON_API_KEY
          valueFrom:
            secretKeyRef:
              name: planton-secrets
              key: api-key
        - name: PLANTON_MCP_TRANSPORT
          value: "http"
        - name: PLANTON_MCP_HTTP_PORT
          value: "8080"
        - name: PLANTON_MCP_HTTP_AUTH_ENABLED
          value: "true"
---
apiVersion: v1
kind: Service
metadata:
  name: mcp-server-planton
spec:
  selector:
    app: mcp-server-planton
  ports:
  - port: 8080
    targetPort: 8080
  type: LoadBalancer
```

**Note:** In production deployments, each user should have their own instance of the MCP server with their own API key, or use the hosted endpoint at `https://mcp.planton.ai/` which handles multi-user authentication automatically.

## Migration Guide

### From STDIO to HTTP

If you're currently using the MCP server with STDIO transport:

1. No changes needed - STDIO remains the default
2. To add HTTP access, set `PLANTON_MCP_TRANSPORT="both"`
3. Configure HTTP port and authentication as needed

### Adding HTTP to Existing Deployments

1. Update environment variables to include HTTP transport config
2. Expose the HTTP port in your deployment
3. Update firewall/security group rules
4. Test the HTTP endpoint before switching clients over

## Troubleshooting

### Port Already in Use

```
Error: listen tcp :8080: bind: address already in use
```

**Solution:** Change the port using `PLANTON_MCP_HTTP_PORT` or stop the conflicting process.

### Connection Refused

**Checklist:**
- Server is running in HTTP or both mode
- Port is correctly exposed (Docker: `-p 8080:8080`)
- Firewall allows connections
- Using correct hostname/IP

### Authentication Errors

If you encounter authentication issues:

**401 Unauthorized - Missing Authorization header:**
```bash
curl http://localhost:8080/sse
# Add the Authorization header with your API key
curl -H "Authorization: Bearer YOUR_PLANTON_API_KEY" http://localhost:8080/sse
```

**401 Unauthorized - Invalid bearer token:**
- Verify you're using your correct `PLANTON_API_KEY`
- Ensure the API key doesn't have leading/trailing spaces
- Check that the API key wasn't truncated
- Verify the API key is valid in the Planton Cloud console

## Performance Considerations

### HTTP Transport

- **Stateless** - No memory overhead for sessions
- **Horizontal Scaling** - Multiple instances can run in parallel
- **Connection Pooling** - Clients should reuse SSE connections
- **Load Balancing** - Compatible with standard HTTP load balancers

### STDIO Transport

- **Lower Latency** - Direct process communication
- **Single Client** - One client per process instance
- **No Network Overhead** - Ideal for local workflows

## Contributing

To contribute to HTTP transport implementation:

1. See `internal/mcp/http_server.go` for HTTP server logic
2. See `internal/config/config.go` for configuration
3. See `cmd/mcp-server-planton/main.go` for transport selection

Key areas needing contribution:
- Bearer token authentication middleware
- Custom HTTP server wrapper
- Health check endpoint
- Metrics and observability



