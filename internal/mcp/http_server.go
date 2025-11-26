package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// HTTPServerOptions configures the HTTP server
type HTTPServerOptions struct {
	Port            string
	AuthEnabled     bool
	BearerToken     string // PLANTON_API_KEY used as bearer token for HTTP authentication
	BaseURL         string
	ShutdownTimeout time.Duration
}

// ServeHTTP starts the MCP server with HTTP transport using SSE (Server-Sent Events).
//
// This method blocks until the server is shut down or an error occurs.
// It supports stateless HTTP transport with bearer token authentication and health checks.
//
// Implementation note: The mcp-go library's SSEServer doesn't expose individual handlers,
// so we run the SSE server on an internal port and proxy requests through our custom
// server which adds health checks and optional authentication.
func (s *Server) ServeHTTP(opts HTTPServerOptions) error {
	log.Printf("Starting MCP server on HTTP port %s", opts.Port)
	log.Printf("Base URL: %s", opts.BaseURL)

	// Determine if authentication is enabled
	authEnabled := opts.AuthEnabled && opts.BearerToken != ""
	if opts.AuthEnabled && opts.BearerToken == "" {
		log.Println("WARNING: Auth enabled but no bearer token configured, disabling authentication")
		authEnabled = false
	}

	if authEnabled {
		log.Println("Bearer token authentication: ENABLED")
	} else {
		log.Println("Bearer token authentication: DISABLED (not recommended for production)")
	}

	// Start SSE server on internal port
	internalPort := "18080" // Internal port for SSE server
	sseServerAddr := "localhost:" + internalPort
	sseServer := server.NewSSEServer(s.mcpServer, "http://"+sseServerAddr)

	// Start SSE server in background
	go func() {
		log.Printf("Starting internal SSE server on %s", sseServerAddr)
		if err := sseServer.Start(":" + internalPort); err != nil {
			log.Printf("Internal SSE server error: %v", err)
		}
	}()

	// Give the internal server time to start
	time.Sleep(100 * time.Millisecond)

	// Create proxy server with health check
	mux := http.NewServeMux()

	// Add health check endpoint (no authentication required)
	mux.HandleFunc("/health", healthCheckHandler)

	// Create proxy handler with optional authentication
	var proxyHandler http.HandlerFunc
	if authEnabled {
		proxyHandler = createAuthenticatedProxy(sseServerAddr, opts.BearerToken)
		log.Println("SSE endpoints protected with bearer token authentication")
	} else {
		proxyHandler = createProxy(sseServerAddr)
	}

	// Register proxied endpoints
	mux.HandleFunc("/sse", proxyHandler)
	mux.HandleFunc("/message", proxyHandler)

	log.Println("MCP endpoints available:")
	log.Println("  - GET  /health   - Health check endpoint")
	if authEnabled {
		log.Println("  - GET  /sse      - SSE connection endpoint (authenticated)")
		log.Println("  - POST /message  - Message endpoint (authenticated)")
	} else {
		log.Println("  - GET  /sse      - SSE connection endpoint")
		log.Println("  - POST /message  - Message endpoint")
	}

	// Create and start HTTP server
	httpServer := &http.Server{
		Addr:              ":" + opts.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      0, // No timeout for SSE connections
		IdleTimeout:       120 * time.Second,
	}

	log.Printf("HTTP server listening on %s", httpServer.Addr)
	return httpServer.ListenAndServe()
}

// createProxy creates a reverse proxy handler without authentication.
// The proxy forwards requests to the internal SSE server.
func createProxy(targetAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, targetAddr)
	}
}

// createAuthenticatedProxy creates a reverse proxy handler with bearer token authentication.
// The proxy forwards authenticated requests to the internal SSE server.
func createAuthenticatedProxy(targetAddr, expectedToken string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Validate bearer token
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			log.Printf("Authentication failed: Missing Authorization header from %s", r.RemoteAddr)
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Extract bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			log.Printf("Authentication failed: Invalid Authorization header format from %s", r.RemoteAddr)
			http.Error(w, "Invalid Authorization header format. Expected: Bearer <token>", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token != expectedToken {
			log.Printf("Authentication failed: Invalid bearer token from %s", r.RemoteAddr)
			http.Error(w, "Invalid bearer token", http.StatusUnauthorized)
			return
		}

		log.Printf("Authentication successful for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Forward to proxy handler
		proxyRequest(w, r, targetAddr)
	}
}

// proxyRequest handles the actual proxying of requests to the internal SSE server.
// It properly handles SSE streaming with flushing for real-time updates.
func proxyRequest(w http.ResponseWriter, r *http.Request, targetAddr string) {
	// Create proxy request to internal SSE server
	proxyURL := "http://" + targetAddr + r.URL.Path
	if r.URL.RawQuery != "" {
		proxyURL += "?" + r.URL.RawQuery
	}

	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, proxyURL, r.Body)
	if err != nil {
		log.Printf("Error creating proxy request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy headers (except Authorization since internal server doesn't need it)
	for key, values := range r.Header {
		if key != "Authorization" {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}
	}

	// Forward request to internal SSE server
	client := &http.Client{
		Timeout: 0, // No timeout for SSE connections
	}
	resp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Error proxying request: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Copy status code
	w.WriteHeader(resp.StatusCode)

	// For SSE connections, we need to flush data as it arrives
	if flusher, ok := w.(http.Flusher); ok {
		// Stream response body
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := w.Write(buf[:n]); writeErr != nil {
					return
				}
				flusher.Flush()
			}
			if err != nil {
				return
			}
		}
	} else {
		// Fallback for non-streaming responses
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
	}
}

// healthCheckHandler handles health check requests.
// Returns a simple JSON response with status "ok" and HTTP 200.
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Write JSON response
	response := map[string]string{"status": "ok"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding health check response: %v", err)
	}
}

// DefaultHTTPOptions returns default HTTP server options.
// When authentication is enabled, the PLANTON_API_KEY is used as the bearer token.
func DefaultHTTPOptions(cfg *config.Config) HTTPServerOptions {
	return HTTPServerOptions{
		Port:            cfg.HTTPPort,
		AuthEnabled:     cfg.HTTPAuthEnabled,
		BearerToken:     cfg.PlantonAPIKey, // Use API key as bearer token
		BaseURL:         fmt.Sprintf("http://localhost:%s", cfg.HTTPPort),
		ShutdownTimeout: 10 * time.Second,
	}
}
