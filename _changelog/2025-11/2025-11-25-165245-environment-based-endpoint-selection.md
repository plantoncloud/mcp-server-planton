# Environment-Based Endpoint Selection

**Date:** 2025-11-25  
**Type:** Enhancement  
**Impact:** Configuration Simplification

## Summary

Implemented environment-based endpoint selection matching the pattern used in Planton Cloud CLI. Users can now specify `PLANTON_CLOUD_ENVIRONMENT` (live/test/local) instead of manually configuring full endpoint URLs.

## Problem

The MCP server required users to manually specify the full gRPC endpoint URL via `PLANTON_APIS_GRPC_ENDPOINT`, defaulting to `localhost:8080`. This was:
- Inconsistent with the Planton Cloud CLI pattern
- Error-prone (users had to remember exact endpoint URLs)
- Not environment-aware (no automatic selection based on deployment environment)

Example of old configuration:
```bash
export PLANTON_APIS_GRPC_ENDPOINT="api.live.planton.cloud:443"
```

## Solution

Adopted the same endpoint selection pattern used in Planton Cloud CLI (`planton-cloud/client-apps/planton/internal/cli/backend`):

1. **Environment variable priority**:
   - `PLANTON_APIS_GRPC_ENDPOINT` (explicit override, highest priority)
   - `PLANTON_CLOUD_ENVIRONMENT` (environment-based selection)
   - Default to `live` environment

2. **Environment-to-endpoint mapping**:
   - `live` → `api.live.planton.ai:443`
   - `test` → `api.test.planton.cloud:443`
   - `local` → `localhost:8080`

3. **Simplified configuration**:
```bash
# Simple environment-based (recommended)
export PLANTON_CLOUD_ENVIRONMENT="live"

# Or explicit override when needed
export PLANTON_APIS_GRPC_ENDPOINT="custom-endpoint:443"
```

## Changes

### Code Changes

#### internal/config/config.go
- Added `Environment` type with constants (`EnvironmentLive`, `EnvironmentTest`, `EnvironmentLocal`)
- Added environment variable constants (`EnvironmentEnvVar`, `EndpointOverrideEnvVar`, `APIKeyEnvVar`)
- Added endpoint constants for each environment (`LiveEndpoint`, `TestEndpoint`, `LocalEndpoint`)
- Implemented `getEndpoint()` function with priority-based selection
- Implemented `getEnvironment()` function for environment detection
- Updated `LoadFromEnv()` to use new endpoint selection logic

#### Documentation
- Updated `README.md`:
  - Revised environment variables table to show new variables
  - Added environment-to-endpoint mapping table
  - Updated all integration examples (LangGraph, Claude Desktop, Docker)
  - Simplified Quick Start configuration
- Updated `docs/configuration.md`:
  - Added `PLANTON_CLOUD_ENVIRONMENT` documentation
  - Clarified `PLANTON_APIS_GRPC_ENDPOINT` as an override
  - Updated all configuration examples
  - Added environment-specific examples (local, test, production)

## Benefits

1. **Consistency**: Matches CLI pattern, reducing cognitive load for users familiar with CLI
2. **Simplicity**: Users only need to specify environment, not full URLs
3. **Safety**: Reduces risk of typos in endpoint URLs
4. **Flexibility**: Still allows custom endpoints via override variable
5. **Defaults**: Sensible default (live) works for most production cases

## Backward Compatibility

✅ **Fully backward compatible**

Existing configurations using `PLANTON_APIS_GRPC_ENDPOINT` continue to work without changes. The new environment variable provides an alternative, simpler configuration method.

Old configuration (still works):
```bash
export PLANTON_APIS_GRPC_ENDPOINT="api.live.planton.cloud:443"
```

New recommended configuration:
```bash
export PLANTON_CLOUD_ENVIRONMENT="live"
```

## Testing

1. **Build verification**: ✅ Successful compilation
2. **Formatting**: ✅ All code properly formatted
3. **Linting**: ✅ No linter errors

Manual testing required:
- [ ] Test with `PLANTON_CLOUD_ENVIRONMENT=live`
- [ ] Test with `PLANTON_CLOUD_ENVIRONMENT=test`
- [ ] Test with `PLANTON_CLOUD_ENVIRONMENT=local`
- [ ] Test with `PLANTON_APIS_GRPC_ENDPOINT` override
- [ ] Test default behavior (no env vars set, should use live)

## References

- CLI endpoint selection: `planton-cloud/client-apps/planton/internal/cli/backend/backend.go`
- CLI environment handling: `planton-cloud/client-apps/planton/internal/cli/backend/env/env.go`
- Resource Manager backend: `planton-cloud/client-apps/planton/internal/plantoncloud/domain/resourcemanager/resourcemanagerbackend/resource_manager_backend.go`

## Migration Guide

### For New Users

Use the simplified environment-based configuration:

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_CLOUD_ENVIRONMENT="live"  # or 'test', 'local'
```

### For Existing Users

No changes required. Your existing configuration continues to work:

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_APIS_GRPC_ENDPOINT="api.live.planton.cloud:443"
```

Optional: Simplify to the new pattern:

```bash
export PLANTON_API_KEY="your-api-key"
export PLANTON_CLOUD_ENVIRONMENT="live"
```

## Related Issues

This change addresses the user's concern about endpoint configuration and brings MCP server configuration in line with CLI patterns, making it easier for users already familiar with the CLI to adopt the MCP server.




