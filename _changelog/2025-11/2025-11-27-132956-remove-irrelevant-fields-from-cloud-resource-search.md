# Remove Irrelevant Fields from Cloud Resource Search Response

**Date**: November 27, 2025

## Summary

Simplified the `CloudResourceSimple` response structure by removing the `CreatedAt` and `IsReady` fields from cloud resource search and lookup operations. These fields were deemed irrelevant for the primary use case of resource discovery and identification, streamlining the response to show only essential metadata.

## Problem Statement

The `search_cloud_resources` and `lookup_cloud_resource_by_name` tools were returning fields that weren't relevant for their primary purpose of resource discovery:

- **`created_at`** (labeled as "Env created at"): Timestamp when the resource was created
- **`is_ready`**: Boolean indicating resource readiness status

### Pain Points

- **Information overload**: Extra fields cluttered the response without adding value for discovery workflows
- **Confusion**: The "Env created at" label suggested environment creation time rather than resource creation time
- **Unnecessary complexity**: AI agents and users had to parse additional fields they didn't need
- **Inconsistent focus**: Search results mixed identification data with operational status

## Solution

Remove the `CreatedAt` and `IsReady` fields from the `CloudResourceSimple` struct, which is used by both search and lookup tools. This focuses the response on identification and metadata fields only.

### Simplified Response Structure

Before:
```json
{
  "id": "eks-abc123",
  "name": "production-cluster",
  "slug": "production-cluster",
  "kind": "API_RESOURCE",
  "cloud_resource_kind": "AwsEksCluster",
  "org": "my-org",
  "env": "prod",
  "created_at": "2025-11-20T10:30:00Z",
  "description": "Production EKS cluster",
  "is_ready": true,
  "tags": ["production", "us-east-1"]
}
```

After:
```json
{
  "id": "eks-abc123",
  "name": "production-cluster",
  "slug": "production-cluster",
  "kind": "API_RESOURCE",
  "cloud_resource_kind": "AwsEksCluster",
  "org": "my-org",
  "env": "prod",
  "description": "Production EKS cluster",
  "tags": ["production", "us-east-1"]
}
```

## Implementation Details

### Files Modified

**`internal/domains/infrahub/cloudresource/search.go`**
- Removed `CreatedAt string` field from `CloudResourceSimple` struct (line 28)
- Removed `IsReady bool` field from `CloudResourceSimple` struct (line 30)
- Removed `CreatedAt` field assignment in `flattenCanvasResponse()` function (line 197)
- Removed `IsReady` field assignment in `flattenCanvasResponse()` function (line 199)
- Removed unused `formatTimestamp()` helper function (lines 216-222)
- Removed unused `timestamppb` import

**`internal/domains/infrahub/cloudresource/lookup.go`**
- Removed `CreatedAt` field assignment in `HandleLookupCloudResourceByName()` (line 166)
- Removed `IsReady` field assignment in `HandleLookupCloudResourceByName()` (line 168)

### Remaining Fields

The simplified `CloudResourceSimple` struct now contains only essential fields:

```go
type CloudResourceSimple struct {
    ID                string   `json:"id"`
    Name              string   `json:"name"`
    Slug              string   `json:"slug"`
    Kind              string   `json:"kind"`
    CloudResourceKind string   `json:"cloud_resource_kind"`
    Org               string   `json:"org"`
    Env               string   `json:"env"`
    Description       string   `json:"description,omitempty"`
    Tags              []string `json:"tags,omitempty"`
}
```

## Benefits

### For AI Agents

- **Focused context**: Only relevant identification and metadata fields in the response
- **Reduced token usage**: Smaller responses mean lower token costs for LLM interactions
- **Clearer prompts**: Agents can describe resources without irrelevant fields
- **Better understanding**: Response structure clearly indicates purpose (discovery, not monitoring)

### For Users

- **Less clutter**: Search results are easier to read and understand
- **Clear separation**: Discovery tools return identification data; monitoring tools would return status
- **Faster responses**: Slightly smaller JSON payloads

### For Developers

- **Clearer intent**: Response structure clearly indicates the tool's purpose
- **Simpler maintenance**: Fewer fields to maintain and document
- **Clean code**: Removed unused helper functions

## Impact

### Tools Affected

- `search_cloud_resources`: Returns simplified records for all matching resources
- `lookup_cloud_resource_by_name`: Returns simplified record for the found resource

### No Breaking Changes

The tools maintain their:
- Input parameter structure
- Core functionality
- Error handling behavior
- Authentication patterns

Only the output JSON structure changed by removing two optional fields.

### Status Information

Users who need resource readiness status can use `get_cloud_resource_by_id` to fetch the complete resource manifest, which includes detailed status information.

## Code Quality

- **Files changed**: 2
- **Lines removed**: ~10
- **Functions removed**: 1 (unused `formatTimestamp`)
- **Imports cleaned**: Removed unused `timestamppb`
- **Linter status**: ✅ No errors
- **Build status**: ✅ Successful

## Related Work

This change aligns with the principle established in other tools:
- **Search tools**: Return simplified records for discovery
- **Get tools**: Return complete manifests for detailed inspection
- **Status tools**: Would return operational status (future consideration)

The separation of concerns is now clearer: search/lookup focus on identification, while get operations provide complete resource details.

---

**Status**: ✅ Complete  
**Scope**: Focused refactoring  
**Impact**: Low - output simplification only
