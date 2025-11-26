package main

import (
	"context"
	"log"
	"strings"
	"time"

	cloudresourcesearchgrpc "buf.build/gen/go/blintora/apis/grpc/go/ai/planton/search/v1/infrahub/cloudresource/cloudresourcegrpc"
	cloudresourcesearch "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/search/v1/infrahub/cloudresource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	apiKey := "5R-2HCqezbpT5g7r_iKDsnWhJ-v99Eqr5g9qdLVzGbs"
	endpoint := "api.live.planton.ai:443"
	orgID := "planton-cloud"

	log.Printf("=== Testing Header Case Sensitivity ===")
	log.Printf("Token: %s", apiKey)

	var transportCreds credentials.TransportCredentials
	if strings.HasSuffix(endpoint, ":443") {
		transportCreds = credentials.NewTLS(nil)
	} else {
		transportCreds = insecure.NewCredentials()
	}

	// Test with lowercase "authorization"
	log.Printf("\n=== Test 1: lowercase 'authorization' ===")
	opts1 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
	}
	conn1, _ := grpc.NewClient(endpoint, opts1...)
	defer conn1.Close()
	client1 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn1)
	
	md1 := metadata.New(map[string]string{
		"authorization": "Bearer " + apiKey,
	})
	ctx1 := metadata.NewOutgoingContext(context.Background(), md1)
	ctx1, cancel1 := context.WithTimeout(ctx1, 10*time.Second)
	defer cancel1()

	req1 := &cloudresourcesearch.ExploreCloudResourcesRequest{Org: orgID}
	_, err1 := client1.GetCloudResourcesCanvasView(ctx1, req1)
	if err1 != nil {
		log.Printf("❌ FAILED: %v", err1)
	} else {
		log.Printf("✅ SUCCESS")
	}

	// Test with capitalized "Authorization"
	log.Printf("\n=== Test 2: capitalized 'Authorization' ===")
	opts2 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
	}
	conn2, _ := grpc.NewClient(endpoint, opts2...)
	defer conn2.Close()
	client2 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn2)
	
	md2 := metadata.New(map[string]string{
		"Authorization": "Bearer " + apiKey,
	})
	ctx2 := metadata.NewOutgoingContext(context.Background(), md2)
	ctx2, cancel2 := context.WithTimeout(ctx2, 10*time.Second)
	defer cancel2()

	req2 := &cloudresourcesearch.ExploreCloudResourcesRequest{Org: orgID}
	_, err2 := client2.GetCloudResourcesCanvasView(ctx2, req2)
	if err2 != nil {
		log.Printf("❌ FAILED: %v", err2)
	} else {
		log.Printf("✅ SUCCESS")
	}

	// Test with all lowercase bearer
	log.Printf("\n=== Test 3: lowercase 'bearer' ===")
	opts3 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
	}
	conn3, _ := grpc.NewClient(endpoint, opts3...)
	defer conn3.Close()
	client3 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn3)
	
	md3 := metadata.New(map[string]string{
		"authorization": "bearer " + apiKey,
	})
	ctx3 := metadata.NewOutgoingContext(context.Background(), md3)
	ctx3, cancel3 := context.WithTimeout(ctx3, 10*time.Second)
	defer cancel3()

	req3 := &cloudresourcesearch.ExploreCloudResourcesRequest{Org: orgID}
	_, err3 := client3.GetCloudResourcesCanvasView(ctx3, req3)
	if err3 != nil {
		log.Printf("❌ FAILED: %v", err3)
	} else {
		log.Printf("✅ SUCCESS")
	}

	// Test without Bearer prefix
	log.Printf("\n=== Test 4: no 'Bearer' prefix ===")
	opts4 := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
	}
	conn4, _ := grpc.NewClient(endpoint, opts4...)
	defer conn4.Close()
	client4 := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn4)
	
	md4 := metadata.New(map[string]string{
		"authorization": apiKey,
	})
	ctx4 := metadata.NewOutgoingContext(context.Background(), md4)
	ctx4, cancel4 := context.WithTimeout(ctx4, 10*time.Second)
	defer cancel4()

	req4 := &cloudresourcesearch.ExploreCloudResourcesRequest{Org: orgID}
	_, err4 := client4.GetCloudResourcesCanvasView(ctx4, req4)
	if err4 != nil {
		log.Printf("❌ FAILED: %v", err4)
	} else {
		log.Printf("✅ SUCCESS")
	}
}



