package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/mcp"
)

func main() {
	// Set up logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration from environment
	cfg, err := config.LoadFromEnv()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Create MCP server
	server := mcp.NewServer(cfg)

	// Handle OS signals for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start server(s) based on transport configuration
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	switch cfg.Transport {
	case config.TransportStdio:
		// STDIO only mode
		log.Println("Starting in STDIO-only mode")
		if err := server.Serve(); err != nil {
			log.Fatalf("STDIO server error: %v", err)
		}

	case config.TransportHTTP:
		// HTTP only mode
		log.Println("Starting in HTTP-only mode")
		httpOpts := mcp.DefaultHTTPOptions(cfg)
		if err := server.ServeHTTP(httpOpts); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}

	case config.TransportBoth:
		// Both transports - run HTTP in goroutine, STDIO in main
		log.Println("Starting in dual transport mode (STDIO + HTTP)")

		// Start HTTP server in background
		wg.Add(1)
		go func() {
			defer wg.Done()
			httpOpts := mcp.DefaultHTTPOptions(cfg)
			if err := server.ServeHTTP(httpOpts); err != nil {
				errChan <- err
			}
		}()

		// Start STDIO server in background (for dual mode)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := server.Serve(); err != nil {
				errChan <- err
			}
		}()

		// Wait for shutdown signal or error
		select {
		case <-sigChan:
			log.Println("Shutdown signal received, stopping servers...")
		case err := <-errChan:
			log.Printf("Server error: %v", err)
		}

		// Wait for both servers to stop
		wg.Wait()

	default:
		log.Fatalf("Invalid transport mode: %s", cfg.Transport)
	}

	log.Println("MCP server stopped")
}
