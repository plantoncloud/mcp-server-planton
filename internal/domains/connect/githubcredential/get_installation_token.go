package githubcredential

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/connect/clients"
)

// GithubInstallationTokenInfo contains installation token for Git operations.
type GithubInstallationTokenInfo struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
}

// CreateGetGithubInstallationTokenTool creates the MCP tool definition for getting GitHub installation token.
func CreateGetGithubInstallationTokenTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_github_installation_token",
		Description: "Get short-lived GitHub App installation token for git operations. " +
			"Token expires after 1 hour. Use this token to clone repositories, push changes, " +
			"and create pull requests using git or gh CLI.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"credential_id": map[string]interface{}{
					"type":        "string",
					"description": "GitHub credential ID",
				},
			},
			Required: []string{"credential_id"},
		},
	}
}

// HandleGetGithubInstallationToken handles the MCP tool invocation for getting GitHub installation token.
func HandleGetGithubInstallationToken(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_github_installation_token")

	// Extract credential_id from arguments
	credentialID, ok := arguments["credential_id"].(string)
	if !ok || credentialID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "credential_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create GitHub query client
	client, err := clients.NewGithubQueryClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		client, err = clients.NewGithubQueryClient(
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

	// Get installation token
	tokenResp, err := client.GetInstallationToken(ctx, credentialID)
	if err != nil {
		return errors.HandleGRPCError(err, credentialID), nil
	}

	// Format expiry timestamp
	expiresAtStr := tokenResp.GetExpiresAt().AsTime().Format("2006-01-02T15:04:05Z")

	// Convert to info struct
	tokenInfo := GithubInstallationTokenInfo{
		Token:     tokenResp.GetToken(),
		ExpiresAt: expiresAtStr,
	}

	log.Printf("Tool completed: get_github_installation_token, credential: %s", credentialID)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(tokenInfo, "", "  ")
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












