package cloudresource

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/clients"
	crinternal "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/cloudresource/internal"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateUpdateCloudResourceTool creates the MCP tool definition for updating a cloud resource.
func CreateUpdateCloudResourceTool() mcp.Tool {
	return mcp.Tool{
		Name: "update_cloud_resource",
		Description: `Update an existing cloud resource in Planton Cloud.

This tool allows you to modify the specification of an existing cloud resource.
It fetches the current resource, merges your changes, validates, and updates it.

WORKFLOW:
1. Get the current resource using 'get_cloud_resource_by_id'
2. Identify fields you want to change
3. Call this tool with resource_id and the spec changes
4. If validation fails, errors will indicate which fields are invalid

Note: You must provide the resource_id and the complete updated spec.`,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"resource_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the resource to update (required)",
				},
				"spec": map[string]interface{}{
					"type":        "object",
					"description": "Complete updated resource specification",
				},
				"version_message": map[string]interface{}{
					"type":        "string",
					"description": "Optional message describing the reason for this update (for audit trail)",
				},
			},
			Required: []string{"resource_id", "spec"},
		},
	}
}

// HandleUpdateCloudResource handles the MCP tool invocation for updating a cloud resource.
//
// This function:
//  1. Fetches the existing resource by ID
//  2. Extracts the kind and metadata from the existing resource
//  3. Wraps the new spec data into CloudResource
//  4. Validates the updated CloudResource
//  5. Calls CloudResourceCommandClient to update the resource
//  6. Unwraps and returns the updated resource
func HandleUpdateCloudResource(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// 1. Extract resource_id
	resourceID, ok := arguments["resource_id"].(string)
	if !ok || resourceID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "resource_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 2. Extract spec data
	specData, ok := arguments["spec"].(map[string]interface{})
	if !ok {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "spec is required and must be an object",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 3. Extract optional version message
	versionMessage, _ := arguments["version_message"].(string)

	log.Printf("Tool invoked: update_cloud_resource, resource_id=%s", resourceID)

	// 4. Fetch existing resource to get kind and metadata with per-user API key
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	queryClient, err := clients.NewCloudResourceQueryClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		queryClient, err = clients.NewCloudResourceQueryClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer queryClient.Close()

	existingResource, err := queryClient.GetById(ctx, resourceID)
	if err != nil {
		return errors.HandleGRPCError(err, ""), nil
	}

	// 5. Extract kind and metadata from existing resource
	kind := existingResource.GetSpec().GetKind()
	metadata := existingResource.GetMetadata()

	// 6. Add version message if provided
	if versionMessage != "" && metadata.GetVersion() != nil {
		metadata.Version.Message = versionMessage
	}

	log.Printf("Updating cloud resource: kind=%s, name=%s", kind.String(), metadata.GetName())

	// 7. Wrap the new spec data into CloudResource
	updatedResource, err := crinternal.WrapCloudResource(kind, specData, metadata)
	if err != nil {
		errResp := map[string]interface{}{
			"error":   "INVALID_SPEC_DATA",
			"message": fmt.Sprintf("Failed to update %s resource: %v", kind.String(), err),
			"hint":    fmt.Sprintf("Call 'get_cloud_resource_schema' with cloud_resource_kind='%s' for the complete schema", kind.String()),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 8. Create command client with per-user API key and update
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	commandClient, err := clients.NewCloudResourceCommandClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		commandClient, err = clients.NewCloudResourceCommandClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer commandClient.Close()

	result, err := commandClient.Update(ctx, updatedResource)
	if err != nil {
		return errors.HandleGRPCError(err, metadata.GetOrg()), nil
	}

	// 9. Unwrap the updated resource
	unwrappedResource, err := crinternal.UnwrapCloudResource(result)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to unwrap updated resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool completed: update_cloud_resource, resource_id=%s", resourceID)

	// 10. Return the updated resource as JSON
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: false,
		UseProtoNames:   true,
	}

	resultJSON, err := marshaler.Marshal(unwrappedResource)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}
