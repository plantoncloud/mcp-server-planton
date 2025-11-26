package cloudresource

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	apiresource "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/clients"
	crinternal "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/cloudresource/internal"
	"google.golang.org/protobuf/encoding/protojson"
)

// CreateCreateCloudResourceTool creates the MCP tool definition for creating a cloud resource.
func CreateCreateCloudResourceTool() mcp.Tool {
	return mcp.Tool{
		Name: "create_cloud_resource",
		Description: `Create a new cloud resource in Planton Cloud.

This tool accepts partial or complete specifications. If required fields are missing or invalid,
it returns validation errors with schema information, allowing you to collect the necessary data and retry.

RECOMMENDED WORKFLOW:
1. Call 'get_cloud_resource_schema' to discover all required fields (optimal)
2. Collect field values from user based on schema
3. Call this tool with complete specification

ALTERNATIVE WORKFLOW (also works):
1. Call this tool with available information
2. If validation fails, errors will indicate missing/invalid fields with their schemas
3. Collect additional information from user
4. Retry with complete information

Common cloud_resource_kind values:
- kubernetes_deployment, kubernetes_postgres, kubernetes_redis
- aws_eks_cluster, aws_rds, aws_lambda, aws_s3_bucket  
- gcp_gke_cluster, gcp_cloud_sql, gcp_cloud_function`,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"cloud_resource_kind": map[string]interface{}{
					"type":        "string",
					"description": "Resource type enum (e.g., aws_rds, kubernetes_postgres)",
				},
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID or slug",
				},
				"env_name": map[string]interface{}{
					"type":        "string",
					"description": "Environment slug (e.g., dev, staging, prod)",
				},
				"resource_name": map[string]interface{}{
					"type":        "string",
					"description": "Resource name (must be unique within environment)",
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Optional description for the resource",
				},
				"tags": map[string]interface{}{
					"type":        "array",
					"description": "Optional tags for the resource",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"spec": map[string]interface{}{
					"type":        "object",
					"description": "Resource-specific specification (fields vary by resource type)",
				},
			},
			Required: []string{"cloud_resource_kind", "org_id", "env_name", "resource_name", "spec"},
		},
	}
}

// HandleCreateCloudResource handles the MCP tool invocation for creating a cloud resource.
//
// This function:
//  1. Extracts and validates all input arguments
//  2. Normalizes the cloud_resource_kind
//  3. Builds ApiResourceMetadata
//  4. Wraps spec data into CloudResource using reflection
//  5. Validates the CloudResource (returns helpful errors if validation fails)
//  6. Calls CloudResourceCommandClient to create the resource
//  7. Unwraps and returns the created resource
func HandleCreateCloudResource(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// 1. Extract cloud_resource_kind
	kindStr, ok := arguments["cloud_resource_kind"].(string)
	if !ok || kindStr == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "cloud_resource_kind is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 2. Normalize cloud_resource_kind
	kind, err := crinternal.NormalizeCloudResourceKind(kindStr)
	if err != nil {
		// Return error with suggestions
		errResp := map[string]interface{}{
			"error":                     "INVALID_CLOUD_RESOURCE_KIND",
			"message":                   err.Error(),
			"input":                     kindStr,
			"popular_kinds_by_category": crinternal.GetPopularKindsByCategory(),
			"hint":                      "Call 'get_cloud_resource_schema' with a valid kind to see required fields",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 3. Extract other required arguments
	orgID, ok := arguments["org_id"].(string)
	if !ok || orgID == "" {
		return errorResponse("INVALID_ARGUMENT", "org_id is required"), nil
	}

	envName, ok := arguments["env_name"].(string)
	if !ok || envName == "" {
		return errorResponse("INVALID_ARGUMENT", "env_name is required"), nil
	}

	resourceName, ok := arguments["resource_name"].(string)
	if !ok || resourceName == "" {
		return errorResponse("INVALID_ARGUMENT", "resource_name is required"), nil
	}

	// Normalize resource name to lowercase
	resourceName = strings.ToLower(resourceName)

	// 4. Extract spec data
	specData, ok := arguments["spec"].(map[string]interface{})
	if !ok {
		return errorResponse("INVALID_ARGUMENT", "spec is required and must be an object"), nil
	}

	// 5. Extract optional fields
	description, _ := arguments["description"].(string)
	var tags []string
	if tagsRaw, ok := arguments["tags"].([]interface{}); ok {
		for _, tag := range tagsRaw {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
	}

	log.Printf("Tool invoked: create_cloud_resource, kind=%s, org=%s, env=%s, name=%s",
		kind.String(), orgID, envName, resourceName)

	// 6. Build ApiResourceMetadata
	// Note: Description is not part of ApiResourceMetadata, it's stored in the resource spec
	metadata := &apiresource.ApiResourceMetadata{
		Name: resourceName,
		Org:  orgID,
		Env:  envName,
		Tags: tags,
	}

	// Add description to spec if provided
	if description != "" {
		specData["description"] = description
	}

	// 7. Wrap spec data into CloudResource
	cloudResource, err := crinternal.WrapCloudResource(kind, specData, metadata)
	if err != nil {
		// Wrapping failed - return error with schema guidance
		errResp := map[string]interface{}{
			"error":   "INVALID_SPEC_DATA",
			"message": fmt.Sprintf("Failed to create %s resource: %v", kind.String(), err),
			"hint":    fmt.Sprintf("Call 'get_cloud_resource_schema' with cloud_resource_kind='%s' for the complete schema", kind.String()),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// 8. Create gRPC command client
	client, err := clients.NewCloudResourceCommandClient(
		cfg.PlantonAPIsGRPCEndpoint,
		cfg.PlantonAPIKey,
	)
	if err != nil {
		return errorResponse("CLIENT_ERROR", fmt.Sprintf("Failed to create gRPC client: %v", err)), nil
	}
	defer client.Close()

	// 9. Call create RPC
	createdResource, err := client.Create(ctx, cloudResource)
	if err != nil {
		// Handle gRPC errors (includes validation errors from backend)
		return errors.HandleGRPCError(err, orgID), nil
	}

	// 10. Unwrap the created resource
	unwrappedResource, err := crinternal.UnwrapCloudResource(createdResource)
	if err != nil {
		return errorResponse("INTERNAL_ERROR", fmt.Sprintf("Failed to unwrap created resource: %v", err)), nil
	}

	log.Printf("Tool completed: create_cloud_resource, created resource_id=%s",
		createdResource.GetMetadata().GetId())

	// 11. Return the created resource as JSON
	marshaler := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: false,
		UseProtoNames:   true,
	}

	resultJSON, err := marshaler.Marshal(unwrappedResource)
	if err != nil {
		return errorResponse("INTERNAL_ERROR", fmt.Sprintf("Failed to marshal resource: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// errorResponse creates a standard error response
func errorResponse(errorCode, message string) *mcp.CallToolResult {
	errResp := errors.ErrorResponse{
		Error:   errorCode,
		Message: message,
	}
	errJSON, _ := json.MarshalIndent(errResp, "", "  ")
	return mcp.NewToolResultText(string(errJSON))
}
