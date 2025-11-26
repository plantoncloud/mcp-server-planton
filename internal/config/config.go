package config

import (
	"fmt"
	"os"
)

// Environment represents the Planton Cloud environment
type Environment string

// TransportMode represents the MCP server transport mode
type TransportMode string

const (
	TransportStdio TransportMode = "stdio"
	TransportHTTP  TransportMode = "http"
	TransportBoth  TransportMode = "both"
)

const (
	// EnvironmentEnvVar is the environment variable to set the target environment
	EnvironmentEnvVar = "PLANTON_CLOUD_ENVIRONMENT"

	// EndpointOverrideEnvVar allows overriding the endpoint regardless of environment
	EndpointOverrideEnvVar = "PLANTON_APIS_GRPC_ENDPOINT"

	// APIKeyEnvVar is the environment variable for the API key
	APIKeyEnvVar = "PLANTON_API_KEY"

	// TransportEnvVar specifies the transport mode (stdio, http, or both)
	TransportEnvVar = "PLANTON_MCP_TRANSPORT"

	// HTTPPortEnvVar specifies the HTTP server port
	HTTPPortEnvVar = "PLANTON_MCP_HTTP_PORT"

	// HTTPAuthEnabledEnvVar enables bearer token authentication for HTTP transport
	HTTPAuthEnabledEnvVar = "PLANTON_MCP_HTTP_AUTH_ENABLED"

	// Environment values
	EnvironmentLive  Environment = "live"
	EnvironmentTest  Environment = "test"
	EnvironmentLocal Environment = "local"

	// Endpoints for each environment
	LocalEndpoint = "localhost:8080"
	TestEndpoint  = "api.test.planton.cloud:443"
	LiveEndpoint  = "api.live.planton.ai:443"

	// Default values
	DefaultTransport = "stdio"
	DefaultHTTPPort  = "8080"
)

// Config holds the MCP server configuration loaded from environment variables.
//
// Unlike agent-fleet-worker (which uses machine account), this server
// expects PLANTON_API_KEY to be passed via environment by LangGraph or other MCP clients.
type Config struct {
	// PlantonAPIKey is the user's API key for authentication with Planton Cloud APIs.
	// This can be either a JWT token or an API key from the Planton Cloud console.
	// This is passed by LangGraph via environment when spawning the MCP server.
	PlantonAPIKey string

	// PlantonAPIsGRPCEndpoint is the gRPC endpoint for Planton Cloud APIs.
	// Defaults based on environment or can be overridden.
	PlantonAPIsGRPCEndpoint string

	// Transport specifies the MCP server transport mode (stdio, http, or both)
	Transport TransportMode

	// HTTPPort specifies the port for HTTP transport
	HTTPPort string

	// HTTPAuthEnabled determines if bearer token authentication is required for HTTP
	HTTPAuthEnabled bool
}

// LoadFromEnv loads configuration from environment variables.
//
// Required environment variables:
//   - PLANTON_API_KEY: User's API key for authentication (can be JWT token or API key)
//     Required for STDIO transport mode (used directly for authentication)
//     Optional for HTTP transport mode (extracted from Authorization header per-request)
//
// Optional environment variables:
//   - PLANTON_APIS_GRPC_ENDPOINT: Override endpoint (takes precedence)
//   - PLANTON_CLOUD_ENVIRONMENT: Target environment (live, test, local)
//     Defaults to "live" which uses api.live.planton.cloud:443
//   - PLANTON_MCP_TRANSPORT: Transport mode (stdio, http, both) - defaults to "stdio"
//   - PLANTON_MCP_HTTP_PORT: HTTP server port - defaults to "8080"
//   - PLANTON_MCP_HTTP_AUTH_ENABLED: Enable bearer token auth - defaults to "true"
//
// For STDIO mode, PLANTON_API_KEY from environment is used for all gRPC calls.
// For HTTP mode, PLANTON_API_KEY from Authorization header is extracted per-request,
// enabling proper multi-user support with Fine-Grained Authorization.
func LoadFromEnv() (*Config, error) {
	apiKey := os.Getenv(APIKeyEnvVar)
	transport := getTransport()

	// For STDIO mode, API key is required (used directly for authentication)
	if transport == TransportStdio && apiKey == "" {
		return nil, fmt.Errorf(
			"%s environment variable required for STDIO transport. "+
				"This should be set by LangGraph when spawning MCP server",
			APIKeyEnvVar,
		)
	}

	// For HTTP mode, API key is optional (extracted from Authorization header)
	// If provided, it can be used as a fallback or default key
	if transport == TransportHTTP && apiKey == "" {
		// Log that we're in HTTP mode without default API key
		// This is normal - API keys will come from HTTP Authorization headers
	}

	// For both mode, API key is required for STDIO transport
	if transport == TransportBoth && apiKey == "" {
		return nil, fmt.Errorf(
			"%s environment variable required for STDIO transport in dual-transport mode",
			APIKeyEnvVar,
		)
	}

	endpoint := getEndpoint()
	httpPort := getHTTPPort()
	httpAuthEnabled := getHTTPAuthEnabled()

	return &Config{
		PlantonAPIKey:           apiKey,
		PlantonAPIsGRPCEndpoint: endpoint,
		Transport:               transport,
		HTTPPort:                httpPort,
		HTTPAuthEnabled:         httpAuthEnabled,
	}, nil
}

// getEndpoint determines the gRPC endpoint to use based on environment variables.
// Priority:
// 1. PLANTON_APIS_GRPC_ENDPOINT (explicit override)
// 2. PLANTON_CLOUD_ENVIRONMENT (environment-based selection)
// 3. Default to "live" environment (api.live.planton.cloud:443)
func getEndpoint() string {
	// Check for explicit endpoint override first
	if endpoint := os.Getenv(EndpointOverrideEnvVar); endpoint != "" {
		return endpoint
	}

	// Determine environment and return corresponding endpoint
	env := getEnvironment()
	switch env {
	case EnvironmentTest:
		return TestEndpoint
	case EnvironmentLocal:
		return LocalEndpoint
	case EnvironmentLive:
		fallthrough
	default:
		return LiveEndpoint
	}
}

// getEnvironment returns the configured environment, defaulting to "live"
func getEnvironment() Environment {
	envStr := os.Getenv(EnvironmentEnvVar)
	if envStr == "" {
		return EnvironmentLive
	}

	env := Environment(envStr)
	switch env {
	case EnvironmentLive, EnvironmentTest, EnvironmentLocal:
		return env
	default:
		return EnvironmentLive
	}
}

// getTransport returns the configured transport mode, defaulting to "stdio"
func getTransport() TransportMode {
	transportStr := os.Getenv(TransportEnvVar)
	if transportStr == "" {
		return TransportStdio
	}

	transport := TransportMode(transportStr)
	switch transport {
	case TransportStdio, TransportHTTP, TransportBoth:
		return transport
	default:
		return TransportStdio
	}
}

// getHTTPPort returns the configured HTTP port, defaulting to "8080"
func getHTTPPort() string {
	port := os.Getenv(HTTPPortEnvVar)
	if port == "" {
		return DefaultHTTPPort
	}
	return port
}

// getHTTPAuthEnabled returns whether HTTP authentication is enabled, defaulting to true
func getHTTPAuthEnabled() bool {
	authStr := os.Getenv(HTTPAuthEnabledEnvVar)
	if authStr == "" {
		return true // Default to enabled for security
	}
	return authStr == "true" || authStr == "1"
}
