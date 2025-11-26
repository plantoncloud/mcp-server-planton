package environment

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/resourcemanager/clients"
)

// EnvironmentSimple is a simplified representation of an Environment for JSON serialization.
type EnvironmentSimple struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateListEnvironmentsTool creates the MCP tool definition for listing environments.
func CreateListEnvironmentsTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_environments_for_org",
		Description: "List all environments available in an organization. " +
			"Returns environment details including id, slug, name, and description. " +
			"Only returns environments the user has permission to view.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID to query environments for",
				},
			},
			Required: []string{"org_id"},
		},
	}
}

// HandleListEnvironmentsForOrg handles the MCP tool invocation for listing environments.
//
// This function:
//  1. Validates the org_id argument
//  2. Creates EnvironmentClient with user's API key
//  3. Queries Planton Cloud APIs for environments
//  4. Converts protobuf responses to JSON-serializable structs
//  5. Returns formatted response or error message
func HandleListEnvironmentsForOrg(
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

	log.Printf("Tool invoked: list_environments_for_org, org_id=%s", orgID)

	// Create gRPC client with user's API key
	client, err := clients.NewEnvironmentClient(
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
	defer client.Close()

	// Query environments
	environments, err := client.FindByOrg(ctx, orgID)
	if err != nil {
		return errors.HandleGRPCError(err, orgID), nil
	}

	// Convert protobuf objects to JSON-serializable structs
	envList := make([]EnvironmentSimple, 0, len(environments))
	for _, env := range environments {
		envSimple := EnvironmentSimple{
			ID:          env.GetMetadata().GetId(),
			Slug:        env.GetMetadata().GetSlug(),
			Name:        env.GetMetadata().GetName(),
			Description: env.GetSpec().GetDescription(),
		}
		envList = append(envList, envSimple)
	}

	log.Printf(
		"Tool completed: list_environments_for_org, returned %d environments",
		len(envList),
	)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(envList, "", "  ")
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
