package main

import (
	"context"
	"log"
	"time"

	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/clients"
)

func main() {
	// Test token - API key after fix
	apiKey := "dQ36m7vRJmiRhHunGJEaZEHQTrJm6ODw2P4z5aeQ1tk"
	endpoint := "api.live.planton.ai:443"
	orgID := "planton-cloud"

	log.Printf("Testing token with endpoint: %s", endpoint)
	log.Printf("Organization: %s", orgID)
	log.Printf("Token (first 10 chars): %s...", apiKey[:10])

	// Create search client
	searchClient, err := clients.NewCloudResourceSearchClient(endpoint, apiKey)
	if err != nil {
		log.Fatalf("Failed to create search client: %v", err)
	}
	defer searchClient.Close()

	// Test search with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("Attempting to search cloud resources...")

	// Call the search API
	resp, err := searchClient.GetCloudResourcesCanvasView(
		ctx,
		orgID,
		[]string{}, // all environments
		nil,        // all kinds
		"",         // no search text
	)

	if err != nil {
		log.Fatalf("Search failed: %v", err)
	}

	// Count total resources across all environments
	totalResources := 0
	for _, canvasEnv := range resp.GetCanvasEnvironments() {
		for _, searchRecords := range canvasEnv.GetResourceKindMapping() {
			totalResources += len(searchRecords.GetEntries())
		}
	}

	log.Printf("✅ SUCCESS! Found %d cloud resources across %d environments", 
		totalResources, len(resp.GetCanvasEnvironments()))
	
	// Print first few resources
	count := 0
	log.Println("\nFirst few resources:")
	for _, canvasEnv := range resp.GetCanvasEnvironments() {
		envSlug := canvasEnv.GetEnvSlug()
		for kindStr, searchRecords := range canvasEnv.GetResourceKindMapping() {
			for _, record := range searchRecords.GetEntries() {
				if count >= 5 {
					goto done
				}
				log.Printf("  - %s (%s) in env: %s", record.GetName(), kindStr, envSlug)
				count++
			}
		}
	}
done:
}

