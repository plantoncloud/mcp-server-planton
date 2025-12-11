# Extract Specific Cloud Resource from CloudResource Wrapper in GET Tool

**Date**: November 25, 2025

## Summary

The MCP server's `get_cloud_resource_by_id` tool now extracts and returns the specific cloud resource object (e.g., `AwsEksCluster`, `GcpGkeCluster`, `KubernetesDeployment`) instead of the CloudResource wrapper. This aligns the MCP server with patterns used in the CLI and web console, hiding internal implementation details from AI agents and providing a cleaner, more intuitive interface.

## Problem Statement

The CloudResource API uses a wrapper structure to unify all cloud resource types under a single backend endpoint. While this simplifies backend code, it exposes internal implementation details to clients. The structure is:

```
CloudResource (wrapper)
├── metadata
├── spec
│   ├── kind (CloudResourceKind enum)
│   └── cloud_object (oneof with 100+ resource types)
│       ├── aws_eks_cluster
│       ├── gcp_gke_cluster
│       ├── kubernetes_deployment
│       └── ... (100+ other types)
└── status
```

### Pain Points

- **Inconsistent client patterns**: The CLI and web console both unwrap the CloudResource to expose only the specific resource, but the MCP server was returning the full wrapper
- **Exposed implementation details**: AI agents saw internal structures (`CloudObject`, `oneof` fields) that are backend implementation details
- **Confusing interface**: Agents had to understand the complex wrapper structure instead of working directly with resource-specific types
- **Poor developer experience**: Navigating the wrapper to find the actual resource data added unnecessary complexity

## Solution

Implement a `UnwrapCloudResource()` function that uses protobuf reflection to extract the specific cloud resource from the CloudResource wrapper, following the same pattern established in the CLI codebase.

### Architecture

```
gRPC API Response          Unwrap Function              MCP Tool Output
┌──────────────────┐      ┌─────────────────┐      ┌────────────────────┐
│ CloudResource    │      │ UnwrapCloud     │      │ AwsEksCluster      │
│   wrapper        │──────│   Resource()    │──────│   (specific type)  │
│                  │      │                 │      │                    │
│ • metadata       │      │ Uses protobuf   │      │ • metadata         │
│ • spec           │      │ reflection to   │      │ • spec             │
│   • kind         │      │ extract oneof   │      │ • status           │
│   • cloud_object │      │ field           │      │ • stack_outputs    │
│ • status         │      └─────────────────┘      └────────────────────┘
└──────────────────┘
```

The unwrapping process:
1. Validate CloudResource and its spec
2. Access CloudObject using protobuf reflection
3. Find the `oneof object` field descriptor
4. Determine which field is set (e.g., `aws_eks_cluster`)
5. Extract the specific resource message
6. Return as `proto.Message` interface

## Implementation Details

### New File: `unwrap.go`

Created a dedicated unwrapping function using protobuf reflection:

```go
func UnwrapCloudResource(cloudResource *cloudresourcev1.CloudResource) (proto.Message, error) {
    // Validate inputs
    if cloudResource == nil || cloudResource.Spec == nil {
        return nil, fmt.Errorf("cloud resource or spec is nil")
    }
    
    cloudObject := cloudResource.Spec.CloudObject
    if cloudObject == nil {
        return nil, fmt.Errorf("cloud object is nil in CloudResource spec")
    }
    
    // Use protobuf reflection to access the oneof field
    cloudObjectReflect := cloudObject.ProtoReflect()
    cloudObjectDescriptor := cloudObjectReflect.Descriptor()
    
    // Find the "object" oneof field
    oneofDescriptor := cloudObjectDescriptor.Oneofs().ByName("object")
    if oneofDescriptor == nil {
        return nil, fmt.Errorf("object oneof not found in CloudObject")
    }
    
    // Get which field is set in the oneof
    whichOneof := cloudObjectReflect.WhichOneof(oneofDescriptor)
    if whichOneof == nil {
        return nil, fmt.Errorf("no field is set in the object oneof")
    }
    
    // Extract and return the specific resource
    fieldValue := cloudObjectReflect.Get(whichOneof)
    messageValue := fieldValue.Message()
    return messageValue.Interface(), nil
}
```

**Key implementation details**:
- Uses Go's `protoreflect` package to dynamically access oneof fields
- No need for switch statements or manual type assertions
- Returns the generic `proto.Message` interface for flexibility
- Comprehensive error handling for nil checks and invalid states

### Updated: `get.go`

Modified the GET handler to unwrap before returning:

