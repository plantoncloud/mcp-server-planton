# Rename Environment Variable to PLANTON_API_KEY

**Date:** November 25, 2025  
**Type:** Breaking Change  
**Impact:** All users must update their configuration

## Summary

Renamed the authentication environment variable from `USER_JWT_TOKEN` to `PLANTON_API_KEY` to accurately reflect that it accepts both JWT tokens and API keys from the Planton Cloud console.

## Motivation

The previous name `USER_JWT_TOKEN` was misleading because:

1. **Not always a JWT**: Users can now obtain API keys directly from the Planton Cloud console (Profile → API Keys), which may not be JWT tokens
2. **Confusing terminology**: The variable name suggested it only accepts JWT tokens, but the Planton Cloud APIs accept both JWT tokens and API keys
3. **Better alignment**: The new name `PLANTON_API_KEY` better represents the purpose and flexibility of the authentication mechanism

## Changes

### Environment Variable Rename

**Before:**
```bash
export USER_JWT_TOKEN="your-jwt-token"
```

**After:**
```bash
export PLANTON_API_KEY="your-api-key-or-jwt-token"
```

### Code Changes

- **Config struct field**: `UserJWTToken` → `PlantonAPIKey`
- **Function parameters**: `userToken` → `apiKey`
- **Error messages**: Updated to reference `PLANTON_API_KEY`
- **Log messages**: Updated to reference "API key" instead of "JWT token"
- **Documentation**: Comprehensive updates across all docs

### Files Modified

**Code:**
- `internal/config/config.go` - Config struct and loading logic
- `internal/grpc/interceptor.go` - Parameter names and comments
- `internal/grpc/client.go` - Parameter names and comments
- `internal/mcp/tools/environment.go` - Config field references
- `internal/mcp/server.go` - Log messages

**Documentation:**
- `README.md` - All examples and instructions
- `docs/configuration.md` - Complete configuration guide
- `docs/installation.md` - Installation and setup instructions
- `docs/development.md` - Development examples and tests
- `CONTRIBUTING.md` - Contribution examples

**Build:**
- `Makefile` - Docker run command

## How to Obtain API Key

The documentation now provides clear instructions for obtaining an API key:

### From Web Console (Recommended)

1. Log in to Planton Cloud web console
2. Click on your profile icon in the top-right corner
3. Select **API Keys** from the menu
4. Click **Create Key** to generate a new API key
5. Copy the generated key

**Note:** Existing API keys may not be visible in the console for security reasons, so it's recommended to create a new key.

### From CLI (Alternative)

```bash
planton auth login
planton auth token
```

## Migration Guide

Users need to update their configurations:

### 1. Update Environment Variables

```bash
# Old
export USER_JWT_TOKEN="your-token"

# New
export PLANTON_API_KEY="your-token"
```

### 2. Update LangGraph Configuration

```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "${PLANTON_API_KEY}",
        "PLANTON_APIS_GRPC_ENDPOINT": "${PLANTON_APIS_GRPC_ENDPOINT}"
      }
    }
  }
}
```

### 3. Update Claude Desktop Configuration

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "your-api-key",
        "PLANTON_APIS_GRPC_ENDPOINT": "apis.planton.cloud:443"
      }
    }
  }
}
```

### 4. Update .env Files

```env
# Old
USER_JWT_TOKEN=your-token

# New
PLANTON_API_KEY=your-token
```

### 5. Update Docker Commands

```bash
# Old
docker run -i --rm \
  -e USER_JWT_TOKEN="your-token" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest

# New
docker run -i --rm \
  -e PLANTON_API_KEY="your-token" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

## Backward Compatibility

**This is a breaking change.** The old `USER_JWT_TOKEN` environment variable is no longer supported. All users must update to `PLANTON_API_KEY`.

## Benefits

1. **Clearer semantics**: The name accurately reflects what it accepts (API keys or tokens)
2. **Better user experience**: Users can now easily obtain keys from the web console
3. **Consistent terminology**: Aligns with industry standards (similar to `OPENAI_API_KEY`, `ANTHROPIC_API_KEY`)
4. **Future-proof**: Accommodates different authentication mechanisms

## Testing

- ✅ All code files updated and linting passed
- ✅ All documentation files updated
- ✅ Build configuration (Makefile) updated
- ✅ No compilation errors

## Related Documentation

- [Configuration Guide](../docs/configuration.md) - Complete configuration documentation
- [Installation Guide](../docs/installation.md) - Updated installation instructions
- [README](../README.md) - Updated main documentation with new API key instructions







