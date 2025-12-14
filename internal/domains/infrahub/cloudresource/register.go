package cloudresource

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
)

// RegisterTools registers all cloud resource tools and resources with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	// Register resources first (makes them available to agents immediately)
	registerKindsResource(s)

	// Query tools
	registerGetTool(s, cfg)
	registerSearchTool(s, cfg)
	registerLookupTool(s, cfg)
	registerListKindsTool(s, cfg)

	// Schema discovery
	registerGetSchemaTool(s, cfg)

	// Command tools (mutations)
	registerCreateTool(s, cfg)
	registerUpdateTool(s, cfg)
	registerDeleteTool(s, cfg)

	log.Println("Registered 1 resource and 8 cloud resource tools")
}

// registerKindsResource registers the cloud resource kinds MCP resource.
func registerKindsResource(s *server.MCPServer) {
	s.AddResource(
		CreateCloudResourceKindsResource(),
		HandleReadCloudResourceKinds,
	)
	log.Println("  - planton://cloud-resource-kinds (resource)")
}

// registerGetTool registers the get_cloud_resource_by_id tool.
func registerGetTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetCloudResourceByIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetCloudResourceById(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_cloud_resource_by_id")
}

// registerSearchTool registers the search_cloud_resources tool.
func registerSearchTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateSearchCloudResourcesTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleSearchCloudResources(ctx, arguments, cfg)
		},
	)
	log.Println("  - search_cloud_resources")
}

// registerLookupTool registers the lookup_cloud_resource_by_name tool.
func registerLookupTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateLookupCloudResourceByNameTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleLookupCloudResourceByName(ctx, arguments, cfg)
		},
	)
	log.Println("  - lookup_cloud_resource_by_name")
}

// registerListKindsTool registers the list_cloud_resource_kinds tool.
func registerListKindsTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListCloudResourceKindsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleListCloudResourceKinds(ctx, arguments, cfg)
		},
	)
	log.Println("  - list_cloud_resource_kinds")
}

// registerGetSchemaTool registers the get_cloud_resource_schema tool.
func registerGetSchemaTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetCloudResourceSchemaTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetCloudResourceSchema(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_cloud_resource_schema")
}

// registerCreateTool registers the create_cloud_resource tool.
func registerCreateTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateCreateCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleCreateCloudResource(ctx, arguments, cfg)
		},
	)
	log.Println("  - create_cloud_resource")
}

// registerUpdateTool registers the update_cloud_resource tool.
func registerUpdateTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateUpdateCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleUpdateCloudResource(ctx, arguments, cfg)
		},
	)
	log.Println("  - update_cloud_resource")
}

// registerDeleteTool registers the delete_cloud_resource tool.
func registerDeleteTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateDeleteCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleDeleteCloudResource(ctx, arguments, cfg)
		},
	)
	log.Println("  - delete_cloud_resource")
}