```go
func HandleGetCloudResourceById(ctx context.Context, arguments map[string]interface{}, cfg *config.Config) (*mcp.CallToolResult, error) {
    // ... existing validation and API call ...
    
    // Get cloud resource by ID (returns CloudResource wrapper)
    cloudResource, err := client.GetById(ctx, resourceID)
    if err != nil {
        return errors.HandleGRPCError(err, ""), nil
    }
    
    // NEW: Unwrap to get the specific cloud resource object
    unwrappedResource, err := UnwrapCloudResource(cloudResource)
    if err != nil {
        errResp := errors.ErrorResponse{
            Error:   "INTERNAL_ERROR",
            Message: fmt.Sprintf("Failed to unwrap cloud resource: %v", err),
        }
        errJSON, _ := json.MarshalIndent(errResp, "", "  ")
        return mcp.NewToolResultText(string(errJSON)), nil
    }
    
    // Marshal the unwrapped resource (not the wrapper)
    resultJSON, err := marshaler.Marshal(unwrappedResource)
    // ...
}
```

### Updated Tool Description

Clarified the tool description to reflect the new behavior:

```go
Description: "Get the complete state and configuration of a cloud resource by its ID. " +
    "Returns the specific cloud resource object (e.g., AwsEksCluster, GcpGkeCluster, KubernetesDeployment) " +
    "with its metadata, spec, and status. The response structure depends on the resource type. " +
    "Use this to inspect the complete manifest of a specific resource. " +
    "Resource IDs are returned by search_cloud_resources or lookup_cloud_resource_by_name.",
```

## Benefits

### Consistency Across Clients

- **CLI pattern**: Already uses `UnwrapCloudResource()` to extract specific resources from YAML manifests
- **Web console pattern**: Uses `extractResourceFromCloudResource()` to unwrap before displaying
- **MCP server pattern**: Now follows the same approach, maintaining consistency

### Improved Developer Experience

Before (with wrapper):
```json
{
  "api_version": "infra-hub.planton.ai/v1",
  "kind": "CloudResource",
  "metadata": { "id": "eks-abc123", ... },
  "spec": {
    "kind": "aws_eks_cluster",
    "cloud_object": {
      "object": {
        "case": "aws_eks_cluster",
        "value": {
          "api_version": "...",
          "kind": "AwsEksCluster",
          "metadata": { ... },
          "spec": { ... }
        }
      }
    }
  }
}
```

After (unwrapped):
```json
{
  "api_version": "code2cloud.planton.cloud/v1",
  "kind": "AwsEksCluster",
  "metadata": { "id": "eks-abc123", ... },
  "spec": {
    "cluster_config": { ... },
    "node_groups": [ ... ]
  },
  "status": { ... },
  "stack_outputs": { ... }
}
```

### Cleaner AI Agent Interface

- AI agents see only the relevant resource-specific fields
- No confusion about internal wrapper structures
- More intuitive to understand what fields are available
- Response structure matches the actual resource type

### Encapsulation of Implementation Details

- The CloudResource wrapper remains a backend optimization
- Clients don't need to know about the unified wrapper pattern
- Future backend refactoring won't affect client interfaces

## Impact

### For AI Agents

- **Simpler prompts**: Agents can refer to resources by their actual type names
- **Better context understanding**: Resource-specific fields are immediately visible
- **Reduced token usage**: No need to include wrapper boilerplate in responses

### For MCP Server

- **Aligned with platform patterns**: Follows established CLI and web console conventions
- **Maintainable**: Uses the same reflection pattern as other clients
- **Flexible**: Works with all 100+ cloud resource types without modification

### Code Quality

- **Files changed**: 2 (1 new, 1 modified)
- **Lines added**: ~80 (unwrap function + updated handler)
- **No breaking changes**: Tool interface remains the same, only output format changes
- **No linter errors**: Clean implementation following Go best practices

## Related Work

This change builds on established patterns from:

- **CLI unwrapping**: `planton-cloud/client-apps/cli/.../cloudresourceyaml/manifest_handler.go`
- **Web console extraction**: `planton-cloud/client-apps/web-console/.../cloud-resource/_services/command.ts`
- **Java backend mapper**: `planton-cloud/backend/libs/java/.../CloudResourceMapper.java`

All client-facing code now follows the same principle: **expose resource-specific types, hide the CloudResource wrapper**.

## Future Enhancements

Potential improvements to consider:

1. **Apply to other tools**: The search and lookup tools could also benefit from unwrapping (though they currently return simplified structures)
2. **Resource-specific descriptions**: Tool descriptions could be dynamically generated per resource type
3. **Type-safe returns**: Go generics could provide compile-time type safety for specific resource types

---

**Status**: ✅ Production Ready  
**Timeline**: ~1 hour implementation  
**Files Changed**: 2 (unwrap.go new, get.go modified)























