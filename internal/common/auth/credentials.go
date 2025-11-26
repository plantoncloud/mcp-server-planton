package auth

import (
	"context"
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





