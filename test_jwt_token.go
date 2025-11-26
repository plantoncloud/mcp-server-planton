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
)

func main() {
	// JWT token from ~/.planton/cache/auth/live/tokens/default/suresh@planton.ai (updated today)
	jwtToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6IjVZclZ0ODhid2t4akszblBfdGlfbCJ9.eyJpc3MiOiJodHRwczovL3BsYW50b24tY2xvdWQtcHJvZC51cy5hdXRoMC5jb20vIiwic3ViIjoiZ29vZ2xlLW9hdXRoMnwxMTQ2NTcwODY2OTE0NzU0ODA2OTciLCJhdWQiOlsiaHR0cHM6Ly9hcGkucGxhbnRvbi5jbG91ZC8iLCJodHRwczovL3BsYW50b24tY2xvdWQtcHJvZC51cy5hdXRoMC5jb20vdXNlcmluZm8iXSwiaWF0IjoxNzY0MDY3NTE2LCJleHAiOjE3NjQxNTM5MTYsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwgb2ZmbGluZV9hY2Nlc3MiLCJhenAiOiJVb21GRmlxNXNZS0xWZU9RVlNxVkxZeU5yMlFqUzk5OCJ9.lqukhewHiNm1PcpiVpfirX9JMVPT_0SgEtB3rzPHFj0lRY9rTtoVkuBP8bFBpsgJa4WhyV0B9FmULs0ztxEBMKbkssiYPkZnhEZnCumbmFoJoHcL8fEqcGzIEXhucYSCIYtUuYWlBlggIw7xa9dsv3_dDUJl8PedabYcdNkfJEIABg2vrt-8RskB-SkzCwnUzgisUsGNrbwAkIz_LZzmN2UMrbs6KVts38WaeLM9v_91GkLgur9w-lDH35bzKqPB8q8r69F_lXvUk9QjSLJ844qr0AgFGH-aKkMk3j1voY3CEEANZXczFFcsMsXoF3kxhmCai-gC52APGrIKtvuscQ"
	
	endpoint := "api.live.planton.ai:443"
	orgID := "planton-cloud"

	log.Printf("=== Testing with JWT Token ===")
	log.Printf("Token (first 50 chars): %s...", jwtToken[:50])

	var transportCreds credentials.TransportCredentials
	if strings.HasSuffix(endpoint, ":443") {
		transportCreds = credentials.NewTLS(nil)
	} else {
		transportCreds = insecure.NewCredentials()
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(transportCreds),
		grpc.WithPerRPCCredentials(auth.NewTokenAuth(jwtToken)),
	}

	conn, err := grpc.NewClient(endpoint, opts...)
	if err != nil {
		log.Fatalf("Failed to create connection: %v", err)
	}
	defer conn.Close()

	client := cloudresourcesearchgrpc.NewCloudResourceSearchQueryControllerClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &cloudresourcesearch.ExploreCloudResourcesRequest{
		Org:  orgID,
		Envs: []string{},
	}

	resp, err := client.GetCloudResourcesCanvasView(ctx, req)
	if err != nil {
		log.Printf("❌ FAILED: %v", err)
	} else {
		totalResources := 0
		for _, canvasEnv := range resp.GetCanvasEnvironments() {
			for _, searchRecords := range canvasEnv.GetResourceKindMapping() {
				totalResources += len(searchRecords.GetEntries())
			}
		}
		log.Printf("✅ SUCCESS! Found %d cloud resources across %d environments", 
			totalResources, len(resp.GetCanvasEnvironments()))
	}
}

