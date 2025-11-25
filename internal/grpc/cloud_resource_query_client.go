package grpc

import (
	"context"
	"fmt"
	"log"

	cloudresourcev1 "github.com/plantoncloud-inc/planton-cloud/apis/stubs/go/ai/planton/infrahub/cloudresource/v1"
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
	client cloudresourcev1.CloudResourceQueryControllerClient
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
		grpc.WithUnaryInterceptor(UserTokenAuthInterceptor(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create cloud resource query client
	client := cloudresourcev1.NewCloudResourceQueryControllerClient(conn)

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

