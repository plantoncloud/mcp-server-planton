package organization

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// RegisterTools registers all organization tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerListTool(s, cfg)
	// Future: registerGetTool(s, cfg)
	// Future: registerCreateTool(s, cfg)

	log.Println("Registered 1 organization tool")
}

// registerListTool registers the list_organizations tool.
func registerListTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListOrganizationsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListOrganizations(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_organizations")
}
