package cloudresource

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// RegisterTools registers all cloud resource tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
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

	log.Println("Registered 8 cloud resource tools")
}

// registerGetTool registers the get_cloud_resource_by_id tool.
func registerGetTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetCloudResourceByIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleGetCloudResourceById(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - get_cloud_resource_by_id")
}

// registerSearchTool registers the search_cloud_resources tool.
func registerSearchTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateSearchCloudResourcesTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleSearchCloudResources(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - search_cloud_resources")
}

// registerLookupTool registers the lookup_cloud_resource_by_name tool.
func registerLookupTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateLookupCloudResourceByNameTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleLookupCloudResourceByName(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - lookup_cloud_resource_by_name")
}

// registerListKindsTool registers the list_cloud_resource_kinds tool.
func registerListKindsTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateListCloudResourceKindsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleListCloudResourceKinds(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - list_cloud_resource_kinds")
}

// registerGetSchemaTool registers the get_cloud_resource_schema tool.
func registerGetSchemaTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetCloudResourceSchemaTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleGetCloudResourceSchema(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - get_cloud_resource_schema")
}

// registerCreateTool registers the create_cloud_resource tool.
func registerCreateTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateCreateCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleCreateCloudResource(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - create_cloud_resource")
}

// registerUpdateTool registers the update_cloud_resource tool.
func registerUpdateTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateUpdateCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleUpdateCloudResource(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - update_cloud_resource")
}

// registerDeleteTool registers the delete_cloud_resource tool.
func registerDeleteTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateDeleteCloudResourceTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			return HandleDeleteCloudResource(context.Background(), arguments, cfg)
		},
	)
	log.Println("  - delete_cloud_resource")
}

