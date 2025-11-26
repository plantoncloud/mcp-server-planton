package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	cloudresourcev1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/infrahub/cloudresource/v1/cloudresourcev1grpc"
	apiresource "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	cloudresourcev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/infrahub/cloudresource/v1"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// CloudResourceCommandClient is a gRPC client for command operations on Planton Cloud Resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud InfraHub APIs for creating,
// updating, and deleting cloud resources. The APIs validate the API key
// and enforce Fine-Grained Authorization (FGA) checks based on the user's
// actual permissions.
type CloudResourceCommandClient struct {
	conn   *grpc.ClientConn
	client cloudresourcev1grpc.CloudResourceCommandControllerClient
}

// NewCloudResourceCommandClient creates a new Cloud Resource Command gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a CloudResourceCommandClient and any error encountered during connection setup.
func NewCloudResourceCommandClient(grpcEndpoint, apiKey string) (*CloudResourceCommandClient, error) {
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
		grpc.WithPerRPCCredentials(auth.NewTokenAuth(apiKey)),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create cloud resource command client
	client := cloudresourcev1grpc.NewCloudResourceCommandControllerClient(conn)

	log.Printf("CloudResourceCommandClient initialized for endpoint: %s", grpcEndpoint)

	return &CloudResourceCommandClient{
		conn:   conn,
		client: client,
	}, nil
}

// Create creates a new cloud resource.
//
// This method makes an authenticated gRPC call to Planton Cloud InfraHub APIs
// using the user's API key. The API validates the key and checks FGA permissions
// to ensure the user has permission to create cloud resources in the specified environment.
//
// Args:
//   - ctx: Context for the request
//   - resource: The CloudResource to create (with metadata, spec, etc.)
//
// Returns the created CloudResource object or an error.
func (c *CloudResourceCommandClient) Create(ctx context.Context, resource *cloudresourcev1.CloudResource) (*cloudresourcev1.CloudResource, error) {
	log.Printf("Creating cloud resource: kind=%s, name=%s", 
		resource.GetSpec().GetKind().String(), 
		resource.GetMetadata().GetName())

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Create(ctx, resource)
	if err != nil {
		log.Printf("gRPC error creating cloud resource: %v", err)
		return nil, err
	}

	log.Printf("Successfully created cloud resource: %s", resp.GetMetadata().GetId())

	return resp, nil
}

// Update updates an existing cloud resource.
//
// This method makes an authenticated gRPC call to Planton Cloud InfraHub APIs
// using the user's API key. The API validates the key and checks FGA permissions
// to ensure the user has permission to update the specified cloud resource.
//
// Args:
//   - ctx: Context for the request
//   - resource: The CloudResource to update (must have ID in metadata)
//
// Returns the updated CloudResource object or an error.
func (c *CloudResourceCommandClient) Update(ctx context.Context, resource *cloudresourcev1.CloudResource) (*cloudresourcev1.CloudResource, error) {
	log.Printf("Updating cloud resource: id=%s, kind=%s", 
		resource.GetMetadata().GetId(),
		resource.GetSpec().GetKind().String())

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Update(ctx, resource)
	if err != nil {
		log.Printf("gRPC error updating cloud resource: %v", err)
		return nil, err
	}

	log.Printf("Successfully updated cloud resource: %s", resp.GetMetadata().GetId())

	return resp, nil
}

// Delete deletes an existing cloud resource.
//
// This method makes an authenticated gRPC call to Planton Cloud InfraHub APIs
// using the user's API key. The API validates the key and checks FGA permissions
// to ensure the user has permission to delete the specified cloud resource.
//
// Args:
//   - ctx: Context for the request
//   - resourceID: Cloud resource ID (e.g., "eks-abc123")
//
// Returns the deleted CloudResource object or an error.
func (c *CloudResourceCommandClient) Delete(ctx context.Context, resourceID string) (*cloudresourcev1.CloudResource, error) {
	log.Printf("Deleting cloud resource by ID: %s", resourceID)

	// Create delete request
	req := &apiresource.ApiResourceDeleteInput{
		ResourceId: resourceID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Delete(ctx, req)
	if err != nil {
		log.Printf("gRPC error deleting cloud resource %s: %v", resourceID, err)
		return nil, err
	}

	log.Printf("Successfully deleted cloud resource: %s", resourceID)

	return resp, nil
}

// Close closes the gRPC connection.
func (c *CloudResourceCommandClient) Close() error {
	if c.conn != nil {
		log.Println("Closing CloudResourceCommandClient connection")
		return c.conn.Close()
	}
	return nil
}

