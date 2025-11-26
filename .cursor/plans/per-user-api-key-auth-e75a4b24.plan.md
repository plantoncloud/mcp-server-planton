<!-- e75a4b24-3864-491e-a5e7-6d782a185a9b 61685218-6e92-49f6-8746-ee9c6a626473 -->
# Per-User API Key Authentication for HTTP Transport

## Problem Statement

Currently, the MCP server uses a **machine account pattern** where:

- Docker starts with `PLANTON_API_KEY` environment variable
- All HTTP users must use the SAME API key to authenticate
- All gRPC calls to Planton Cloud use the SAME API key
- **Security Issue**: User A can access User B's data if they share the MCP server instance

## Solution: Per-User Passthrough Authentication

Extract each user's API key from the HTTP `Authorization` header and pass it directly to Planton Cloud APIs (no intermediate validation).

```
User A → Bearer user_a_key → MCP Server → gRPC with user_a_key → Planton APIs (validates + FGA)
User B → Bearer user_b_key → MCP Server → gRPC with user_b_key → Planton APIs (validates + FGA)
```

## Architecture Changes

### 1. Context-Based API Key Passing

**Key File**: `internal/mcp/http_server.go`

**Current Flow**:

```go
HTTP request → Validate token against config.PlantonAPIKey → Forward to MCP tools
```

**New Flow**:

```go
HTTP request → Extract user's API key from Authorization header → Store in context → Pass to MCP tools
```

### 2. Make PLANTON_API_KEY Optional for HTTP Mode

**Key File**: `internal/config/config.go`

**Current**: `PLANTON_API_KEY` is always required

**New**:

