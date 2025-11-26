package main

import (
	"context"
	"log"
	"strings"
	"time"

	apikeyv1 "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/iam/apikey/v1/apikeyv1grpc"
	cloudresourcesearchgrpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/search/v1/infrahub/cloudresource/cloudresourcegrpc"
	cloudresourcesearch "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/infrahub/cloudresource"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	apiKey := "5R-2HCqezbpT5g7r_iKDsnWhJ-v99Eqr5g9qdLVzGbs"
	endpoint := "api.live.planton.ai:443"
	orgID := "planton-cloud"

	log.Printf("=== Testing IAM vs Search Services ===")
	log.Printf("Token: %s...", apiKey[:20])
	log.Printf("Endpoint: %s", endpoint)

	var transportCreds credentials.TransportCredentials
	if strings.HasSuffix(endpoint, ":443") {
		transportCreds = credentials.NewTLS(nil)
	} else {
		transportCreds = insecure.NewCredentials()
	}

	// Test 1: IAM Service (API Key List)
	log.Printf("\n=== Test 1: IAM Service - List API Keys ===")
	opts1 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(auth.NewTokenAuth(apiKey)),
	}
	conn1, err := grpc.NewClient(endpoint, opts1...)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn1.Close()

	iamClient := apikeyv1.NewApiKeyQueryControllerClient(conn1)
	ctx1, cancel1 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel1()

	resp1, err1 := iamClient.FindAll(ctx1, &emptypb.Empty{})
	if err1 != nil {
		log.Printf("❌ IAM Service FAILED: %v", err1)
	} else {
		log.Printf("✅ IAM Service SUCCESS: Found %d API keys", len(resp1.Entries))
	}

	// Test 2: Search Service (Cloud Resources)
	log.Printf("\n=== Test 2: Search Service - List Cloud Resources ===")
	opts2 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(auth.NewTokenAuth(apiKey)),
	}
	conn2, err := grpc.NewClient(endpoint, opts2...)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn2.Close()

	searchClient := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn2)
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()

	req2 := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:  orgID,
		Envs: []string{},
	}
	resp2, err2 := searchClient.GetCloudResourcesCanvasView(ctx2, req2)
	if err2 != nil {
		log.Printf("❌ Search Service FAILED: %v", err2)
	} else {
		log.Printf("✅ Search Service SUCCESS: Found %d environments", len(resp2.GetCanvasEnvironments()))
	}
}



