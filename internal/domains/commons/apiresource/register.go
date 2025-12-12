package apiresource

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// RegisterTools registers all API resource tools and resources with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	// Register resources first (makes them available to agents immediately)
	registerKindsResource(s)

	// Register tools
	registerListKindsTool(s, cfg)

	log.Println("Registered 1 resource and 1 API resource tool")
}

// registerKindsResource registers the API resource kinds MCP resource.
func registerKindsResource(s *server.MCPServer) {
	s.AddResource(
		CreateApiResourceKindsResource(),
		HandleReadApiResourceKinds,
	)
	log.Println("  - planton://api-resource-kinds (resource)")
}

// registerListKindsTool registers the list_api_resource_kinds tool.
func registerListKindsTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListApiResourceKindsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListApiResourceKinds(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_api_resource_kinds")
}







