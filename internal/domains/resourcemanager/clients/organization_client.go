package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	organizationv1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/resourcemanager/organization/v1/organizationv1grpc"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/protobuf"
	organizationv1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/resourcemanager/organization/v1"
	commonauth "github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// OrganizationClient is a gRPC client for querying Planton Cloud Organization resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type OrganizationClient struct {
	conn   *grpc.ClientConn
	client organizationv1grpc.OrganizationQueryControllerClient
}

// NewOrganizationClient creates a new Organization gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns an OrganizationClient and any error encountered during connection setup.
func NewOrganizationClient(grpcEndpoint, apiKey string) (*OrganizationClient, error) {
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

	// Create organization query client
	client := organizationv1grpc.NewOrganizationQueryControllerClient(conn)

	log.Printf("OrganizationClient initialized for endpoint: %s", grpcEndpoint)

	return &OrganizationClient{
		conn:   conn,
		client: client,
	}, nil
}

// List queries all organizations that the authenticated user is a member of.
//
// This method makes an authenticated gRPC call to Planton Cloud APIs
// using the user's API key. The API validates the key and returns only
// organizations where the user has membership.
//
// Args:
//   - ctx: Context for the request
//
// Returns a list of Organization objects or an error.
func (c *OrganizationClient) List(ctx context.Context) ([]*organizationv1.Organization, error) {
	log.Printf("Querying organizations for authenticated user")

	// Create empty request (user determined by API key in context)
	req := &protobuf.CustomEmpty{}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.FindOrganizations(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying organizations: %v", err)
		return nil, err
	}

	// Extract organizations from response
	organizations := resp.GetEntries()
	log.Printf("Found %d organizations for authenticated user", len(organizations))

	return organizations, nil
}

// NewOrganizationClientFromContext creates a new Organization gRPC client
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
// Returns an OrganizationClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewOrganizationClientFromContext(ctx context.Context, grpcEndpoint string) (*OrganizationClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewOrganizationClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *OrganizationClient) Close() error {
	if c.conn != nil {
		log.Println("Closing OrganizationClient connection")
		return c.conn.Close()
	}
	return nil
}












