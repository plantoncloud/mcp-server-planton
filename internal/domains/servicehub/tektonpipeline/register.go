package tektonpipeline

import (
	"context"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
)

// RegisterTools registers all Tekton pipeline tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	registerGetTektonPipelineTool(s, cfg)

	log.Println("Registered 1 Tekton pipeline tool")
}

// registerGetTektonPipelineTool registers the get_tekton_pipeline tool.
func registerGetTektonPipelineTool(s *server.MCPServer, cfg *config.Config) {
	s.AddTool(
		CreateGetTektonPipelineTool(),
		func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
			ctx := auth.GetContextWithAPIKey(context.Background())
			return HandleGetTektonPipeline(ctx, arguments, cfg)
		},
	)
	log.Println("  - get_tekton_pipeline")
}
