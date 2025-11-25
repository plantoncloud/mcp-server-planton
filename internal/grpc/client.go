package grpc

import (
	"context"
	"fmt"
	"log"

	environmentv1 "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/resourcemanager/environment/v1"
	organizationv1 "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/resourcemanager/organization/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// EnvironmentClient is a gRPC client for querying Planton Cloud Environment resources.
//
// This client uses the user's JWT token (not machine account) to make
// authenticated gRPC calls to Planton Cloud APIs. The APIs validate the
// JWT and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type EnvironmentClient struct {
	conn   *grpc.ClientConn
	client environmentv1.EnvironmentQueryControllerClient
}

// NewEnvironmentClient creates a new Environment gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080")
//   - userToken: User's JWT token from environment variable
//
// Returns an EnvironmentClient and any error encountered during connection setup.
func NewEnvironmentClient(grpcEndpoint, userToken string) (*EnvironmentClient, error) {
	// Create gRPC dial options with auth interceptor
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(UserTokenAuthInterceptor(userToken)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create environment query client
	client := environmentv1.NewEnvironmentQueryControllerClient(conn)

	log.Printf("EnvironmentClient initialized for endpoint: %s", grpcEndpoint)

	return &EnvironmentClient{
		conn:   conn,
		client: client,
	}, nil
}

// FindByOrg queries all environments for an organization.
//
// This method makes an authenticated gRPC call to Planton Cloud APIs
// using the user's JWT token. The API validates the JWT and checks
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

	// Make gRPC call (interceptor attaches JWT automatically)
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

// Close closes the gRPC connection.
func (c *EnvironmentClient) Close() error {
	if c.conn != nil {
		log.Println("Closing EnvironmentClient connection")
		return c.conn.Close()
	}
	return nil
}

