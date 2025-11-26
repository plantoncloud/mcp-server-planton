# Migrate to Buf-Published Modules

**Date**: 2025-11-25  
**Type**: Bug Fix / Infrastructure  
**Impact**: Critical - Fixes CI/CD pipeline

## Problem

The GitHub Actions CI was failing during the "Download dependencies" step with the following error:

```
go: github.com/plantoncloud-inc/planton-cloud/apis@v0.0.0 (replaced by ../planton-cloud/apis): 
reading ../planton-cloud/apis/go.mod: open /home/runner/work/mcp-server-planton/planton-cloud/apis/go.mod: 
no such file or directory
```

The root cause was a `replace` directive in `go.mod` pointing to a local file path:

```go
replace github.com/plantoncloud-inc/planton-cloud/apis => ../planton-cloud/apis
```

This local path exists in the development environment but not in the GitHub Actions CI environment, causing the build to fail.

## Solution

Migrated from the local `replace` directive to using Buf Schema Registry's published Go modules. The Planton Cloud APIs are publicly available at https://buf.build/blintora/apis.

### Key Changes

1. **Removed Local Dependency**
   - Removed the `replace` directive from `go.mod`

2. **Added Buf Modules**
   - `buf.build/gen/go/blintora/apis/grpc/go` - gRPC service clients
   - `buf.build/gen/go/blintora/apis/protocolbuffers/go` - Protobuf message types
   - `buf.build/gen/go/project-planton/apis/protocolbuffers/go` - Project Planton shared types

3. **Updated Import Paths**
   
   **Old pattern:**
   ```go
   import "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/..."
   ```
   
   **New pattern:**
   ```go
   // For protobuf message types
   import "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/..."
   
   // For gRPC service clients
   import "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/.../xxxgrpc"
   ```

## Files Modified

### Go Source Files
- `internal/grpc/environment_client.go`
- `internal/grpc/cloud_resource_query_client.go`
- `internal/grpc/cloud_resource_search_client.go`
- `internal/mcp/tools/cloud_resource_search.go`
- `internal/mcp/tools/cloud_resource_lookup.go`

### Documentation
- `docs/development.md` - Updated example imports
- `CONTRIBUTING.md` - Updated example imports

### Configuration
- `go.mod` - Removed replace directive, added Buf modules
- `go.sum` - Updated with new module checksums

## Technical Details

### Buf Module Structure

The Buf Schema Registry publishes two types of Go modules for each protobuf repository:

1. **`protocolbuffers/go`**: Contains only protobuf message type definitions
   - Used for: Request/response types, enums, common messages
   - Import pattern: `buf.build/gen/go/{owner}/{repo}/protocolbuffers/go/{package}`

2. **`grpc/go`**: Contains gRPC service client stubs
   - Used for: Service clients (e.g., `NewEnvironmentQueryControllerClient`)
   - Import pattern: `buf.build/gen/go/{owner}/{repo}/grpc/go/{package}/{service}grpc`

### Example Migration

**Before:**
```go
import (
    environmentv1 "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/resourcemanager/environment/v1"
)

client := environmentv1.NewEnvironmentQueryControllerClient(conn)
```

**After:**
```go
import (
    environmentv1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/resourcemanager/environment/v1"
    environmentv1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/resourcemanager/environment/v1/environmentv1grpc"
)

client := environmentv1grpc.NewEnvironmentQueryControllerClient(conn)
```

## Verification

### Build Verification
```bash
$ go mod tidy
$ go build -v ./cmd/mcp-server-planton
# Build successful - binary created (29MB)
```

### Dependency Versions
- `buf.build/gen/go/blintora/apis/grpc/go`: v1.5.1-20251125011413-52ef5c4f2840.2
- `buf.build/gen/go/blintora/apis/protocolbuffers/go`: v1.36.10-20251125011413-52ef5c4f2840.1
- `buf.build/gen/go/project-planton/apis/protocolbuffers/go`: v1.36.10-20251124125039-9c224fb3651e.1

## Impact

✅ **CI/CD Pipeline**: GitHub Actions can now successfully download dependencies from public Buf registry  
✅ **Local Development**: Works identically - no changes needed to development workflow  
✅ **Dependency Management**: All dependencies are now versioned and fetched from remote registry  
✅ **Reproducibility**: Builds are now reproducible across all environments  

## Testing

- ✅ Local build successful
- ✅ All imports resolved correctly
- ⏳ CI pipeline verification (awaiting next push to GitHub)

## Notes

- The Buf modules are automatically generated from the protobuf definitions in the blintora/apis repository
- Version timestamps (e.g., `20251125011413`) represent the commit time in the Buf registry
- No code logic changes were required - only import path updates

## References

- Buf Module: https://buf.build/blintora/apis
- Buf Go Module Proxy: https://buf.build/docs/bsr/remote-plugins/go








