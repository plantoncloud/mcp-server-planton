package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	servicev1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/servicehub/service/v1/servicev1grpc"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/protobuf"
	servicev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/servicehub/service/v1"
	commonauth "github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// ServiceClient is a gRPC client for querying Planton Cloud Service Hub Service resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Service Hub APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type ServiceClient struct {
	conn   *grpc.ClientConn
	client servicev1grpc.ServiceQueryControllerClient
}

// NewServiceClient creates a new Service gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a ServiceClient and any error encountered during connection setup.
func NewServiceClient(grpcEndpoint, apiKey string) (*ServiceClient, error) {
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

	// Create service query client
	client := servicev1grpc.NewServiceQueryControllerClient(conn)

	log.Printf("ServiceClient initialized for endpoint: %s", grpcEndpoint)

	return &ServiceClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetById retrieves a service by its ID.
//
// This method makes an authenticated gRPC call to Planton Cloud Service Hub APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view the service.
//
// Args:
//   - ctx: Context for the request
//   - serviceID: Service ID (e.g., "svc-abc123")
//
// Returns the full Service object or an error.
func (c *ServiceClient) GetById(ctx context.Context, serviceID string) (*servicev1.Service, error) {
	log.Printf("Querying service by ID: %s", serviceID)

	// Create request
	req := &servicev1.ServiceId{
		Value: serviceID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Get(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying service %s: %v", serviceID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved service: %s", serviceID)

	return resp, nil
}

// GetByOrgBySlug retrieves a service by organization ID and slug.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID
//   - slug: Service slug/name
//
// Returns the full Service object or an error.
func (c *ServiceClient) GetByOrgBySlug(ctx context.Context, orgID, slug string) (*servicev1.Service, error) {
	log.Printf("Querying service by org: %s, slug: %s", orgID, slug)

	// Create request
	req := &apiresource.ApiResourceByOrgBySlugRequest{
		Org:  orgID,
		Slug: slug,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.GetByOrgBySlug(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying service by org/slug: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved service: %s/%s", orgID, slug)

	return resp, nil
}

// ListBranches lists all Git branches for a service's repository.
//
// Args:
//   - ctx: Context for the request
//   - serviceID: Service ID
//
// Returns a list of branch names or an error.
func (c *ServiceClient) ListBranches(ctx context.Context, serviceID string) (*protobuf.StringList, error) {
	log.Printf("Listing branches for service: %s", serviceID)

	// Create request
	req := &servicev1.ServiceId{
		Value: serviceID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.ListBranches(ctx, req)
	if err != nil {
		log.Printf("gRPC error listing branches for service %s: %v", serviceID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved %d branches for service %s", len(resp.GetEntries()), serviceID)

	return resp, nil
}

// Find queries services with pagination and filtering.
//
// Args:
//   - ctx: Context for the request
//   - request: Find request with filters
//
// Returns a list of services or an error.
func (c *ServiceClient) Find(ctx context.Context, request *apiresource.FindApiResourcesRequest) (*servicev1.ServiceList, error) {
	log.Printf("Finding services with filters")

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Find(ctx, request)
	if err != nil {
		log.Printf("gRPC error finding services: %v", err)
		return nil, err
	}

	log.Printf("Successfully found %d services", len(resp.GetEntries()))

	return resp, nil
}

// NewServiceClientFromContext creates a new Service gRPC client
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
// Returns a ServiceClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewServiceClientFromContext(ctx context.Context, grpcEndpoint string) (*ServiceClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewServiceClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *ServiceClient) Close() error {
	if c.conn != nil {
		log.Println("Closing ServiceClient connection")
		return c.conn.Close()
	}
	return nil
}
