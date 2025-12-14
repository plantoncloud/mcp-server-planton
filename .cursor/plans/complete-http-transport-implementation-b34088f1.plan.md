<!-- b34088f1-991f-4101-9aee-82946c443fa2 3a5be82a-1e01-489b-984a-f3914fa14253 -->
# Complete HTTP Transport Implementation

## Current State Analysis

The HTTP transport implementation is partially complete:

- ✅ Basic SSE server setup using `mark3labs/mcp-go`
- ✅ Transport mode configuration (stdio/http/both)
- ✅ Environment variables defined
- ❌ Bearer token authentication (configured but not implemented)
- ❌ Health check endpoint (referenced in Dockerfile but not implemented)
- ❌ Custom HTTP server wrapper for middleware
- ❌ Complete documentation in README

## Research Findings

From GitHub MCP server implementations:

1. **Health Check**: Standard endpoint at `/health` returning `{"status":"ok"}` with 200 OK
2. **SSE Endpoints**: `/sse` (GET) and `/message` (POST) - already handled by mcp-go library
3. **Authentication**: Custom HTTP middleware wrapping the SSE handlers
4. **Documentation**: Clear sections for HTTP transport in README with examples

## Implementation Strategy

### 1. Remove Test Files

Delete all `test_*.go` files from the repository root:

- `test_token.go`
- `test_token_debug.go`
- `test_jwt_token.go`
- `test_iam_vs_search.go`
- `test_header_case.go`

### 2. Implement Custom HTTP Server Wrapper

**File**: `internal/mcp/http_server.go`

The current implementation uses `SSEServer.Start()` which creates its own HTTP server, preventing middleware integration. Replace this with a custom HTTP server that:

1. **Creates a custom `http.ServeMux`** for routing
2. **Adds health check endpoint** at `/health`
3. **Wraps SSE handlers with bearer token middleware** (if enabled)
4. **Manually integrates the SSE handlers** from the mcp-go library

**Key Components**:

```go
// Custom HTTP server with middleware support
func (s *Server) ServeHTTP(opts HTTPServerOptions) error {
    mux := http.NewServeMux()
    
    // Health check endpoint
    mux.HandleFunc("/health", healthCheckHandler)
    
    // SSE handlers with optional auth middleware
    sseServer := server.NewSSEServer(s.mcpServer, opts.BaseURL)
    
    if opts.AuthEnabled {
        // Wrap with bearer token middleware
        mux.HandleFunc("/sse", authMiddleware(opts.BearerToken, sseServer.HandleSSE))
        mux.HandleFunc("/message", authMiddleware(opts.BearerToken, sseServer.HandleMessage))
    } else {
        mux.HandleFunc("/sse", sseServer.HandleSSE)
        mux.HandleFunc("/message", sseServer.HandleMessage)
    }
    
    // Create and start HTTP server
    httpServer := &http.Server{
        Addr:    ":" + opts.Port,
        Handler: mux,
    }
    
    return httpServer.ListenAndServe()
}
```

**Bearer Token Middleware**:

```go
func authMiddleware(expectedToken string, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        
        if authHeader == "" {
            http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
            return
        }
        
        // Extract bearer token
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }
        
        if parts[1] != expectedToken {
            http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
            return
        }
        
        next(w, r)
    }
}
```

**Health Check Handler**:

```go
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
```

### 3. Update Documentation

#### README.md

Add a dedicated **"HTTP Transport"** section after the **"Quick Start"** section:

````markdown
## HTTP Transport

The MCP server supports HTTP transport using Server-Sent Events (SSE) for remote access and integrations.

### Transport Modes

- **stdio** (default): Standard input/output for local AI clients
- **http**: HTTP/SSE transport for remote access
- **both**: Run both transports simultaneously

### Running with HTTP Transport

#### Local Testing (No Authentication)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"

mcp-server-planton
````

Access the server:

- Health check: `curl http://localhost:8080/health`
- SSE endpoint: `curl http://localhost:8080/sse`

