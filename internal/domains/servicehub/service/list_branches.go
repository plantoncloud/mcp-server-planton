package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

// BranchList represents a list of Git branches.
type BranchList struct {
	Branches []string `json:"branches"`
}

// CreateListServiceBranchesTool creates the MCP tool definition for listing service branches.
func CreateListServiceBranchesTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_service_branches",
		Description: "List all Git branches for a service's repository. " +
			"Uses the GitHub/GitLab API to fetch branch information.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
			},
			Required: []string{"service_id"},
		},
	}
}

// HandleListServiceBranches handles the MCP tool invocation for listing service branches.
func HandleListServiceBranches(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_service_branches")

	// Extract service_id from arguments
	serviceID, ok := arguments["service_id"].(string)
	if !ok || serviceID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "service_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client
	client, err := clients.NewServiceClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		client, err = clients.NewServiceClient(
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

	// List branches
	branchList, err := client.ListBranches(ctx, serviceID)
	if err != nil {
		return errors.HandleGRPCError(err, serviceID), nil
	}

	// Convert to simple struct
	result := BranchList{
		Branches: branchList.GetEntries(),
	}

	log.Printf(
		"Tool completed: list_service_branches, returned %d branches for service %s",
		len(result.Branches),
		serviceID,
	)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(result, "", "  ")
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
