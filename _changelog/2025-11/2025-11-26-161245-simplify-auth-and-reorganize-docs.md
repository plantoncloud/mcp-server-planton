# Simplify Authentication and Reorganize Documentation for User-Focused Integration

**Date**: November 26, 2025

## Summary

Simplified HTTP authentication by using `PLANTON_API_KEY` as the bearer token (eliminating the separate `PLANTON_MCP_HTTP_BEARER_TOKEN`), and restructured all documentation to be user-focused with clear Cursor integration examples for both remote (`https://mcp.planton.ai/`) and local deployments.

## Problem Statement

### Complex Authentication

The HTTP transport implementation used two separate credentials:
- `PLANTON_API_KEY` - For authenticating with Planton Cloud APIs
- `PLANTON_MCP_HTTP_BEARER_TOKEN` - For authenticating HTTP requests to the MCP server

This created confusion:
- Users had to manage two different tokens
- Documentation was unclear about which token to use where
- Configuration examples showed conflicting patterns
- Security model had unnecessary complexity

### Documentation Not User-Focused

The README was overly technical and implementation-focused:
- Started with installation instructions before showing integration examples
- Verbose HTTP transport details cluttered the main README
- No clear examples for Cursor integration
- Missing prominent section for the hosted endpoint (`https://mcp.planton.ai/`)
- Local setup instructions scattered across multiple sections
- Configuration table included too many details for quick start

### Missing Cursor Integration Examples

Users couldn't easily find:
- How to configure `~/.cursor/mcp.json` for remote endpoint
- How to configure Cursor for local Docker testing
- How to configure Cursor for local binary usage
- Clear distinction between different usage modes

## Solution

### 1. Unified Authentication

**Simplified to single credential approach:**
- `PLANTON_API_KEY` now serves both purposes:
  - Bearer token for HTTP transport authentication
  - API key for Planton Cloud API authorization
- Removed `PLANTON_MCP_HTTP_BEARER_TOKEN` environment variable
- Updated code to use `cfg.PlantonAPIKey` as bearer token

**Benefits:**
- Simpler user experience - one API key to manage
- Clearer security model - single credential for authentication and authorization
- Easier configuration - fewer environment variables
- Consistent with user-scoped permissions model

### 2. User-Focused README

**Restructured README.md to prioritize user needs:**

1. **Quick Start First**: Integration examples before installation
2. **Cursor Integration Prominent**: Three clear configuration paths:
   - Remote endpoint (recommended): `https://mcp.planton.ai/`
   - Local Docker testing: `http://localhost:8080/`
   - Local binary (STDIO): Process spawning
3. **Simplified Installation**: Brief commands with links to detailed docs
4. **Essential Configuration Only**: Core environment variables in main README
5. **Tool Overview**: Clear listing of available tools grouped by category
6. **Documentation Links**: Easy access to detailed guides

### 3. Enhanced HTTP Transport Documentation

**Updated `docs/http-transport.md` with:**

- **Remote Access Section** (new):
  - Hosted endpoint configuration for Cursor
  - Benefits of using managed endpoint
  - No local installation required

- **Running Locally Section** (enhanced):
  - Complete Docker setup with Cursor configuration
  - Binary installation and setup
  - Test commands for verification
  - Clear authentication examples

- **Simplified Examples**:
  - All examples use unified authentication
  - Removed separate bearer token references
  - Added Cursor mcp.json snippets throughout

## Implementation Details

### Code Changes

#### 1. Configuration Simplification

**File**: `internal/config/config.go`

**Removed:**
```go
// HTTPBearerTokenEnvVar specifies the bearer token for HTTP authentication
HTTPBearerTokenEnvVar = "PLANTON_MCP_HTTP_BEARER_TOKEN"

// HTTPBearerToken is the bearer token for HTTP authentication
HTTPBearerToken string
```

**Updated LoadFromEnv():**
- Removed bearer token validation
- Removed bearer token from Config struct
- Updated documentation comments

#### 2. HTTP Server Update

**File**: `internal/mcp/http_server.go`

**Changed:**
```go
// Before
func DefaultHTTPOptions(cfg *config.Config) HTTPServerOptions {
    return HTTPServerOptions{
        BearerToken: cfg.HTTPBearerToken,
        // ...
    }
}

// After
func DefaultHTTPOptions(cfg *config.Config) HTTPServerOptions {
    return HTTPServerOptions{
        BearerToken: cfg.PlantonAPIKey, // Use API key as bearer token
        // ...
    }
}
```

**Added documentation:**
```go
// HTTPServerOptions configures the HTTP server
type HTTPServerOptions struct {
    BearerToken string // PLANTON_API_KEY used as bearer token for HTTP authentication
    // ...
}
```

### Documentation Changes

#### README.md - Complete Restructure

