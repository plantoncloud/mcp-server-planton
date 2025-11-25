package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	grpcclient "github.com/plantoncloud-inc/mcp-server-planton/internal/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// EnvironmentSimple is a simplified representation of an Environment for JSON serialization.
type EnvironmentSimple struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ErrorResponse represents an error response for MCP tool calls.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	OrgID   string `json:"org_id,omitempty"`
}

// CreateEnvironmentTool creates the MCP tool definition for listing environments.
func CreateEnvironmentTool() mcp.Tool {
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
//  2. Creates EnvironmentClient with user JWT
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
		errResp := ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "org_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool invoked: list_environments_for_org, org_id=%s", orgID)

	// Create gRPC client with user JWT
	client, err := grpcclient.NewEnvironmentClient(
		cfg.PlantonAPIsGRPCEndpoint,
		cfg.UserJWTToken,
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

	// Query environments
	environments, err := client.FindByOrg(ctx, orgID)
	if err != nil {
		// Handle gRPC errors with user-friendly messages
		return handleGRPCError(err, orgID), nil
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

// handleGRPCError converts gRPC errors to user-friendly error responses.
func handleGRPCError(err error, orgID string) *mcp.CallToolResult {
	st, ok := status.FromError(err)
	if !ok {
		errResp := ErrorResponse{
			Error:   "UNKNOWN_ERROR",
			Message: fmt.Sprintf("An unexpected error occurred: %v", err),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON))
	}

	var errResp ErrorResponse
	errResp.OrgID = orgID

	switch st.Code() {
	case codes.Unauthenticated:
		errResp.Error = "UNAUTHENTICATED"
		errResp.Message = "Authentication failed. Your session may have expired. Please refresh and try again."

	case codes.PermissionDenied:
		errResp.Error = "PERMISSION_DENIED"
		errResp.Message = fmt.Sprintf(
			"You don't have permission to view environments for organization '%s'. "+
				"Please contact your organization administrator.",
			orgID,
		)

	case codes.Unavailable:
		errResp.Error = "UNAVAILABLE"
		errResp.Message = "Planton Cloud APIs are currently unavailable. Please try again in a moment."

	case codes.NotFound:
		errResp.Error = "NOT_FOUND"
		errResp.Message = fmt.Sprintf("Organization '%s' not found.", orgID)

	default:
		errResp.Error = st.Code().String()
		errResp.Message = st.Message()
		if errResp.Message == "" {
			errResp.Message = "An unexpected error occurred."
		}
	}

	log.Printf(
		"Tool error: list_environments_for_org, org_id=%s, code=%s, message=%s",
		orgID, errResp.Error, errResp.Message,
	)

	errJSON, _ := json.MarshalIndent(errResp, "", "  ")
	return mcp.NewToolResultText(string(errJSON))
}

