package infrahub

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/cloudresource"
)

// RegisterTools registers all InfraHub domain tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	log.Println("Registering InfraHub tools...")

	// Register cloud resource tools
	cloudresource.RegisterTools(s, cfg)

	// Future: Register stack job tools
	// stackjob.RegisterTools(s, cfg)

	log.Println("InfraHub tools registration complete")
}










