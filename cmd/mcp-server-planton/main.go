package main

import (
	"log"
	"os"

	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/mcp"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create and start MCP server
	server := mcp.NewServer(cfg)

	// Serve blocks until shutdown or error
	if err := server.Serve(); err != nil {
		log.Fatalf("MCP server error: %v", err)
		os.Exit(1)
	}
}
