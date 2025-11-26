package main

import (
	"context"
	"log"
	"strings"
	"time"

	cloudresourcesearchgrpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/search/v1/infrahub/cloudresource/cloudresourcegrpc"
	cloudresourcesearch "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/infrahub/cloudresource"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// loggingAuth wraps tokenAuth to log what's being sent
type loggingAuth struct {
	token string
}

func (t loggingAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	headers := map[string]string{
		"Authorization": "Bearer " + t.token,
	}
	log.Printf("🔑 GetRequestMetadata called:")
	log.Printf("   URI: %v", uri)
	log.Printf("   Token (first 20 chars): %s...", t.token[:20])
	log.Printf("   Authorization header: Bearer %s...", t.token[:20])
	log.Printf("   Full headers being sent: %+v", headers)
	return headers, nil
}

func (loggingAuth) RequireTransportSecurity() bool {
	return false
}

// loggingInterceptor logs the actual metadata being sent
func loggingInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Printf("📤 gRPC Interceptor - Outgoing request:")
		log.Printf("   Method: %s", method)
		
		// Check if metadata exists in context
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			log.Printf("   Outgoing metadata: %+v", md)
		} else {
			log.Printf("   ⚠️  No outgoing metadata in context!")
		}
		
		// Call the actual RPC
		err := invoker(ctx, method, req, reply, cc, opts...)
		
		if err != nil {
			log.Printf("❌ gRPC call failed: %v", err)
		} else {
			log.Printf("✅ gRPC call succeeded")
		}
		
		return err
	}
}

func main() {
	// Test token provided by user
	apiKey := "Cr4W_UnpTde83GPh995E1nfxssG5KymE2TWgxodG0bA"
	endpoint := "api.live.planton.ai:443"
	orgID := "planton-cloud"

	log.Printf("=== Token Debug Test ===")
	log.Printf("Endpoint: %s", endpoint)
	log.Printf("Organization: %s", orgID)
	log.Printf("Token length: %d", len(apiKey))
	log.Printf("Token (full): %s", apiKey)
	log.Printf("")

	// Determine transport credentials
	var transportCreds credentials.TransportCredentials
	if strings.HasSuffix(endpoint, ":443") {
		transportCreds = credentials.NewTLS(nil)
		log.Printf("✓ Using TLS transport")
	} else {
		transportCreds = insecure.NewCredentials()
		log.Printf("✓ Using insecure transport")
	}

	// Test 1: Using auth.NewTokenAuth (MCP server's current approach)
	log.Printf("\n=== Test 1: Using auth.NewTokenAuth ===")
	opts1 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(auth.NewTokenAuth(apiKey)),
		grpc.WithChainUnaryInterceptor(loggingInterceptor()),
	}

	conn1, err := grpc.NewClient(endpoint, opts1...)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn1.Close()

	client1 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn1)

	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()

	req1 := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:        orgID,
		Envs:       []string{},
		SearchText: "",
		Kinds:      nil,
	}

	log.Printf("Making gRPC call...")
	resp1, err1 := client1.GetCloudResourcesCanvasView(ctx1, req1)
	if err1 != nil {
		log.Printf("❌ Test 1 FAILED: %v", err1)
	} else {
		log.Printf("✅ Test 1 SUCCEEDED: %d environments", len(resp1.GetCanvasEnvironments()))
	}

	// Test 2: Using loggingAuth wrapper
	log.Printf("\n=== Test 2: Using loggingAuth (with detailed logging) ===")
	opts2 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(loggingAuth{token: apiKey}),
		grpc.WithChainUnaryInterceptor(loggingInterceptor()),
	}

	conn2, err := grpc.NewClient(endpoint, opts2...)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn2.Close()

	client2 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn2)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	req2 := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:        orgID,
		Envs:       []string{},
		SearchText: "",
		Kinds:      nil,
	}

	log.Printf("Making gRPC call...")
	resp2, err2 := client2.GetCloudResourcesCanvasView(ctx2, req2)
	if err2 != nil {
		log.Printf("❌ Test 2 FAILED: %v", err2)
	} else {
		log.Printf("✅ Test 2 SUCCEEDED: %d environments", len(resp2.GetCanvasEnvironments()))
	}

	// Test 3: Manually adding metadata to context
	log.Printf("\n=== Test 3: Manual metadata in context (no PerRPCCredentials) ===")
	opts3 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithChainUnaryInterceptor(loggingInterceptor()),
	}

	conn3, err := grpc.NewClient(endpoint, opts3...)
	if err != nil {
		log.Fatalf("Failed to create gRPC client: %v", err)
	}
	defer conn3.Close()

	client3 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn3)

	// Manually add authorization header to context
	md := metadata.New(map[string]string{
		"authorization": "Bearer " + apiKey,
	})
	ctx3 := metadata.NewOutgoingContext(context.Background(), md)
	ctx3, cancel3 := context.WithTimeout(ctx3, 10*time.Second)
	defer cancel3()

	req3 := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:        orgID,
		Envs:       []string{},
		SearchText: "",
		Kinds:      nil,
	}

	log.Printf("Making gRPC call with manual metadata...")
	log.Printf("Manual metadata: %+v", md)
	resp3, err3 := client3.GetCloudResourcesCanvasView(ctx3, req3)
	if err3 != nil {
		log.Printf("❌ Test 3 FAILED: %v", err3)
	} else {
		log.Printf("✅ Test 3 SUCCEEDED: %d environments", len(resp3.GetCanvasEnvironments()))
	}
}



