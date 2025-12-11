package auth

import (
	"context"
	"errors"
	"log"
	"sync"
)

// tokenAuth implements credentials.PerRPCCredentials interface to attach
// authentication tokens to gRPC requests.
//
// This implementation matches the pattern used in Planton Cloud CLI
// (client-apps/cli/internal/cli/backend/backend.go) which has proven
// to work reliably with Planton Cloud APIs.
//
// Key difference from interceptor approach:
//   - PerRPCCredentials is the standard gRPC pattern for auth
//   - Properly integrates with gRPC's credential system
//   - Works reliably with both TLS and insecure connections
type tokenAuth struct {
	token string
}

// NewTokenAuth creates a new tokenAuth instance for use with grpc.WithPerRPCCredentials.
//
// The token should be either a JWT token or an API key from Planton Cloud console.
// It will be attached as "Bearer <token>" in the Authorization header.
func NewTokenAuth(token string) *tokenAuth {
	return &tokenAuth{token: token}
}

// GetRequestMetadata returns the authorization header for each RPC call.
// This method is called by gRPC for every request to attach credentials.
func (t tokenAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": "Bearer " + t.token,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials require TLS.
// Returns false to allow both secure (TLS) and insecure (local dev) connections.
// The actual transport security is determined by the WithTransportCredentials dial option.
func (tokenAuth) RequireTransportSecurity() bool {
	return false
}

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

const apiKeyContextKey contextKey = "planton-api-key"

// WithAPIKey adds an API key to the context for use in downstream requests.
// This is used in HTTP transport mode to pass per-user API keys from HTTP headers
// to gRPC clients.
//
// Example:
//
//	ctx := auth.WithAPIKey(r.Context(), "user-api-key")
//	client, err := NewClientFromContext(ctx, endpoint)
func WithAPIKey(ctx context.Context, apiKey string) context.Context {
	return context.WithValue(ctx, apiKeyContextKey, apiKey)
}

// GetAPIKey retrieves the API key from the context.
// Returns an error if no API key is found in the context.
//
// This is used by gRPC clients to extract the per-user API key for authentication
// with Planton Cloud APIs, enabling proper Fine-Grained Authorization per user.
//
// Example:
//
//	apiKey, err := auth.GetAPIKey(ctx)
//	if err != nil {
//	    return nil, fmt.Errorf("no API key in context: %w", err)
//	}
func GetAPIKey(ctx context.Context) (string, error) {
	apiKey, ok := ctx.Value(apiKeyContextKey).(string)
	if !ok || apiKey == "" {
		return "", errors.New("no API key found in context")
	}
	return apiKey, nil
}

// apiKeyStore provides a simple storage for API keys from HTTP requests.
// This is a workaround for mcp-go's AddTool not supporting context parameters.
//
// Since SSE connections are typically single-threaded (one request at a time per connection),
// we store the API key when a request comes in and retrieve it in tool handlers.
//
// Limitations:
//   - Race conditions possible with concurrent requests (rare in SSE)
//   - Not suitable for high-concurrency scenarios
//   - Better solution: upstream fix to mcp-go to support context in AddTool
type apiKeyStore struct {
	mu         sync.RWMutex
	currentKey string
}

var globalAPIKeyStore = &apiKeyStore{}

// SetCurrentAPIKey stores the API key for the current request context.
// Called by HTTP proxy before forwarding requests to internal SSE server.
func SetCurrentAPIKey(apiKey string) {
	globalAPIKeyStore.mu.Lock()
	defer globalAPIKeyStore.mu.Unlock()
	globalAPIKeyStore.currentKey = apiKey
	log.Printf("API key stored for current request")
}

// getCurrentAPIKey retrieves the API key for the current request context.
// Called by tool handlers to get user's API key.
func getCurrentAPIKey() string {
	globalAPIKeyStore.mu.RLock()
	defer globalAPIKeyStore.mu.RUnlock()
	return globalAPIKeyStore.currentKey
}

// GetContextWithAPIKey creates a context with the API key from the current request.
// This is used in tool handlers to create authenticated gRPC clients.
//
// This is a workaround for mcp-go's AddTool not supporting context parameters.
// The API key is stored when HTTP requests arrive and retrieved here for use
// in gRPC client creation.
//
// Usage in tool handlers:
//
//	ctx := auth.GetContextWithAPIKey(context.Background())
//	client, err := clients.NewClientFromContext(ctx, endpoint)
func GetContextWithAPIKey(baseContext context.Context) context.Context {
	apiKey := getCurrentAPIKey()
	if apiKey != "" {
		return WithAPIKey(baseContext, apiKey)
	}
	return baseContext
}




