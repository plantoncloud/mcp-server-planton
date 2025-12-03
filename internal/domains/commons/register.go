package commons

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/commons/apiresource"
)

// RegisterTools registers all Commons domain tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	log.Println("Registering Commons tools...")

	// Register API resource tools
	apiresource.RegisterTools(s, cfg)

	log.Println("Commons tools registration complete")
}
