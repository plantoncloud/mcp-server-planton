package mcp

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/mcp/tools"
)

// Server wraps the MCP server instance and configuration.
type Server struct {
	mcpServer *server.MCPServer
	config    *config.Config
}

// NewServer creates a new MCP server instance.
func NewServer(cfg *config.Config) *Server {
	// Create MCP server with server info
	mcpServer := server.NewMCPServer(
		"planton-cloud",
		"0.1.0",
	)

	s := &Server{
		mcpServer: mcpServer,
		config:    cfg,
	}

	// Register tool handlers
	s.registerTools()

	log.Println("MCP server initialized with stdio transport")
	log.Printf("Planton APIs endpoint: %s", cfg.PlantonAPIsGRPCEndpoint)
	log.Println("User API key loaded from environment")

	return s
}

// registerTools registers all available MCP tools with their handlers.
func (s *Server) registerTools() {
	// Register list_environments_for_org tool
	s.mcpServer.AddTool(
		tools.CreateEnvironmentTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := context.Background()
			return tools.HandleListEnvironmentsForOrg(ctx, arguments, s.config)
		},
	)
	log.Println("Registered tool: list_environments_for_org")

	// Register list_cloud_resource_kinds tool
	s.mcpServer.AddTool(
		tools.CreateListCloudResourceKindsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := context.Background()
			return tools.HandleListCloudResourceKinds(ctx, arguments, s.config)
		},
	)
	log.Println("Registered tool: list_cloud_resource_kinds")

	// Register search_cloud_resources tool
	s.mcpServer.AddTool(
		tools.CreateSearchCloudResourcesTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := context.Background()
			return tools.HandleSearchCloudResources(ctx, arguments, s.config)
		},
	)
	log.Println("Registered tool: search_cloud_resources")

	// Register lookup_cloud_resource_by_name tool
	s.mcpServer.AddTool(
		tools.CreateLookupCloudResourceByNameTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := context.Background()
			return tools.HandleLookupCloudResourceByName(ctx, arguments, s.config)
		},
	)
	log.Println("Registered tool: lookup_cloud_resource_by_name")

	// Register get_cloud_resource_by_id tool
	s.mcpServer.AddTool(
		tools.CreateGetCloudResourceByIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := context.Background()
			return tools.HandleGetCloudResourceById(ctx, arguments, s.config)
		},
	)
	log.Println("Registered tool: get_cloud_resource_by_id")
}

// Serve starts the MCP server with stdio transport.
//
// This method blocks until the server is shut down or an error occurs.
func (s *Server) Serve() error {
	log.Println("Starting MCP server on stdio...")
	return server.ServeStdio(s.mcpServer)
}
