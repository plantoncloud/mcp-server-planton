package environment

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// RegisterTools registers all environment tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerListTool(s, cfg)
	// Future: registerGetTool(s, cfg)
	// Future: registerCreateTool(s, cfg)

	log.Println("Registered 1 environment tool")
}

// registerListTool registers the list_environments_for_org tool.
func registerListTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListEnvironmentsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListEnvironmentsForOrg(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_environments_for_org")
}




