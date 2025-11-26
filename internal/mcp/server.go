package mcp

import (
	"log"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/resourcemanager"
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

	log.Println("MCP server initialized")
	log.Printf("Transport mode: %s", cfg.Transport)
	log.Printf("Planton APIs endpoint: %s", cfg.PlantonAPIsGRPCEndpoint)
	log.Println("User API key loaded from environment")

	return s
}

// registerTools registers all available MCP tools with their handlers.
func (s *Server) registerTools() {
	log.Println("Registering MCP tools...")

	// Register InfraHub tools
	infrahub.RegisterTools(s.mcpServer, s.config)

	// Register ResourceManager tools
	resourcemanager.RegisterTools(s.mcpServer, s.config)

	log.Println("All tools registered successfully")
}

// Serve starts the MCP server with stdio transport.
//
// This method blocks until the server is shut down or an error occurs.
func (s *Server) Serve() error {
	log.Println("Starting MCP server on stdio...")
	return server.ServeStdio(s.mcpServer)
}
