package resourcemanager

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/resourcemanager/environment"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/resourcemanager/organization"
)

// RegisterTools registers all ResourceManager domain tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	log.Println("Registering ResourceManager tools...")

	// Register environment tools
	environment.RegisterTools(s, cfg)

	// Register organization tools
	organization.RegisterTools(s, cfg)

	log.Println("ResourceManager tools registration complete")
}
