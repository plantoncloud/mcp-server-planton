package githubcredential

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/connect/clients"
)

// GithubRepositoryInfo contains GitHub repository information.
type GithubRepositoryInfo struct {
	Owner      string `json:"owner"`
	Name       string `json:"name"`
	BrowserURL string `json:"browser_url"`
	CloneURL   string `json:"clone_url"`
}

// CreateListGithubRepositoriesTool creates the MCP tool definition for listing GitHub repositories.
func CreateListGithubRepositoriesTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_github_repositories",
		Description: "List all GitHub repositories accessible via a GitHub credential. " +
			"Useful for discovering available repositories when onboarding a new service.",
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

// HandleListGithubRepositories handles the MCP tool invocation for listing GitHub repositories.
func HandleListGithubRepositories(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_github_repositories")

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

	// Find GitHub repositories
	repoList, err := client.FindGithubRepositories(ctx, credentialID)
	if err != nil {
		return errors.HandleGRPCError(err, credentialID), nil
	}

	// Convert to simple structs
	repos := make([]GithubRepositoryInfo, 0, len(repoList.GetRepos()))
	for _, repo := range repoList.GetRepos() {
		repoInfo := GithubRepositoryInfo{
			Owner:      repo.GetOwnerName(),
			Name:       repo.GetName(),
			BrowserURL: repo.GetWebUrl(),
			CloneURL:   repo.GetWebUrl(), // Use web URL as clone URL is not provided
		}
		repos = append(repos, repoInfo)
	}

	log.Printf(
		"Tool completed: list_github_repositories, returned %d repositories for credential %s",
		len(repos),
		credentialID,
	)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(repos, "", "  ")
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

