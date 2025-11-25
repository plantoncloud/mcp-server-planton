package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	grpcclient "github.com/plantoncloud-inc/mcp-server-planton/internal/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateGetCloudResourceByIdTool creates the MCP tool definition for getting a cloud resource by ID.
func CreateGetCloudResourceByIdTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_cloud_resource_by_id",
		Description: "Get the complete state and configuration of a cloud resource by its ID. " +
			"Returns the full CloudResource object including metadata, spec with detailed configuration, " +
			"and status information. Use this to inspect the complete manifest of a specific resource. " +
			"Resource IDs are returned by search_cloud_resources or lookup_cloud_resource_by_name.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"resource_id": map[string]interface{}{
					"type":        "string",
					"description": "Cloud resource ID (required). Examples: 'eks-abc123', 'gke-xyz789', 'k8sms-def456'",
				},
			},
			Required: []string{"resource_id"},
		},
	}
}

// HandleGetCloudResourceById handles the MCP tool invocation for getting a cloud resource by ID.
//
// This function:
//  1. Validates the resource_id argument
//  2. Calls CloudResourceQueryClient to get the full resource
//  3. Serializes the protobuf response to JSON
//  4. Returns the complete CloudResource manifest
func HandleGetCloudResourceById(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// Extract resource_id from arguments
	resourceID, ok := arguments["resource_id"].(string)
	if !ok || resourceID == "" {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "resource_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool invoked: get_cloud_resource_by_id, resource_id=%s", resourceID)

	// Create gRPC client with user's API key
	client, err := grpcclient.NewCloudResourceQueryClient(
		cfg.PlantonAPIsGRPCEndpoint,
		cfg.PlantonAPIKey,
	)
	if err != nil {
		errResp := ErrorResponse{
			Error:   "CLIENT_ERROR",
			Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}
	defer client.Close()

	// Get cloud resource by ID
	cloudResource, err := client.GetById(ctx, resourceID)
	if err != nil {
		return handleGRPCError(err, ""), nil
	}

	log.Printf("Tool completed: get_cloud_resource_by_id, retrieved resource: %s", resourceID)

	// Convert protobuf to JSON
	// Use protojson for better handling of Any types and other protobuf specifics
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: false, // Skip fields with default values
		UseProtoNames:   true,  // Use proto field names (snake_case)
	}

	resultJSON, err := marshaler.Marshal(cloudResource)
	if err != nil {
		errResp := ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal cloud resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

