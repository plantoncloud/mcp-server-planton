package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	tektonpipelinev1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/servicehub/tektonpipeline/v1/tektonpipelinev1grpc"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	tektonpipelinev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/servicehub/tektonpipeline/v1"
	commonauth "github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// TektonPipelineClient is a gRPC client for querying Planton Cloud Tekton Pipeline resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Service Hub APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type TektonPipelineClient struct {
	conn   *grpc.ClientConn
	client tektonpipelinev1grpc.TektonPipelineQueryControllerClient
}

// NewTektonPipelineClient creates a new Tekton Pipeline gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a TektonPipelineClient and any error encountered during connection setup.
func NewTektonPipelineClient(grpcEndpoint, apiKey string) (*TektonPipelineClient, error) {
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

	// Create tekton pipeline query client
	client := tektonpipelinev1grpc.NewTektonPipelineQueryControllerClient(conn)

	log.Printf("TektonPipelineClient initialized for endpoint: %s", grpcEndpoint)

	return &TektonPipelineClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetById retrieves a Tekton pipeline by its ID.
//
// This method makes an authenticated gRPC call to Planton Cloud Service Hub APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view the pipeline.
//
// Args:
//   - ctx: Context for the request
//   - pipelineID: Tekton pipeline ID (e.g., "tknpipe-abc123")
//
// Returns the full TektonPipeline object or an error.
func (c *TektonPipelineClient) GetById(ctx context.Context, pipelineID string) (*tektonpipelinev1.TektonPipeline, error) {
	log.Printf("Querying Tekton pipeline by ID: %s", pipelineID)

	// Create request
	req := &apiresource.ApiResourceId{
		Value: pipelineID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Get(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying Tekton pipeline %s: %v", pipelineID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved Tekton pipeline: %s", pipelineID)

	return resp, nil
}

// GetByOrgAndName retrieves a Tekton pipeline by organization ID and name.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID
//   - name: Pipeline name (will be converted to slug for lookup)
//
// Returns the full TektonPipeline object or an error.
func (c *TektonPipelineClient) GetByOrgAndName(ctx context.Context, orgID, name string) (*tektonpipelinev1.TektonPipeline, error) {
	log.Printf("Querying Tekton pipeline by org: %s, name: %s", orgID, name)

	// Create request
	req := &tektonpipelinev1.GetByOrgAndNameInput{
		Org:  orgID,
		Name: name,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.GetByOrgAndName(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying Tekton pipeline by org/name: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved Tekton pipeline: %s/%s", orgID, name)

	return resp, nil
}


// NewTektonPipelineClientFromContext creates a new Tekton Pipeline gRPC client
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
// Returns a TektonPipelineClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewTektonPipelineClientFromContext(ctx context.Context, grpcEndpoint string) (*TektonPipelineClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewTektonPipelineClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *TektonPipelineClient) Close() error {
	if c.conn != nil {
		log.Println("Closing TektonPipelineClient connection")
		return c.conn.Close()
	}
	return nil
}

