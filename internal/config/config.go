package config

import (
	"fmt"
	"os"
)

// Config holds the MCP server configuration loaded from environment variables.
//
// Unlike agent-fleet-worker (which uses machine account), this server
// expects USER_JWT_TOKEN to be passed via environment by LangGraph or other MCP clients.
type Config struct {
	// UserJWTToken is the user's JWT token for authentication with Planton Cloud APIs.
	// This is passed by LangGraph via environment when spawning the MCP server.
	UserJWTToken string

	// PlantonAPIsGRPCEndpoint is the gRPC endpoint for Planton Cloud APIs.
	// Defaults to "localhost:8080" if not set.
	PlantonAPIsGRPCEndpoint string
}

// LoadFromEnv loads configuration from environment variables.
//
// Required environment variables:
//   - USER_JWT_TOKEN: User's JWT token for authentication
//
// Optional environment variables:
//   - PLANTON_APIS_GRPC_ENDPOINT: Planton Cloud APIs gRPC endpoint (default: localhost:8080)
//
// Returns an error if USER_JWT_TOKEN is missing.
func LoadFromEnv() (*Config, error) {
	userJWT := os.Getenv("USER_JWT_TOKEN")
	if userJWT == "" {
		return nil, fmt.Errorf(
			"USER_JWT_TOKEN environment variable required. " +
				"This should be set by LangGraph when spawning MCP server",
		)
	}

	endpoint := os.Getenv("PLANTON_APIS_GRPC_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:8080"
	}

	return &Config{
		UserJWTToken:            userJWT,
		PlantonAPIsGRPCEndpoint: endpoint,
	}, nil
}

