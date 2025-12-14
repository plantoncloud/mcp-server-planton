package clients

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	pipelinev1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/servicehub/pipeline/v1/pipelinev1grpc"
	pipelinev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/servicehub/pipeline/v1"
	servicev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/servicehub/service/v1"
	commonauth "github.com/plantoncloud/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// PipelineClient is a gRPC client for querying Planton Cloud Pipeline resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Service Hub APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type PipelineClient struct {
	conn   *grpc.ClientConn
	client pipelinev1grpc.PipelineQueryControllerClient
}

// NewPipelineClient creates a new Pipeline gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a PipelineClient and any error encountered during connection setup.
func NewPipelineClient(grpcEndpoint, apiKey string) (*PipelineClient, error) {
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

	// Create gRPC dial options with per-RPC credentials and keepalive settings
	// Keepalive prevents connection drops during long-running streaming operations
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(commonauth.NewTokenAuth(apiKey)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second, // Send pings every 30 seconds if no activity
			Timeout:             10 * time.Second, // Wait 10 seconds for ping ack before considering connection dead
			PermitWithoutStream: true,             // Allow pings even when no streams are active
		}),
	}

	// Establish connection
	conn, err := grpc.NewClient(grpcEndpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	// Create pipeline query client
	client := pipelinev1grpc.NewPipelineQueryControllerClient(conn)

	log.Printf("PipelineClient initialized for endpoint: %s", grpcEndpoint)

	return &PipelineClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetById retrieves a pipeline by its ID.
//
// This method makes an authenticated gRPC call to Planton Cloud Service Hub APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view the pipeline.
//
// Args:
//   - ctx: Context for the request
//   - pipelineID: Pipeline ID (e.g., "pipe-abc123")
//
// Returns the full Pipeline object or an error.
func (c *PipelineClient) GetById(ctx context.Context, pipelineID string) (*pipelinev1.Pipeline, error) {
	log.Printf("Querying pipeline by ID: %s", pipelineID)

	// Create request
	req := &pipelinev1.PipelineId{
		Value: pipelineID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Get(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying pipeline %s: %v", pipelineID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved pipeline: %s", pipelineID)

	return resp, nil
}

// GetLastPipelineByServiceId retrieves the most recent pipeline for a service.
//
// This method makes an authenticated gRPC call to Planton Cloud Service Hub APIs
// to fetch the latest pipeline execution for a given service. The API validates
// the user's API key and checks FGA permissions.
//
// Args:
//   - ctx: Context for the request
//   - serviceID: Service ID (e.g., "svc-abc123")
//
// Returns the full Pipeline object or an error.
func (c *PipelineClient) GetLastPipelineByServiceId(ctx context.Context, serviceID string) (*pipelinev1.Pipeline, error) {
	log.Printf("Querying latest pipeline for service: %s", serviceID)

	// Create request
	req := &servicev1.ServiceId{
		Value: serviceID,
	}

	// Make gRPC call
	resp, err := c.client.GetLastPipelineByServiceId(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying latest pipeline for service %s: %v", serviceID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved latest pipeline for service: %s", serviceID)

	return resp, nil
}

// GetLogStream streams Tekton task logs for a pipeline.
//
// This method makes an authenticated gRPC streaming call to fetch build logs.
// Logs are streamed from Redis for in-progress pipelines or from R2 for completed ones.
//
// Args:
//   - ctx: Context for the request
//   - pipelineID: Pipeline ID (e.g., "pipe-abc123")
//
// Returns a stream client for reading TektonTaskLogEntry messages or an error.
func (c *PipelineClient) GetLogStream(ctx context.Context, pipelineID string) (pipelinev1grpc.PipelineQueryController_GetLogStreamClient, error) {
	log.Printf("Starting log stream for pipeline: %s", pipelineID)

	// Create request
	req := &pipelinev1.PipelineId{
		Value: pipelineID,
	}

	// Make gRPC streaming call
	stream, err := c.client.GetLogStream(ctx, req)
	if err != nil {
		log.Printf("gRPC error starting log stream for pipeline %s: %v", pipelineID, err)
		return nil, err
	}

	log.Printf("Successfully started log stream for pipeline: %s", pipelineID)

	return stream, nil
}

// GetStatusStream streams pipeline status updates for a pipeline.
//
// This method makes an authenticated gRPC streaming call to get real-time status updates.
//
// Args:
//   - ctx: Context for the request
//   - pipelineID: Pipeline ID (e.g., "pipe-abc123")
//
// Returns a stream client for reading PipelineStatus messages or an error.
func (c *PipelineClient) GetStatusStream(ctx context.Context, pipelineID string) (pipelinev1grpc.PipelineQueryController_GetStatusStreamClient, error) {
	log.Printf("Starting status stream for pipeline: %s", pipelineID)

	// Create request
	req := &pipelinev1.PipelineId{
		Value: pipelineID,
	}

	// Make gRPC streaming call
	stream, err := c.client.GetStatusStream(ctx, req)
	if err != nil {
		log.Printf("gRPC error starting status stream for pipeline %s: %v", pipelineID, err)
		return nil, err
	}

	log.Printf("Successfully started status stream for pipeline: %s", pipelineID)

	return stream, nil
}

// NewPipelineClientFromContext creates a new Pipeline gRPC client
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
// Returns a PipelineClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewPipelineClientFromContext(ctx context.Context, grpcEndpoint string) (*PipelineClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewPipelineClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *PipelineClient) Close() error {
	if c.conn != nil {
		log.Println("Closing PipelineClient connection")
		return c.conn.Close()
	}
	return nil
}
