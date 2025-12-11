package connect

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/connect/githubcredential"
)

// RegisterTools registers all Connect domain tools with the MCP server.
func RegisterTools(s *server.MCPServer, cfg *config.Config) {
	log.Println("Registering Connect tools...")

	// Register GitHub credential tools
	githubcredential.RegisterTools(s, cfg)

	log.Println("Connect tools registration complete")
}




