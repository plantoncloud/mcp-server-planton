package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// UserTokenAuthInterceptor creates a gRPC unary client interceptor that attaches
// the user's JWT token to all outgoing requests.
//
// This interceptor passes through the user's JWT token (from environment)
// to Planton Cloud APIs, enabling Fine-Grained Authorization (FGA) checks
// using the user's actual permissions.
//
// Key Difference from agent-fleet-worker's AuthClientInterceptor:
//   - agent-fleet-worker: Fetches machine account token from Auth0
//   - MCP server: Uses user JWT directly (no token fetching)
func UserTokenAuthInterceptor(userToken string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Add user JWT to request metadata as Authorization header
		ctx = metadata.AppendToOutgoingContext(
			ctx,
			"authorization", fmt.Sprintf("Bearer %s", userToken),
		)

		// Invoke the actual RPC
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

