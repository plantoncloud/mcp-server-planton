package mcp

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/server"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

// HTTPServerOptions configures the HTTP server
type HTTPServerOptions struct {
	Port            string
	AuthEnabled     bool
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
	authEnabled := opts.AuthEnabled

	if authEnabled {
		log.Println("Bearer token authentication: ENABLED (per-user API keys from Authorization header)")
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
		proxyHandler = createAuthenticatedProxy(sseServerAddr)
		log.Println("SSE endpoints protected with per-user bearer token authentication")
	} else {
		proxyHandler = createProxy(sseServerAddr)
	}

	// Register catch-all handler that rewrites paths to internal SSE server
	// This allows users to configure just "http://localhost:8080/" without knowing about /sse
	mux.HandleFunc("/", proxyHandler)

	log.Println("MCP endpoints available:")
	log.Println("  - GET  /health   - Health check endpoint")
	if authEnabled {
		log.Println("  - GET  /sse      - SSE connection endpoint (authenticated)")
		log.Println("  - POST /message  - Message endpoint (authenticated)")
	} else {
		log.Println("  - GET  /sse      - SSE connection endpoint")
		log.Println("  - POST /message  - Message endpoint")
	}

	// Create logging middleware
	loggingHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		mux.ServeHTTP(w, r)
	})

	// Create and start HTTP server
	httpServer := &http.Server{
		Addr:              ":" + opts.Port,
		Handler:           loggingHandler,
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
// The proxy extracts the user's API key from the Authorization header and stores it in the
// request context for use by downstream gRPC clients. This enables per-user authentication
// with proper Fine-Grained Authorization.
func createAuthenticatedProxy(targetAddr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract bearer token from Authorization header
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			log.Printf("Authentication failed: Missing Authorization header from %s", r.RemoteAddr)
			http.Error(w, "Missing Authorization header. Include 'Authorization: Bearer YOUR_API_KEY' header.", http.StatusUnauthorized)
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
		if token == "" {
			log.Printf("Authentication failed: Empty bearer token from %s", r.RemoteAddr)
			http.Error(w, "Empty bearer token", http.StatusUnauthorized)
			return
		}

		// Store API key in request context for downstream use by gRPC clients
		ctx := auth.WithAPIKey(r.Context(), token)
		r = r.WithContext(ctx)

		log.Printf("Authentication: Extracted API key from Authorization header for %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)

		// Forward to proxy handler with enriched context
		proxyRequest(w, r, targetAddr)
	}
}

// proxyRequest handles the actual proxying of requests to the internal SSE server.
// It properly handles SSE streaming with flushing for real-time updates.
// It also rewrites internal port references to external port in SSE responses.
func proxyRequest(w http.ResponseWriter, r *http.Request, targetAddr string) {
	// Rewrite path for internal SSE server
	// Users configure http://localhost:8080/ but internal server expects /sse
	internalPath := r.URL.Path

	// Map root path and common MCP client paths to /sse
	if internalPath == "/" || internalPath == "" {
		internalPath = "/sse"
	}

	// Create proxy request to internal SSE server
	proxyURL := "http://" + targetAddr + internalPath
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
	// and rewrite internal port references to external port
	if flusher, ok := w.(http.Flusher); ok {
		// Stream response body
		buf := make([]byte, 4096)
		for {
			n, err := resp.Body.Read(buf)
			if n > 0 {
				// Rewrite internal port (18080) to external port (from Host header)
				data := buf[:n]
				dataStr := string(data)
				// Replace localhost:18080 with the external host
				if strings.Contains(dataStr, "localhost:18080") {
					host := r.Host
					if host == "" {
						host = "localhost:8080"
					}
					dataStr = strings.ReplaceAll(dataStr, "localhost:18080", host)
					data = []byte(dataStr)
				}

				if _, writeErr := w.Write(data); writeErr != nil {
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
// With per-user authentication, each user's API key is extracted from the Authorization
// header rather than using a shared bearer token from the environment.
func DefaultHTTPOptions(cfg *config.Config) HTTPServerOptions {
	return HTTPServerOptions{
		Port:            cfg.HTTPPort,
		AuthEnabled:     cfg.HTTPAuthEnabled,
		BaseURL:         fmt.Sprintf("http://localhost:%s", cfg.HTTPPort),
		ShutdownTimeout: 10 * time.Second,
	}
}
