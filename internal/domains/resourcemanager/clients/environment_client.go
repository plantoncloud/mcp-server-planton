package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	environmentv1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/resourcemanager/environment/v1/environmentv1grpc"
	environmentv1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/resourcemanager/environment/v1"
	organizationv1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/resourcemanager/organization/v1"
	commonauth "github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// EnvironmentClient is a gRPC client for querying Planton Cloud Environment resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type EnvironmentClient struct {
	conn   *grpc.ClientConn
	client environmentv1grpc.EnvironmentQueryControllerClient
}

// NewEnvironmentClient creates a new Environment gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns an EnvironmentClient and any error encountered during connection setup.
func NewEnvironmentClient(grpcEndpoint, apiKey string) (*EnvironmentClient, error) {
	// Determine transport credentials based on endpoint port
	var transportCreds credentials.TransportCredentials
	if strings.HasSuffix(grpcEndpoint, ":443") {
		// Use TLS for port 443 (production endpoints)
		transportCreds = credentials.NewTLS(nil)
		log.Printf("Using TLS transport for endpoint: %s", grpcEndpoint)
	} else {
		// Use insecure for other ports (local development)
		transportCreds = insecure.NewCredentials()
		log.Printf("Using insecure transport for endpoint: %s", grpcEndpoint)
	}

	// Create gRPC dial options with per-RPC credentials (matches CLI pattern)
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(commonauth.NewTokenAuth(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create environment query client
	client := environmentv1grpc.NewEnvironmentQueryControllerClient(conn)

	log.Printf("EnvironmentClient initialized for endpoint: %s", grpcEndpoint)

	return &EnvironmentClient{
		conn:   conn,
		client: client,
	}, nil
}

// FindByOrg queries all environments for an organization.
//
// This method makes an authenticated gRPC call to Planton Cloud APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view environments
// in the specified organization.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID to query environments for
//
// Returns a list of Environment objects or an error.
func (c *EnvironmentClient) FindByOrg(ctx context.Context, orgID string) ([]*environmentv1.Environment, error) {
	log.Printf("Querying environments for org: %s", orgID)

	// Create request
	req := &organizationv1.OrganizationId{
		Value: orgID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.FindByOrg(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying environments for org %s: %v", orgID, err)
		return nil, err
	}

	// Extract environments from response
	environments := resp.GetEntries()
	log.Printf("Found %d environments for org: %s", len(environments), orgID)

	return environments, nil
}

// NewEnvironmentClientFromContext creates a new Environment gRPC client
// using the API key from the request context.
//
// This constructor is used in HTTP transport mode to create clients with per-user API keys
// extracted from Authorization headers, enabling proper multi-user support with Fine-Grained
// Authorization.
//
// Args:
//   - ctx: Context containing the API key (set by HTTP authentication middleware)
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.ai:443")
//
// Returns an EnvironmentClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewEnvironmentClientFromContext(ctx context.Context, grpcEndpoint string) (*EnvironmentClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewEnvironmentClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *EnvironmentClient) Close() error {
	if c.conn != nil {
		log.Println("Closing EnvironmentClient connection")
		return c.conn.Close()
	}
	return nil
}
