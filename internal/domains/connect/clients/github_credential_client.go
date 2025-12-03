package clients

import (
	"context"
	"fmt"
	"log"
	"strings"

	githubcredentialv1grpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/connect/githubcredential/v1/githubcredentialv1grpc"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	githubcredentialv1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/connect/githubcredential/v1"
	commonauth "github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// GithubCredentialClient is a gRPC client for querying Planton Cloud GitHub Credential resources.
//
// This client uses the user's API key (not machine account) to make
// authenticated gRPC calls to Planton Cloud Connect APIs. The APIs validate the
// API key and enforce Fine-Grained Authorization (FGA) checks based on the
// user's actual permissions.
type GithubCredentialClient struct {
	conn   *grpc.ClientConn
	client githubcredentialv1grpc.GithubCredentialQueryControllerClient
}

// GithubQueryClient is a gRPC client for querying GitHub information via Planton Cloud.
type GithubQueryClient struct {
	conn   *grpc.ClientConn
	client githubcredentialv1grpc.GithubQueryControllerClient
}

// NewGithubCredentialClient creates a new GitHub Credential gRPC client.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a GithubCredentialClient and any error encountered during connection setup.
func NewGithubCredentialClient(grpcEndpoint, apiKey string) (*GithubCredentialClient, error) {
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

	// Create GitHub credential query client
	client := githubcredentialv1grpc.NewGithubCredentialQueryControllerClient(conn)

	log.Printf("GithubCredentialClient initialized for endpoint: %s", grpcEndpoint)

	return &GithubCredentialClient{
		conn:   conn,
		client: client,
	}, nil
}

// GetById retrieves a GitHub credential by its ID.
//
// This method makes an authenticated gRPC call to Planton Cloud Connect APIs
// using the user's API key. The API validates the key and checks
// FGA permissions to ensure the user has access to view the credential.
//
// Args:
//   - ctx: Context for the request
//   - credentialID: GitHub credential ID (e.g., "ghcred-abc123")
//
// Returns the full GithubCredential object or an error.
func (c *GithubCredentialClient) GetById(ctx context.Context, credentialID string) (*githubcredentialv1.GithubCredential, error) {
	log.Printf("Querying GitHub credential by ID: %s", credentialID)

	// Create request
	req := &apiresource.ApiResourceId{
		Value: credentialID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.Get(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying GitHub credential %s: %v", credentialID, err)
		return nil, err
	}

	log.Printf("Successfully retrieved GitHub credential: %s", credentialID)

	return resp, nil
}

// GetByOrgBySlug retrieves a GitHub credential by organization ID and slug.
//
// Args:
//   - ctx: Context for the request
//   - orgID: Organization ID
//   - slug: Credential slug/name
//
// Returns the full GithubCredential object or an error.
func (c *GithubCredentialClient) GetByOrgBySlug(ctx context.Context, orgID, slug string) (*githubcredentialv1.GithubCredential, error) {
	log.Printf("Querying GitHub credential by org: %s, slug: %s", orgID, slug)

	// Create request
	req := &apiresource.ApiResourceByOrgBySlugRequest{
		Org:  orgID,
		Slug: slug,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.GetByOrgBySlug(ctx, req)
	if err != nil {
		log.Printf("gRPC error querying GitHub credential by org/slug: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved GitHub credential: %s/%s", orgID, slug)

	return resp, nil
}

// NewGithubCredentialClientFromContext creates a new GitHub Credential gRPC client
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
// Returns a GithubCredentialClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewGithubCredentialClientFromContext(ctx context.Context, grpcEndpoint string) (*GithubCredentialClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewGithubCredentialClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *GithubCredentialClient) Close() error {
	if c.conn != nil {
		log.Println("Closing GithubCredentialClient connection")
		return c.conn.Close()
	}
	return nil
}

// NewGithubQueryClient creates a new GitHub Query gRPC client for GitHub API operations.
//
// Args:
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.cloud:443")
//   - apiKey: User's API key from environment variable (can be JWT token or API key)
//
// Returns a GithubQueryClient and any error encountered during connection setup.
func NewGithubQueryClient(grpcEndpoint, apiKey string) (*GithubQueryClient, error) {
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

	// Create GitHub query client
	client := githubcredentialv1grpc.NewGithubQueryControllerClient(conn)

	log.Printf("GithubQueryClient initialized for endpoint: %s", grpcEndpoint)

	return &GithubQueryClient{
		conn:   conn,
		client: client,
	}, nil
}

// FindGithubRepositories lists all GitHub repositories accessible via a credential.
//
// Args:
//   - ctx: Context for the request
//   - credentialID: GitHub credential ID
//
// Returns a list of GitHub repositories or an error.
func (c *GithubQueryClient) FindGithubRepositories(ctx context.Context, credentialID string) (*githubcredentialv1.FindGithubRepositoriesResponse, error) {
	log.Printf("Finding GitHub repositories for credential: %s", credentialID)

	// Create request
	req := &githubcredentialv1.FindGithubRepositoriesInput{
		GithubCredentialId: credentialID,
	}

	// Make gRPC call (interceptor attaches API key automatically)
	resp, err := c.client.FindGithubRepositories(ctx, req)
	if err != nil {
		log.Printf("gRPC error finding GitHub repositories: %v", err)
		return nil, err
	}

	log.Printf("Successfully retrieved %d GitHub repositories", len(resp.GetRepos()))

	return resp, nil
}

// NewGithubQueryClientFromContext creates a new GitHub Query gRPC client
// using the API key from the request context.
//
// Args:
//   - ctx: Context containing the API key (set by HTTP authentication middleware)
//   - grpcEndpoint: Planton Cloud APIs endpoint (e.g., "localhost:8080" or "api.live.planton.ai:443")
//
// Returns a GithubQueryClient and any error encountered during connection setup.
// Returns an error if no API key is found in the context.
func NewGithubQueryClientFromContext(ctx context.Context, grpcEndpoint string) (*GithubQueryClient, error) {
	apiKey, err := commonauth.GetAPIKey(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get API key from context: %w", err)
	}
	return NewGithubQueryClient(grpcEndpoint, apiKey)
}

// Close closes the gRPC connection.
func (c *GithubQueryClient) Close() error {
	if c.conn != nil {
		log.Println("Closing GithubQueryClient connection")
		return c.conn.Close()
	}
	return nil
}