**New Structure:**
1. Overview (concise)
2. Quick Start
   - Get Your API Key
   - Integration with Cursor (3 options)
   - Integration with Claude Desktop
   - Integration with LangGraph
3. Available Tools (organized by category)
4. Configuration (essential variables only)
5. Security (simplified explanation)
6. Documentation Links
7. Support

**Key Sections Added:**

**Get Your API Key:**
```markdown
1. Log in to [Planton Cloud Console](https://console.planton.cloud)
2. Click your profile icon → **API Keys**
3. Click **Create Key** and copy the generated key
```

**Cursor Remote Endpoint:**
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

**Cursor Local Docker:**
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

**Cursor Local Binary (STDIO):**
```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "YOUR_PLANTON_API_KEY",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```

**Removed from README:**
- Verbose HTTP transport configuration details
- Detailed environment variable descriptions
- Development workflow instructions
- Project structure details
- Make targets and build instructions

**Moved to Linked Docs:**
- HTTP transport details → `docs/http-transport.md`
- Configuration reference → `docs/configuration.md`
- Development setup → `docs/development.md`
- Installation details → `docs/installation.md`

#### docs/http-transport.md - Enhanced for Users

**Added Remote Access Section:**
```markdown
## Remote Access (Recommended)

### Using the Hosted Endpoint

The easiest way to use the MCP server is via the hosted endpoint at `https://mcp.planton.ai/`.

**Cursor Configuration:**
[Cursor mcp.json example]

**Benefits:**
- No local installation required
- Always up-to-date with latest features
- Managed and monitored by Planton Cloud
- High availability and performance
```

**Enhanced Running Locally Section:**

**Local Setup with Docker:**
1. Run the Docker container (with full command)
2. Configure Cursor (with mcp.json example)
3. Test the connection (with curl commands)

**Local Setup with Binary:**
1. Install the binary (with platform-specific commands)
2. Start the server (with environment variables)
3. Configure Cursor (same as Docker)
4. Test the connection (same as Docker)

**Updated Security Section:**
```markdown
### Security Model

Client → PLANTON_API_KEY (Bearer) → MCP Server → PLANTON_API_KEY → Planton APIs
         (Authentication)                          (Authorization & FGA)

- **PLANTON_API_KEY as Bearer Token** - Authenticates access to your MCP server instance
- **PLANTON_API_KEY to Planton APIs** - Enforces your actual permissions (Fine-Grained Authorization)

This unified approach simplifies authentication while maintaining security through your user permissions.
```

**Removed:**
- All references to `PLANTON_MCP_HTTP_BEARER_TOKEN`
- Separate token generation instructions
- Two-layer security model explanation (now simplified)

#### docs/configuration.md - Updated Reference

**Updated PLANTON_MCP_HTTP_AUTH_ENABLED section:**
```markdown
**Authentication mechanism:**
When enabled, your `PLANTON_API_KEY` is used as the bearer token for HTTP authentication. 
This simplifies configuration by using a single credential for both MCP server access and 
Planton Cloud API authorization.
```

**Removed PLANTON_MCP_HTTP_BEARER_TOKEN section:**
- Removed environment variable documentation
- Removed from configuration struct example
- Removed from .env file examples
- Removed from integration examples

**Updated Config struct example:**
```go
type Config struct {
    PlantonAPIKey           string
    PlantonAPIsGRPCEndpoint string
    Transport               TransportMode
    HTTPPort                string
    HTTPAuthEnabled         bool
    // HTTPBearerToken removed
}
```

## Files Modified

### Code Files
- `internal/config/config.go` - Removed bearer token configuration
- `internal/mcp/http_server.go` - Use API key as bearer token

### Documentation Files
- `README.md` - Complete restructure with user-focused content
- `docs/http-transport.md` - Added remote access section and Cursor examples
- `docs/configuration.md` - Removed bearer token references

### Lines Changed
- Code: ~30 lines modified
- Documentation: ~400 lines restructured/rewritten

## Configuration Changes

### Before
```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_AUTH_ENABLED="true"
export PLANTON_MCP_HTTP_BEARER_TOKEN="your-separate-bearer-token"
```

### After
```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_TRANSPORT="http"
export PLANTON_MCP_HTTP_AUTH_ENABLED="true"
# PLANTON_API_KEY is used as bearer token automatically
```

## Testing

### Verified Authentication Works

**Test 1: HTTP with Authentication**
```bash
docker run -p 8080:8080 \
  -e PLANTON_API_KEY="test-key" \
  -e PLANTON_MCP_TRANSPORT="http" \
  -e PLANTON_MCP_HTTP_AUTH_ENABLED="true" \
  ghcr.io/plantoncloud/mcp-server-planton:latest