- Required for STDIO mode (user's personal key)
- Optional for HTTP mode (extracted from headers)
- Can still be provided for HTTP as a fallback/default

### 3. gRPC Clients Use Context API Key

**Key Files**:

- `internal/domains/infrahub/clients/cloudresource_client.go`
- `internal/domains/resourcemanager/clients/environment_client.go`

**Current**: Clients created once with config.PlantonAPIKey

**New**: Clients extract API key from request context per-call

## Implementation Steps

### Step 1: Add Context Key for API Key

**File**: `internal/common/auth/credentials.go`

Add context key and helper functions:

```go
type contextKey string

const apiKeyContextKey contextKey = "planton-api-key"

// WithAPIKey adds API key to context
func WithAPIKey(ctx context.Context, apiKey string) context.Context

// GetAPIKey retrieves API key from context
func GetAPIKey(ctx context.Context) (string, error)
```

### Step 2: Update HTTP Authentication Middleware

**File**: `internal/mcp/http_server.go`

Modify `createAuthenticatedProxy()` to:

1. Extract token from Authorization header
2. Store in request context (not validate against config)
3. Forward request with enriched context
```go
// Extract bearer token
token := parts[1]

// Store in context for downstream use
ctx := auth.WithAPIKey(r.Context(), token)
r = r.WithContext(ctx)

// Forward to MCP server
proxyRequest(w, r, targetAddr)
```


### Step 3: Make Config API Key Optional for HTTP

**File**: `internal/config/config.go`

Update `LoadFromEnv()`:

```go
apiKey := os.Getenv(APIKeyEnvVar)

// For STDIO mode, API key is required
transport := getTransport()
if transport == TransportStdio && apiKey == "" {
    return nil, fmt.Errorf(
        "%s required for STDIO transport",
        APIKeyEnvVar,
    )
}

// For HTTP mode, API key is optional (extracted from headers)
// If provided, it can be used as fallback/validation
```

### Step 4: Update gRPC Client Creation

**Files**:

- `internal/domains/infrahub/clients/cloudresource_client.go`
- `internal/domains/infrahub/clients/cloudresource_command_client.go`
- `internal/domains/resourcemanager/clients/environment_client.go`

Add new constructors that accept context:

```go
// NewCloudResourceQueryClientFromContext creates client using API key from context
func NewCloudResourceQueryClientFromContext(ctx context.Context, grpcEndpoint string) (*CloudResourceQueryClient, error) {
    apiKey, err := auth.GetAPIKey(ctx)
    if err != nil {
        return nil, fmt.Errorf("no API key in context: %w", err)
    }
    return NewCloudResourceQueryClient(grpcEndpoint, apiKey)
}
```

### Step 5: Update MCP Tool Handlers

**Files**:

- `internal/domains/infrahub/cloudresource/*.go` (all tool handlers)
- `internal/domains/resourcemanager/environment/*.go`

Update handlers to extract API key from context:

```go
func handleGetCloudResource(cfg *config.Config) mcp.ToolHandler {
    return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Create client from context (gets per-user API key)
        client, err := clients.NewCloudResourceQueryClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
        if err != nil {
            return nil, err
        }
        defer client.Close()
        
        // ... rest of handler
    }
}
```

## Configuration Changes

### Updated Environment Variables

**For STDIO Mode** (unchanged):

```bash
export PLANTON_API_KEY="user-personal-key"
export PLANTON_MCP_TRANSPORT="stdio"
```

**For HTTP Mode** (NEW):

```bash
# API key NOT required in environment
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_PORT="8080"
export PLANTON_MCP_HTTP_AUTH_ENABLED="true"

# Optional: Fallback API key (advanced use case)
# export PLANTON_API_KEY="fallback-key"
```

**Cursor Configuration** (unchanged):

```json
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "http://localhost:8080/",
      "headers": {
        "Authorization": "Bearer USER_PERSONAL_KEY"
      }
    }
  }
}
```

## Documentation Updates

### Files to Update

1. **README.md**: Update Docker example to remove PLANTON_API_KEY for HTTP mode
2. **docs/http-transport.md**: 

   - Remove machine account limitation note
   - Add multi-user support documentation
   - Update security model diagram

3. **docs/configuration.md**: Update PLANTON_API_KEY description for HTTP mode

### Key Documentation Changes

**Remove this limitation**:

```markdown
**Note:** In production deployments, each user should have their own instance 
of the MCP server with their own API key, or use the hosted endpoint...
```

**Add this capability**:

```markdown
**Multi-User Support**: The HTTP transport extracts each user's API key from 
the Authorization header and passes it to Planton Cloud APIs, ensuring proper 
Fine-Grained Authorization per user.
```

## Testing Strategy

### Test Cases

1. **Multiple users with different API keys**

   - User A queries their resources
   - User B queries their resources
   - Verify User A cannot see User B's data

2. **Invalid API key handling**

   - Send request with invalid token
   - Verify Planton APIs reject with proper error

3. **STDIO mode unchanged**

   - Verify STDIO still requires PLANTON_API_KEY
   - Verify STDIO uses config API key

4. **HTTP mode without auth**

   - Set PLANTON_MCP_HTTP_AUTH_ENABLED=false
   - Verify requests work without Authorization header
   - Verify proper error handling

## Security Benefits

### Before (Machine Account)

- ❌ All users share one API key
- ❌ One compromised key = all data exposed
- ❌ No per-user audit trail
- ❌ Cannot restrict per-user permissions

### After (Per-User Passthrough)

- ✅ Each user uses their own API key
- ✅ Compromised key affects only that user
- ✅ Complete audit trail per user
- ✅ Full FGA enforcement per user
- ✅ True multi-tenant security

## Backward Compatibility

### STDIO Mode

- ✅ No changes - continues to work as before
- ✅ Still requires PLANTON_API_KEY in environment

### HTTP Mode (Self-Hosted)

- ⚠️ Breaking change for existing deployments
- Migration: Users must pass API key in Authorization header
- Docker PLANTON_API_KEY no longer used for HTTP transport

### Hosted Endpoint

- ✅ No changes - already uses this pattern (presumably)

## Files to Modify

1. `internal/common/auth/credentials.go` - Add context helpers
2. `internal/config/config.go` - Make API key optional for HTTP
3. `internal/mcp/http_server.go` - Extract token to context
4. `internal/domains/infrahub/clients/*.go` - Add context-based constructors
5. `internal/domains/infrahub/cloudresource/*.go` - Update all tool handlers
6. `internal/domains/resourcemanager/clients/*.go` - Add context-based constructors
7. `internal/domains/resourcemanager/environment/*.go` - Update tool handlers
8. `README.md` - Update Docker examples
9. `docs/http-transport.md` - Update security model
10. `docs/configuration.md` - Update API key documentation

## Migration Guide for Users

### If Running HTTP Mode Self-Hosted

**Old Setup**:

```bash
docker run -e PLANTON_API_KEY="shared-key" ...
```

**New Setup**:

```bash
# No API key in Docker environment
docker run -e PLANTON_MCP_TRANSPORT="http" ...
```

**Cursor Config** (no change):

```json
{
  "headers": {
    "Authorization": "Bearer YOUR_PERSONAL_KEY"
  }
}
```

### If Running STDIO Mode

No changes required - continues to work exactly as before.

### To-dos

- [ ] Add context key and helper functions for API key in internal/common/auth/credentials.go
- [ ] Make PLANTON_API_KEY optional for HTTP mode in internal/config/config.go
- [ ] Modify HTTP authentication to extract token to context in internal/mcp/http_server.go
- [ ] Add context-based constructors for all gRPC clients (infrahub and resourcemanager)
- [ ] Update all MCP tool handlers to use context-based client creation
- [ ] Update README.md Docker examples to remove PLANTON_API_KEY for HTTP mode
- [ ] Update docs/http-transport.md to document multi-user support and remove limitation note
- [ ] Update docs/configuration.md to clarify PLANTON_API_KEY usage per transport mode
- [ ] Test multiple users with different API keys to verify proper isolation