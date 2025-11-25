package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	grpcclient "github.com/plantoncloud-inc/mcp-server-planton/internal/grpc"
	apiresourcekind "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/commons/apiresource/apiresourcekind"
	cloudresourcekind "github.com/project-planton/project-planton/apis/org/project_planton/shared/cloudresourcekind"
)

// CreateLookupCloudResourceByNameTool creates the MCP tool definition for looking up a cloud resource by name.
func CreateLookupCloudResourceByNameTool() mcp.Tool {
	return mcp.Tool{
		Name: "lookup_cloud_resource_by_name",
		Description: "Find a specific cloud resource by exact name match. " +
			"Requires organization, environment, resource kind, and exact resource name. " +
			"Use this when you know the precise name of the resource you're looking for. " +
			"Returns simplified resource metadata if found.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID or slug (required)",
				},
				"env_name": map[string]interface{}{
					"type":        "string",
					"description": "Environment slug (required)",
				},
				"cloud_resource_kind": map[string]interface{}{
					"type":        "string",
					"description": "CloudResourceKind name (required). Examples: 'AwsEksCluster', 'GcpGkeCluster', 'KubernetesPostgres'",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Exact resource name to search for (required, will be converted to lowercase)",
				},
			},
			Required: []string{"org_id", "env_name", "cloud_resource_kind", "name"},
		},
	}
}

// HandleLookupCloudResourceByName handles the MCP tool invocation for looking up a cloud resource by name.
//
// This function:
//  1. Validates and parses input arguments
//  2. Converts CloudResourceKind name to enum value
//  3. Calls CloudResourceSearchClient to lookup the resource
//  4. Returns simplified JSON response
func HandleLookupCloudResourceByName(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// Extract org_id from arguments
	orgID, ok := arguments["org_id"].(string)
	if !ok || orgID == "" {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "org_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Extract env_name
	envName, ok := arguments["env_name"].(string)
	if !ok || envName == "" {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "env_name is required",
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Extract cloud_resource_kind
	kindStr, ok := arguments["cloud_resource_kind"].(string)
	if !ok || kindStr == "" {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "cloud_resource_kind is required",
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Convert kind string to enum
	kindValue, found := cloudresourcekind.CloudResourceKind_value[kindStr]
	if !found {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: fmt.Sprintf("Unknown CloudResourceKind: %s. Use list_cloud_resource_kinds to see available kinds", kindStr),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}
	kind := cloudresourcekind.CloudResourceKind(kindValue)

	// Extract name
	name, ok := arguments["name"].(string)
	if !ok || name == "" {
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "name is required",
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Convert name to lowercase as per API requirement
	name = strings.ToLower(name)

	log.Printf("Tool invoked: lookup_cloud_resource_by_name, org_id=%s, env=%s, kind=%s, name=%s",
		orgID, envName, kindStr, name)

	// Create gRPC client with user's API key
	client, err := grpcclient.NewCloudResourceSearchClient(
		cfg.PlantonAPIsGRPCEndpoint,
		cfg.PlantonAPIKey,
	)
	if err != nil {
		errResp := ErrorResponse{
			Error:   "CLIENT_ERROR",
			Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}
	defer client.Close()

	// Lookup cloud resource
	record, err := client.LookupCloudResource(ctx, orgID, envName, kind, name)
	if err != nil {
		return handleGRPCError(err, orgID), nil
	}

	// Convert to simplified structure
	resource := CloudResourceSimple{
		ID:                record.GetId(),
		Name:              record.GetName(),
		Slug:              record.GetSlug(),
		Kind:              apiresourcekind.ApiResourceKind_name[int32(record.GetKind())],
		CloudResourceKind: getKindName(int32(record.GetCloudResourceKind())),
		Org:               record.GetOrg(),
		Env:               record.GetEnv(),
		CreatedAt:         formatTimestamp(record.GetCreatedAt()),
		Description:       record.GetDescription(),
		IsReady:           record.GetIsReady(),
		Tags:              record.GetTags(),
	}

	log.Printf("Tool completed: lookup_cloud_resource_by_name, found resource: %s", resource.ID)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(resource, "", "  ")
	if err != nil {
		errResp := ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

