package clients

import (
	"context"
	"fmt"
	"log"

	cloudresourcev1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/infrahub/cloudresource/v1/cloudresourcev1grpc"
	cloudresourcesearchgrpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/search/v1/infrahub/cloudresource/cloudresourcegrpc"
	cloudresourcev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/infrahub/cloudresource/v1"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/apiresource"
	cloudresourcesearch "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/infrahub/cloudresource"
	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CloudResourceQueryClient is a gRPC client for querying Planton Cloud Cloud Resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud InfraHub APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type CloudResourceQueryClient struct {
	conn   *grpc.ClientConn
	client cloudresourcev1grpc.CloudResourceQueryControllerClient
}

// NewCloudResourceQueryClient creates a new Cloud Resource Query gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a CloudResourceQueryClient and any error encountered during connection setup.
func NewCloudResourceQueryClient(grpcEndpoint, apiKey string) (*CloudResourceQueryClient, error) {
	// Create gRPC dial options with auth interceptor
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(auth.UserTokenAuthInterceptor(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create cloud resource query client
	client := cloudresourcev1grpc.NewCloudResourceQueryControllerClient(conn)

	log.Printf("CloudResourceQueryClient initialized for endpoint: %s", grpcEndpoint)

	return &CloudResourceQueryClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetById retrieves a cloud resource by its ID.
//
// This method makes an authenticated gRPC call to Planton Cloud InfraHub APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view the cloud resource.
//
// Args:
//   - ctx: Context for the request
//   - resourceID: Cloud resource ID (e.g., "eks-abc123")
//
// Returns the full CloudResource object or an error.
func (c *CloudResourceQueryClient) GetById(ctx context.Context, resourceID string) (*cloudresourcev1.CloudResource, error) {
	log.Printf("Querying cloud resource by ID: %s", resourceID)

	// Create request
	req := &cloudresourcev1.CloudResourceId{
		Value: resourceID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Get(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying cloud resource %s: %v", resourceID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved cloud resource: %s", resourceID)

	return resp, nil
}

// Close closes the gRPC connection.
func (c *CloudResourceQueryClient) Close() error {
	if c.conn != nil {
		log.Println("Closing CloudResourceQueryClient connection")
		return c.conn.Close()
	}
	return nil
}

// CloudResourceSearchClient is a gRPC client for searching Planton Cloud Cloud Resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Search APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type CloudResourceSearchClient struct {
	conn   *grpc.ClientConn
	client cloudresourcesearchgrpc.CloudResourceSearchQueryControllerClient
}

// NewCloudResourceSearchClient creates a new Cloud Resource Search gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a CloudResourceSearchClient and any error encountered during connection setup.
func NewCloudResourceSearchClient(grpcEndpoint, apiKey string) (*CloudResourceSearchClient, error) {
	// Create gRPC dial options with auth interceptor
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(auth.UserTokenAuthInterceptor(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create cloud resource search query client
	client := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn)

	log.Printf("CloudResourceSearchClient initialized for endpoint: %s", grpcEndpoint)

	return &CloudResourceSearchClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetCloudResourcesCanvasView queries cloud resources for canvas view with filtering.
//
// This method makes an authenticated gRPC call to Planton Cloud Search APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view cloud resources
// in the specified organization and environments.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID
//   - envNames: List of environment slugs (empty = all environments)
//   - kinds: List of CloudResourceKind enums to filter by (empty = all kinds)
//   - searchText: Optional free-text search
//
// Returns ExploreCloudResourcesCanvasViewResponse or an error.
func (c *CloudResourceSearchClient) GetCloudResourcesCanvasView(
	ctx context.Context,
	orgID string,
	envNames []string,
	kinds []cloudresourcekind.CloudResourceKind,
	searchText string,
) (*cloudresourcesearch.ExploreCloudResourcesCanvasViewResponse, error) {
	log.Printf("Querying cloud resources canvas view for org: %s, envs: %v, kinds: %v, searchText: %q",
		orgID, envNames, kinds, searchText)

	// Create request
	req := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:        orgID,
		Envs:       envNames,
		SearchText: searchText,
		Kinds:      kinds,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.GetCloudResourcesCanvasView(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying cloud resources canvas view: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved cloud resources canvas view")

	return resp, nil
}

// LookupCloudResource looks up a specific cloud resource by org, env, kind, and name.
//
// This method makes an authenticated gRPC call to Planton Cloud Search APIs
// to find a specific cloud resource by exact name match.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID
//   - envName: Environment slug
//   - kind: CloudResourceKind enum
//   - name: Resource name (should be lowercase)
//
// Returns ApiResourceSearchRecord or an error.
func (c *CloudResourceSearchClient) LookupCloudResource(
	ctx context.Context,
	orgID string,
	envName string,
	kind cloudresourcekind.CloudResourceKind,
	name string,
) (*apiresource.ApiResourceSearchRecord, error) {
	log.Printf("Looking up cloud resource: org=%s, env=%s, kind=%v, name=%s",
		orgID, envName, kind, name)

	// Create request
	req := &cloudresourcesearch.LookupCloudResourceInput{
		Org:               orgID,
		Env:               envName,
		CloudResourceKind: kind,
		Name:              name,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.LookupCloudResource(ctx, req)
	if err != nil {
		log.Printf("gRPC error looking up cloud resource: %v", err)
		return nil, err
	}

	log.Printf("Successfully found cloud resource: %s", resp.GetId())

	return resp, nil
}

// Close closes the gRPC connection.
func (c *CloudResourceSearchClient) Close() error {
	if c.conn != nil {
		log.Println("Closing CloudResourceSearchClient connection")
		return c.conn.Close()
	}
	return nil
}

