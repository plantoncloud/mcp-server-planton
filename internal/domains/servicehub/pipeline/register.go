package pipeline

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
)

// RegisterTools registers all pipeline tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerGetPipelineByIdTool(s, cfg)
	registerGetLatestPipelineByServiceIdTool(s, cfg)
	registerGetPipelineBuildLogsTool(s, cfg)

	log.Println("Registered 3 pipeline tools")
}

// registerGetPipelineByIdTool registers the get_pipeline_by_id tool.
func registerGetPipelineByIdTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetPipelineByIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetPipelineById(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_pipeline_by_id")
}

// registerGetLatestPipelineByServiceIdTool registers the get_latest_pipeline_by_service_id tool.
func registerGetLatestPipelineByServiceIdTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetLatestPipelineByServiceIdTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetLatestPipelineByServiceId(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_latest_pipeline_by_service_id")
}

// registerGetPipelineBuildLogsTool registers the get_pipeline_build_logs tool.
func registerGetPipelineBuildLogsTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetPipelineBuildLogsTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetPipelineBuildLogs(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_pipeline_build_logs")
}
