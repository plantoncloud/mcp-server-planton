package mcp

import (
	"fmt"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// HTTPServerOptions configures the HTTP server
type HTTPServerOptions struct {
	Port            string
	AuthEnabled     bool
	BearerToken     string
	BaseURL         string
	ShutdownTimeout time.Duration
}

// ServeHTTP starts the MCP server with HTTP transport using SSE (Server-Sent Events).
//
// This method blocks until the server is shut down or an error occurs.
// It supports stateless HTTP transport with optional bearer token authentication.
//
// Note: SSEServer.Start() creates its own HTTP server. For more control over routing,
// middleware, and custom endpoints, use a custom HTTP server implementation.
func (s *Server) ServeHTTP(opts HTTPServerOptions) error {
	log.Printf("Starting MCP server on HTTP port %s", opts.Port)
	log.Printf("Base URL: %s", opts.BaseURL)
	
	if opts.AuthEnabled {
		log.Println("Bearer token authentication: ENABLED")
		log.Println("Note: Bearer token auth requires custom HTTP server wrapper (not yet implemented)")
		log.Println("For now, starting SSE server without bearer token middleware")
	} else {
		log.Println("Bearer token authentication: DISABLED (not recommended for production)")
	}

	// Create SSE server for MCP HTTP transport
	// The SSEServer creates its own HTTP server and handles /sse and /message endpoints
	sseServer := server.NewSSEServer(s.mcpServer, opts.BaseURL)
	
	log.Println("MCP endpoints available:")
	log.Println("  - GET  /sse      - SSE connection endpoint")
	log.Println("  - POST /message  - Message endpoint")

	// Start SSE server - this blocks until shutdown
	return sseServer.Start(":" + opts.Port)
}

// TODO: Implement custom HTTP server wrapper to support:
// - Bearer token authentication middleware
// - Custom health check endpoint
// - Request logging middleware
// - CORS middleware
// - Metrics endpoint
//
// The current SSEServer.Start() creates its own HTTP server,
// which makes it difficult to integrate custom middleware.
// A custom wrapper would create an http.Server with custom mux
// and manually integrate the SSE handlers from the library.

// DefaultHTTPOptions returns default HTTP server options
func DefaultHTTPOptions(cfg *config.Config) HTTPServerOptions {
	return HTTPServerOptions{
		Port:            cfg.HTTPPort,
		AuthEnabled:     cfg.HTTPAuthEnabled,
		BearerToken:     cfg.HTTPBearerToken,
		BaseURL:         fmt.Sprintf("http://localhost:%s", cfg.HTTPPort),
		ShutdownTimeout: 10 * time.Second,
	}
}

