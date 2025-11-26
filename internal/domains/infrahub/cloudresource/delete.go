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

// CreateDeleteCloudResourceTool creates the MCP tool definition for deleting a cloud resource.
func CreateDeleteCloudResourceTool() mcp.Tool {
	return mcp.Tool{
		Name: "delete_cloud_resource",
		Description: `Delete an existing cloud resource from Planton Cloud.

This tool permanently deletes a cloud resource and all associated infrastructure.
This operation cannot be undone.

IMPORTANT:
- Deletion triggers infrastructure teardown (e.g., deletes the actual cloud resources)
- Use with caution - this is a destructive operation
- Provide a version_message explaining why the resource is being deleted (for audit trail)

WORKFLOW:
1. Confirm the resource_id you want to delete
2. Optionally provide a version_message explaining the deletion
3. Call this tool
4. The resource and its infrastructure will be deleted`,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"resource_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the resource to delete (required)",
				},
				"version_message": map[string]interface{}{
					"type":        "string",
					"description": "Message explaining why the resource is being deleted (recommended for audit trail)",
				},
				"force": map[string]interface{}{
					"type":        "boolean",
					"description": "Force deletion even if there are dependencies (use with extreme caution)",
				},
			},
			Required: []string{"resource_id"},
		},
	}
}

// HandleDeleteCloudResource handles the MCP tool invocation for deleting a cloud resource.
//
// This function:
//  1. Validates the resource_id argument
//  2. Calls CloudResourceCommandClient to delete the resource
//  3. Returns confirmation with the deleted resource details
func HandleDeleteCloudResource(
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

	// 2. Extract optional parameters
	versionMessage, _ := arguments["version_message"].(string)
	force, _ := arguments["force"].(bool)

	log.Printf("Tool invoked: delete_cloud_resource, resource_id=%s, force=%v", resourceID, force)

	// 3. Create command client
	client, err := clients.NewCloudResourceCommandClient(
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
	defer client.Close()

	// 4. Call delete RPC
	// Note: The version_message and force parameters would need to be added to the Delete method
	// For now, we're using the simple Delete(resourceID) signature
	deletedResource, err := client.Delete(ctx, resourceID)
	if err != nil {
		return errors.HandleGRPCError(err, ""), nil
	}

	// 5. Unwrap the deleted resource
	unwrappedResource, err := crinternal.UnwrapCloudResource(deletedResource)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to unwrap deleted resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool completed: delete_cloud_resource, deleted resource_id=%s", resourceID)

	// 6. Return deletion confirmation with resource details
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: false,
		UseProtoNames:   true,
	}

	resourceJSON, err := marshaler.Marshal(unwrappedResource)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Wrap in a deletion response
	response := map[string]interface{}{
		"status":           "DELETED",
		"message":          fmt.Sprintf("Cloud resource %s has been deleted successfully", resourceID),
		"version_message":  versionMessage,
		"deleted_resource": json.RawMessage(resourceJSON),
	}

	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(responseJSON)), nil
}

