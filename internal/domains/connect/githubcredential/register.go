package githubcredential

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
)

// RegisterTools registers all GitHub credential tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerGetGithubCredentialForServiceTool(s, cfg)
	registerGetGithubCredentialByOrgBySlugTool(s, cfg)
	registerListGithubRepositoriesTool(s, cfg)
	registerGetGithubInstallationTokenTool(s, cfg)

	log.Println("Registered 4 GitHub credential tools")
}

// registerGetGithubCredentialForServiceTool registers the get_github_credential_for_service tool.
func registerGetGithubCredentialForServiceTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetGithubCredentialForServiceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetGithubCredentialForService(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_github_credential_for_service")
}

// registerGetGithubCredentialByOrgBySlugTool registers the get_github_credential_by_org_by_slug tool.
func registerGetGithubCredentialByOrgBySlugTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetGithubCredentialByOrgBySlugTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetGithubCredentialByOrgBySlug(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_github_credential_by_org_by_slug")
}

// registerListGithubRepositoriesTool registers the list_github_repositories tool.
func registerListGithubRepositoriesTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListGithubRepositoriesTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListGithubRepositories(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_github_repositories")
}

// registerGetGithubInstallationTokenTool registers the get_github_installation_token tool.
func registerGetGithubInstallationTokenTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetGithubInstallationTokenTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetGithubInstallationToken(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_github_installation_token")
}
