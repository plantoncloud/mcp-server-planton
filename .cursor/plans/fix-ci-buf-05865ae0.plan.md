<!-- 05865ae0-2318-4ec3-a7a0-cc16f3d2eb39 ca072b16-b92c-4203-9466-20435bf7ce1b -->
# Fix CI Build by Using Buf-Published APIs

## Problem

The CI is failing because `go.mod` has a `replace` directive pointing to a local path (`../planton-cloud/apis`) that doesn't exist in the GitHub Actions environment.

## Solution

Replace the local dependency with the Buf-published Go module from https://buf.build/blintora/apis.

## Changes Required

### 1. Update go.mod

- Remove the `replace` directive (line 22)
- Update the require statement to use: `buf.build/gen/go/blintora/apis/protocolbuffers/go`
- Add proper version tag (will need to check latest version from Buf)

### 2. Update Import Paths

Replace all imports across the codebase:

- **Old pattern**: `github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/...`
- **New pattern**: `buf.build/gen/go/blintora/apis/protocolbuffers/go/...`

Files to update:

- `internal/mcp/tools/cloud_resource_search.go`
- `internal/mcp/tools/cloud_resource_lookup.go`
- `internal/grpc/environment_client.go`
- `internal/grpc/cloud_resource_search_client.go`
- `internal/grpc/cloud_resource_query_client.go`

### 3. Update Documentation

Update example imports in:

- `docs/development.md`
- `CONTRIBUTING.md`

### 4. Verify Changes

- Run `go mod tidy` to update dependencies
- Run `go build ./cmd/mcp-server-planton` to verify it compiles
- Ensure CI workflow will now work

### To-dos

- [ ] Update go.mod: remove replace directive and fix require statement
- [ ] Update all Go file imports to use Buf module path
- [ ] Update documentation files with new import examples
- [ ] Run go mod tidy and verify the build works locally