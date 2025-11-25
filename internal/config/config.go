package config

import (
	"fmt"
	"os"
)

// Environment represents the Planton Cloud environment
type Environment string

const (
	// EnvironmentEnvVar is the environment variable to set the target environment
	EnvironmentEnvVar = "PLANTON_CLOUD_ENVIRONMENT"

	// EndpointOverrideEnvVar allows overriding the endpoint regardless of environment
	EndpointOverrideEnvVar = "PLANTON_APIS_GRPC_ENDPOINT"

	// APIKeyEnvVar is the environment variable for the API key
	APIKeyEnvVar = "PLANTON_API_KEY"

	// Environment values
	EnvironmentLive  Environment = "live"
	EnvironmentTest  Environment = "test"
	EnvironmentLocal Environment = "local"

	// Endpoints for each environment
	LocalEndpoint = "localhost:8080"
	TestEndpoint  = "api.test.planton.cloud:443"
	LiveEndpoint  = "api.live.planton.cloud:443"
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
}

// LoadFromEnv loads configuration from environment variables.
//
// Required environment variables:
//   - PLANTON_API_KEY: User's API key for authentication (can be JWT token or API key)
//
// Optional environment variables:
//   - PLANTON_APIS_GRPC_ENDPOINT: Override endpoint (takes precedence)
//   - PLANTON_CLOUD_ENVIRONMENT: Target environment (live, test, local)
//     Defaults to "live" which uses api.live.planton.cloud:443
//
// Returns an error if PLANTON_API_KEY is missing.
func LoadFromEnv() (*Config, error) {
	apiKey := os.Getenv(APIKeyEnvVar)
	if apiKey == "" {
		return nil, fmt.Errorf(
			"%s environment variable required. "+
				"This should be set by LangGraph when spawning MCP server",
			APIKeyEnvVar,
		)
	}

	endpoint := getEndpoint()

	return &Config{
		PlantonAPIKey:           apiKey,
		PlantonAPIsGRPCEndpoint: endpoint,
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
