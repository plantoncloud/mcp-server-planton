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

// CreateGetCloudResourceByIdTool creates the MCP tool definition for getting a cloud resource by ID.
func CreateGetCloudResourceByIdTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_cloud_resource_by_id",
		Description: "Get the complete state and configuration of a cloud resource by its ID. " +
			"Returns the specific cloud resource object (e.g., AwsEksCluster, GcpGkeCluster, KubernetesDeployment) " +
			"with its metadata, spec, and status. The response structure depends on the resource type. " +
			"Use this to inspect the complete manifest of a specific resource. " +
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
//  2. Calls CloudResourceQueryClient to get the CloudResource wrapper
//  3. Unwraps to extract the specific cloud resource object (e.g., AwsEksCluster, GcpGkeCluster)
//  4. Serializes the specific resource to JSON
//  5. Returns the resource-specific manifest (not the CloudResource wrapper)
func HandleGetCloudResourceById(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// Extract resource_id from arguments
	resourceID, ok := arguments["resource_id"].(string)
	if !ok || resourceID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "resource_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool invoked: get_cloud_resource_by_id, resource_id=%s", resourceID)

	// Create gRPC client with per-user API key from context
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	client, err := clients.NewCloudResourceQueryClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		client, err = clients.NewCloudResourceQueryClient(
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
	defer client.Close()

	// Get cloud resource by ID (returns CloudResource wrapper)
	cloudResource, err := client.GetById(ctx, resourceID)
	if err != nil {
		return errors.HandleGRPCError(err, ""), nil
	}

	// Unwrap to get the specific cloud resource object
	// This extracts the actual resource (e.g., AwsEksCluster, GcpGkeCluster) from the wrapper
	unwrappedResource, err := crinternal.UnwrapCloudResource(cloudResource)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to unwrap cloud resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool completed: get_cloud_resource_by_id, retrieved resource: %s", resourceID)

	// Convert protobuf to JSON
	// Use protojson for better handling of Any types and other protobuf specifics
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: false, // Skip fields with default values
		UseProtoNames:   true,  // Use proto field names (snake_case)
	}

	resultJSON, err := marshaler.Marshal(unwrappedResource)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal cloud resource: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}
