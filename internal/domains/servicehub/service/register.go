package service

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
)

// RegisterTools registers all service tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerListServicesForOrgTool(s, cfg)
	registerGetServiceByIdTool(s, cfg)
	registerGetServiceByOrgBySlugTool(s, cfg)
	registerListServiceBranchesTool(s, cfg)

	log.Println("Registered 4 service tools")
}

// registerListServicesForOrgTool registers the list_services_for_org tool.
func registerListServicesForOrgTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListServicesForOrgTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListServicesForOrg(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_services_for_org")
}

// registerGetServiceByIdTool registers the get_service_by_id tool.
func registerGetServiceByIdTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetServiceByIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetServiceById(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_service_by_id")
}

// registerGetServiceByOrgBySlugTool registers the get_service_by_org_by_slug tool.
func registerGetServiceByOrgBySlugTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetServiceByOrgBySlugTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetServiceByOrgBySlug(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_service_by_org_by_slug")
}

// registerListServiceBranchesTool registers the list_service_branches tool.
func registerListServiceBranchesTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListServiceBranchesTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListServiceBranches(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_service_branches")
}












