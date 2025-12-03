package cloudresource

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	apiresourcekind "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource/apiresourcekind"
	cloudresourcesearch "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/infrahub/cloudresource"
	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/clients"
	crinternal "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/cloudresource/internal"
)

// CloudResourceSimple is a simplified representation of a cloud resource for JSON serialization.
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

// CreateSearchCloudResourcesTool creates the MCP tool definition for searching cloud resources.
func CreateSearchCloudResourcesTool() mcp.Tool {
	return mcp.Tool{
		Name: "search_cloud_resources",
		Description: "Search and list cloud resources deployed in an organization. " +
			"Filter by environment(s), resource kind(s), and optional text search. " +
			"Returns simplified resource records with essential metadata. " +
			"Use this to discover what resources exist before fetching full details.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID or slug to query resources for (required)",
				},
				"env_names": map[string]interface{}{
					"type":        "array",
					"description": "List of environment slugs to filter by (optional, empty = all environments)",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"cloud_resource_kinds": map[string]interface{}{
					"type":        "array",
					"description": "List of CloudResourceKind names to filter by (optional, empty = all kinds). Use names like 'AwsEksCluster', 'GcpGkeCluster', 'KubernetesPostgres'",
					"items": map[string]interface{}{
						"type": "string",
					},
				},
				"search_text": map[string]interface{}{
					"type":        "string",
					"description": "Free-text search to filter resources (optional)",
				},
			},
			Required: []string{"org_id"},
		},
	}
}

// HandleSearchCloudResources handles the MCP tool invocation for searching cloud resources.
//
// This function:
//  1. Validates and parses input arguments
//  2. Converts CloudResourceKind names to enum values
//  3. Calls CloudResourceSearchClient to get canvas view
//  4. Flattens nested response structure
//  5. Returns simplified JSON array
func HandleSearchCloudResources(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// Extract org_id from arguments
	orgID, ok := arguments["org_id"].(string)
	if !ok || orgID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "org_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Extract optional env_names
	var envNames []string
	if envNamesRaw, ok := arguments["env_names"].([]interface{}); ok {
		for _, env := range envNamesRaw {
			if envStr, ok := env.(string); ok {
				envNames = append(envNames, envStr)
			}
		}
	}

	// Extract and convert optional cloud_resource_kinds
	var kinds []cloudresourcekind.CloudResourceKind
	if kindsRaw, ok := arguments["cloud_resource_kinds"].([]interface{}); ok {
		for _, kindName := range kindsRaw {
			if kindStr, ok := kindName.(string); ok {
				// Convert string name to enum value using normalization
				// This handles PascalCase, snake_case, natural language, etc.
				if kindValue, err := crinternal.NormalizeCloudResourceKind(kindStr); err == nil {
					kinds = append(kinds, kindValue)
				} else {
					log.Printf("Warning: Unknown CloudResourceKind: %s, error: %v", kindStr, err)
				}
			}
		}
	}

	// Extract optional search_text
	searchText, _ := arguments["search_text"].(string)

	log.Printf("Tool invoked: search_cloud_resources, org_id=%s, envs=%v, kinds=%v, searchText=%q",
		orgID, envNames, kinds, searchText)

	// Create gRPC client with per-user API key from context
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	client, err := clients.NewCloudResourceSearchClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		client, err = clients.NewCloudResourceSearchClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
				OrgID:   orgID,
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer client.Close()

	// Query cloud resources
	resp, err := client.GetCloudResourcesCanvasView(ctx, orgID, envNames, kinds, searchText)
	if err != nil {
		return errors.HandleGRPCError(err, orgID), nil
	}

	// Flatten the nested response structure
	resources := flattenCanvasResponse(resp)

	log.Printf("Tool completed: search_cloud_resources, returned %d resources", len(resources))

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// flattenCanvasResponse flattens the nested CanvasEnvironment structure into a simple array.
func flattenCanvasResponse(resp *cloudresourcesearch.ExploreCloudResourcesCanvasViewResponse) []CloudResourceSimple {
	resources := make([]CloudResourceSimple, 0)

	// Iterate through canvas environments
	for _, canvasEnv := range resp.GetCanvasEnvironments() {
		envSlug := canvasEnv.GetEnvSlug()

		// Iterate through resource kind mapping
		for kindStr, searchRecords := range canvasEnv.GetResourceKindMapping() {
			// Each searchRecords is ApiResourceSearchRecords with Entries[]
			for _, record := range searchRecords.GetEntries() {
				// Convert to simplified structure
				resource := CloudResourceSimple{
					ID:                record.GetId(),
					Name:              record.GetName(),
					Slug:              record.GetSlug(),
					Kind:              apiresourcekind.ApiResourceKind_name[int32(record.GetKind())],
					CloudResourceKind: getKindName(int32(record.GetCloudResourceKind())),
					Org:               record.GetOrg(),
					Env:               envSlug,
					Description:       record.GetDescription(),
					Tags:              record.GetTags(),
				}

				// Handle empty kind string
				if resource.Kind == "" {
					resource.Kind = kindStr
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources
}

// getKindName converts CloudResourceKind enum to string name.
func getKindName(kind int32) string {
	if name, ok := cloudresourcekind.CloudResourceKind_name[kind]; ok {
		return name
	}
	return fmt.Sprintf("Unknown(%d)", kind)
}