curl -H "Authorization: Bearer test-key" http://localhost:8080/health
# Expected: {"status":"ok"}
```

**Test 2: Verify Bearer Token Matches API Key**
```bash
# Should succeed
curl -H "Authorization: Bearer test-key" http://localhost:8080/sse

# Should fail with 401
curl -H "Authorization: Bearer wrong-key" http://localhost:8080/sse
```

### Cursor Integration Tests

**Test 3: Remote Endpoint Configuration**
```json
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "https://mcp.planton.ai/",
      "headers": {
        "Authorization": "Bearer pck_..."
      }
    }
  }
}
```
✅ Cursor successfully connects to remote endpoint

**Test 4: Local Docker Configuration**
```json
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "http://localhost:8080/",
      "headers": {
        "Authorization": "Bearer pck_..."
      }
    }
  }
}
```
✅ Cursor successfully connects to local Docker instance

**Test 5: STDIO Configuration**
```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "pck_...",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```
✅ Cursor successfully spawns and connects to local binary

## Benefits

### For Users

1. **Simpler Setup**: One API key instead of two separate tokens
2. **Clear Integration Paths**: Obvious choices for different use cases
3. **Better Documentation**: User-focused with practical examples
4. **Faster Time to Value**: Can start using hosted endpoint immediately
5. **Easier Troubleshooting**: Fewer configuration variables to debug

### For Documentation

1. **Better Organization**: Technical details in appropriate docs
2. **Clearer Examples**: Cursor configurations prominent
3. **Reduced Confusion**: No conflicting information about tokens
4. **Easier Maintenance**: Changes only need updating in one place

### For Security

1. **Unified Model**: Simpler to understand and explain
2. **User-Scoped**: API key enforces actual user permissions
3. **Less Exposure**: Only one credential to manage
4. **Clear Flow**: Authentication → Authorization path is obvious

## Migration Guide

### For Existing Users

If you're currently using the MCP server with HTTP transport:

**Old Configuration:**
```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_MCP_HTTP_BEARER_TOKEN="your-bearer-token"
```

**New Configuration:**
```bash
export PLANTON_API_KEY="your-api-key"
# That's it! No separate bearer token needed
```

**Cursor Configuration Update:**

Old:
```json
{
  "type": "http",
  "url": "http://localhost:8080/",
  "headers": {
    "Authorization": "Bearer your-bearer-token"
  }
}
```

New:
```json
{
  "type": "http",
  "url": "http://localhost:8080/",
  "headers": {
    "Authorization": "Bearer YOUR_PLANTON_API_KEY"
  }
}
```

### Deprecation Notice

The `PLANTON_MCP_HTTP_BEARER_TOKEN` environment variable is no longer supported. If set, it will be ignored. The server now uses `PLANTON_API_KEY` for HTTP authentication.

## Documentation Structure

### Before
```
README.md (500+ lines)
├── Installation (verbose)
├── Quick Start (buried)
├── HTTP Transport (detailed)
├── Configuration (all variables)
├── Development (detailed)
├── Distribution
├── Roadmap
└── Contributing

docs/http-transport.md
├── Configuration examples
└── Use cases
```

### After
```
README.md (200 lines, user-focused)
├── Overview (concise)
├── Quick Start (prominent)
│   ├── Get API Key
│   ├── Cursor Integration (3 modes)
│   ├── Claude Desktop Integration
│   └── LangGraph Integration
├── Available Tools (organized)
├── Configuration (essential only)
├── Security (simplified)
└── Links to detailed docs

docs/http-transport.md (enhanced)
├── Remote Access (new)
│   ├── Hosted endpoint
│   ├── Cursor configuration
│   └── Benefits
├── Running Locally
│   ├── Docker setup with Cursor config
│   ├── Binary setup with Cursor config
│   └── Testing instructions
├── Security Model (simplified)
└── Deployment Examples

docs/configuration.md (updated)
├── Essential variables
├── Authentication mechanism
└── Complete reference

docs/development.md (unchanged)
└── Technical details for contributors
```

## Next Steps

### Documentation
- ✅ README.md restructured
- ✅ HTTP transport docs enhanced
- ✅ Configuration docs updated
- ✅ Cursor integration examples added

### Future Enhancements
- [ ] Add Claude Desktop HTTP configuration examples
- [ ] Create video tutorial for Cursor integration
- [ ] Add troubleshooting section with common Cursor issues
- [ ] Document hosted endpoint availability/SLA

## Conclusion

This change significantly improves the user experience by:

1. **Simplifying authentication** from two tokens to one
2. **Restructuring documentation** to prioritize user integration
3. **Adding clear Cursor examples** for all usage modes
4. **Highlighting the hosted endpoint** as the recommended approach
5. **Providing complete local setup instructions** for development

The unified authentication model makes the MCP server easier to use while maintaining the same level of security through user-scoped permissions. The reorganized documentation helps users get started quickly with clear examples for their preferred integration method.
