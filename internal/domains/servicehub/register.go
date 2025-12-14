package servicehub

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/servicehub/pipeline"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/servicehub/service"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/servicehub/tektonpipeline"
)

// RegisterTools registers all Service Hub domain tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	log.Println("Registering Service Hub tools...")

	// Register service tools
	service.RegisterTools(s, cfg)

	// Register pipeline tools
	pipeline.RegisterTools(s, cfg)

	// Register Tekton pipeline tools
	tektonpipeline.RegisterTools(s, cfg)

	log.Println("Service Hub tools registration complete")
}
