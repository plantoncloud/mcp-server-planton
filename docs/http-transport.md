# HTTP Transport Support

The MCP server now supports HTTP transport using Server-Sent Events (SSE) in addition to the original STDIO transport.

## Transport Modes

The server can run in three modes:

1. **stdio** (default) - Standard input/output transport for local AI clients
2. **http** - HTTP/SSE transport for remote access and webhooks
3. **both** - Run both transports simultaneously

## Configuration

Configure the transport using environment variables:

### Required Variables

- `PLANTON_API_KEY` - Your Planton Cloud API key (required for all modes)

### HTTP Transport Variables

- `PLANTON_MCP_TRANSPORT` - Transport mode: `stdio` | `http` | `both` (default: `stdio`)
- `PLANTON_MCP_HTTP_PORT` - HTTP server port (default: `8080`)
- `PLANTON_MCP_HTTP_AUTH_ENABLED` - Enable bearer token auth (default: `true`)
- `PLANTON_MCP_HTTP_BEARER_TOKEN` - Bearer token (required if auth enabled)

### Example Configurations

#### STDIO Mode (Local Development)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="stdio"
./mcp-server-planton
```

#### HTTP Mode (Remote Access)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"  # Set to "true" in production
./mcp-server-planton
```

#### Both Modes (Development + Remote)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="both"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"
./mcp-server-planton
```

## HTTP Endpoints

When running in HTTP mode, the following endpoints are available:

- `GET /sse` - SSE connection endpoint for MCP protocol
- `POST /message` - Message endpoint for MCP protocol

## Docker Usage

The Docker image supports all transport modes:

### STDIO Mode

```bash
docker run -e PLANTON_API_KEY="your-key" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

### HTTP Mode

```bash
docker run -p 8080:8080 \
  -e PLANTON_API_KEY="your-key" \
  -e PLANTON_MCP_TRANSPORT="http" \
  -e PLANTON_MCP_HTTP_AUTH_ENABLED="false" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

## Testing

### STDIO Mode Test

```bash
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | \
  PLANTON_API_KEY="test-key" \
  PLANTON_MCP_TRANSPORT="stdio" \
  ./mcp-server-planton
```

### HTTP Mode Test

Start the server:

```bash
PLANTON_API_KEY="test-key" \
  PLANTON_MCP_TRANSPORT="http" \
  PLANTON_MCP_HTTP_PORT="8080" \
  PLANTON_MCP_HTTP_AUTH_ENABLED="false" \
  ./mcp-server-planton
```

Connect to the SSE endpoint:

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

## Limitations and Future Enhancements

### Current Limitations

1. **Bearer Token Authentication** - Currently not fully implemented
   - The configuration is in place but middleware integration pending
   - The SSEServer creates its own HTTP server, making middleware integration complex

2. **Custom Endpoints** - No health check or metrics endpoints yet
   - SSEServer.Start() creates its own HTTP server without custom routes

3. **CORS Configuration** - Uses default CORS settings from the library

### Planned Enhancements

- [ ] Custom HTTP server wrapper for middleware support
- [ ] Bearer token authentication middleware
- [ ] Health check endpoint (`/health`)
- [ ] Metrics endpoint (`/metrics`)
- [ ] Request logging middleware
- [ ] Rate limiting
- [ ] TLS/HTTPS support
- [ ] Configurable CORS policies

## Security Considerations

### Production Deployments

When deploying HTTP transport in production:

1. **Enable Authentication** - Set `PLANTON_MCP_HTTP_AUTH_ENABLED="true"`
2. **Use Strong Tokens** - Generate secure random bearer tokens (32+ characters)
3. **TLS Termination** - Always run behind HTTPS (reverse proxy, load balancer)
4. **Network Isolation** - Use VPCs, security groups, firewalls
5. **API Key Protection** - Securely manage `PLANTON_API_KEY` (use secrets management)

### Two-Layer Security Model

```
Client → Bearer Token → MCP Server → PLANTON_API_KEY → Planton APIs
         (Who can access?)            (What can they access?)
```

- **Bearer Token** - Controls WHO can use your MCP server instance
- **PLANTON_API_KEY** - Controls WHAT resources the user can access (FGA from Planton Cloud)

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
      - PLANTON_MCP_HTTP_BEARER_TOKEN=${MCP_BEARER_TOKEN}
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
        - name: PLANTON_MCP_HTTP_BEARER_TOKEN
          valueFrom:
            secretKeyRef:
              name: planton-secrets
              key: bearer-token
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

### Cloudflare Workers (Future)

Cloudflare Workers deployment will be documented in a future update, aligned with existing webhook service patterns.

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

**Note:** Full bearer token authentication is pending implementation. For now:
- Set `PLANTON_MCP_HTTP_AUTH_ENABLED="false"` for testing
- Ensure network-level security (VPN, firewall) for production

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