#### Production (With Authentication)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="true"
export PLANTON_MCP_HTTP_BEARER_TOKEN="your-secure-token"

mcp-server-planton
```

Access with bearer token:

```bash
curl -H "Authorization: Bearer your-secure-token" http://localhost:8080/sse
```

### Docker with HTTP Transport

```bash
docker run -p 8080:8080 \
  -e PLANTON_API_KEY="your-api-key" \
  -e PLANTON_MCP_TRANSPORT="http" \
  -e PLANTON_MCP_HTTP_AUTH_ENABLED="false" \
  ghcr.io/plantoncloud/mcp-server-planton:latest
```

### HTTP Endpoints

- `GET /health` - Health check endpoint
- `GET /sse` - SSE connection for MCP protocol
- `POST /message` - Message endpoint for MCP protocol

See [HTTP Transport Guide](docs/http-transport.md) for detailed documentation.

````

#### Configuration Table Update

Add HTTP transport variables to the configuration table:

```markdown
| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PLANTON_API_KEY` | Yes | - | User's API key |
| `PLANTON_CLOUD_ENVIRONMENT` | No | `live` | Target environment |
| `PLANTON_APIS_GRPC_ENDPOINT` | No | (based on env) | Override endpoint |
| `PLANTON_MCP_TRANSPORT` | No | `stdio` | Transport mode: `stdio`, `http`, or `both` |
| `PLANTON_MCP_HTTP_PORT` | No | `8080` | HTTP server port |
| `PLANTON_MCP_HTTP_AUTH_ENABLED` | No | `true` | Enable bearer token auth |
| `PLANTON_MCP_HTTP_BEARER_TOKEN` | Conditional | - | Bearer token (required if auth enabled) |
````

#### docs/configuration.md

Update to include HTTP transport configuration section with detailed examples of all transport modes.

#### docs/http-transport.md

Update the limitations section to reflect completed implementation:

- ✅ Bearer token authentication implemented
- ✅ Health check endpoint implemented
- Document how to test locally and in production

#### Dockerfile

The Dockerfile already has the health check configured correctly - it will work once we implement the `/health` endpoint.

### 4. Local Development Documentation

Add a **"Running Locally"** section in README under Development:

````markdown
### Running Locally

#### STDIO Mode (for Claude Desktop, Cursor)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_CLOUD_ENVIRONMENT="local"  # or "live"
./bin/mcp-server-planton
````

#### HTTP Mode (for testing remote access)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_CLOUD_ENVIRONMENT="local"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"  # disable for local testing

./bin/mcp-server-planton
```

Test the server:

```bash
# Health check
curl http://localhost:8080/health

# SSE connection (will stay open)
curl http://localhost:8080/sse
```

#### Both Modes (STDIO + HTTP)

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="both"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="false"

./bin/mcp-server-planton
```
```

## Files to Modify

1. `internal/mcp/http_server.go` - Rewrite with custom HTTP server
2. `README.md` - Add HTTP transport section and local running guide
3. `docs/configuration.md` - Add HTTP transport variables
4. `docs/http-transport.md` - Update limitations and examples
5. Delete: `test_*.go` files (5 files)

## Testing Checklist

After implementation:

- [ ] Build the binary: `make build`
- [ ] Test STDIO mode: `PLANTON_API_KEY=xxx ./bin/mcp-server-planton`
- [ ] Test HTTP mode without auth: Server starts and `/health` returns 200
- [ ] Test HTTP mode with auth: Requests without bearer token get 401
- [ ] Test Docker build: `make docker-build`
- [ ] Test Docker run in HTTP mode: `docker run -p 8080:8080 ...`
- [ ] Verify health check works in Docker container

### To-dos

- [ ] Remove all test_*.go files from repository root
- [ ] Rewrite http_server.go with custom HTTP server, bearer token middleware, and health check
- [ ] Add HTTP Transport section and local development guide to README
- [ ] Update configuration.md and http-transport.md with implementation details
- [ ] Test all transport modes (stdio, http, both) locally and with Docker