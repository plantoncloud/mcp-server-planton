package grpc

import (
	"context"
	"fmt"
	"log"

	cloudresourcesearch "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/search/v1/infrahub/cloudresource"
	"github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/search/v1/apiresource"
	cloudresourcekind "github.com/project-planton/project-planton/apis/org/project_planton/shared/cloudresourcekind"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// CloudResourceSearchClient is a gRPC client for searching Planton Cloud Cloud Resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Search APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type CloudResourceSearchClient struct {
	conn   *grpc.ClientConn
	client cloudresourcesearch.CloudResourceSearchQueryControllerClient
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
		grpc.WithUnaryInterceptor(UserTokenAuthInterceptor(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create cloud resource search query client
	client := cloudresourcesearch.NewCloudResourceSearchQueryControllerClient(conn)

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

