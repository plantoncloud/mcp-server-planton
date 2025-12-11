package organization

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

// OrganizationSimple is a simplified representation of an Organization for JSON serialization.
type OrganizationSimple struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateListOrganizationsTool creates the MCP tool definition for listing organizations.
func CreateListOrganizationsTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_organizations",
		Description: "List all organizations the user is a member of. " +
			"Returns organization details including id, slug, name, and description. " +
			"Only returns organizations the user has permission to view.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
		},
	}
}

// HandleListOrganizations handles the MCP tool invocation for listing organizations.
//
// This function:
//  1. Creates OrganizationClient with user's API key
//  2. Queries Planton Cloud APIs for organizations
//  3. Converts protobuf responses to JSON-serializable structs
//  4. Returns formatted response or error message
func HandleListOrganizations(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_organizations")

	// Create gRPC client with per-user API key from context
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	client, err := clients.NewOrganizationClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		client, err = clients.NewOrganizationClient(
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

	// Query organizations
	organizations, err := client.List(ctx)
	if err != nil {
		return errors.HandleGRPCError(err, ""), nil
	}

	// Convert protobuf objects to JSON-serializable structs
	orgList := make([]OrganizationSimple, 0, len(organizations))
	for _, org := range organizations {
		orgSimple := OrganizationSimple{
			ID:          org.GetMetadata().GetId(),
			Slug:        org.GetMetadata().GetSlug(),
			Name:        org.GetMetadata().GetName(),
			Description: org.GetSpec().GetDescription(),
		}
		orgList = append(orgList, orgSimple)
	}

	log.Printf(
		"Tool completed: list_organizations, returned %d organizations",
		len(orgList),
	)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(orgList, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}




