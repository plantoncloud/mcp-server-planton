package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserTokenAuthInterceptor creates a gRPC unary client interceptor that attaches
// the user's API key to all outgoing requests.
//
// This interceptor passes through the user's API key (from environment)
// to Planton Cloud APIs, enabling Fine-Grained Authorization (FGA) checks
// using the user's actual permissions. The API key can be either a JWT token
// or an API key obtained from the Planton Cloud console.
//
// Key Difference from agent-fleet-worker's AuthClientInterceptor:
//   - agent-fleet-worker: Fetches machine account token from Auth0
//   - MCP server: Uses user's API key directly (no token fetching)
func UserTokenAuthInterceptor(apiKey string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Add user's API key to request metadata as Authorization header
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"authorization", fmt.Sprintf("Bearer %s", apiKey),
		)

		// Invoke the actual RPC
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

