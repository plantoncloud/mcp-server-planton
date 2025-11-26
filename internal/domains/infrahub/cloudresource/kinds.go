package cloudresource

import (
	"context"
	"encoding/json"
	"log"

	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	crinternal "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/infrahub/cloudresource/internal"
)

// CloudResourceKindInfo represents simplified cloud resource kind information for agents.
type CloudResourceKindInfo struct {
	Kind        string `json:"kind"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
}

// CreateListCloudResourceKindsTool creates the MCP tool definition for listing cloud resource kinds.
func CreateListCloudResourceKindsTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_cloud_resource_kinds",
		Description: "List all available cloud resource kinds in the Planton Cloud system. " +
			"Returns the complete taxonomy of deployable infrastructure resource types including " +
			"AWS, GCP, Azure, Kubernetes, and SaaS platform resources. " +
			"Each kind is returned in snake_case format (e.g., 'aws_rds_instance') which can be " +
			"used directly with other tools like 'get_cloud_resource_schema' and 'create_cloud_resource'.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}
}

// HandleListCloudResourceKinds handles the MCP tool invocation for listing cloud resource kinds.
//
// This function:
//  1. Iterates through the CloudResourceKind enum values
//  2. Skips the 'unspecified' value (0)
//  3. Groups kinds by provider based on enum value ranges
//  4. Returns JSON array with kind info
func HandleListCloudResourceKinds(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_cloud_resource_kinds")

	// Build list of cloud resource kinds from enum
	kinds := make([]CloudResourceKindInfo, 0)

	// Iterate through all enum values
	for name, value := range cloudresourcekind.CloudResourceKind_value {
		// Skip unspecified
		if value == 0 {
			continue
		}

		// Determine provider based on enum value ranges (from proto comments)
		provider := getProviderByValue(value)
		snakeCaseKind := crinternal.PascalToSnakeCase(name)
		description := getDescriptionByProvider(provider, snakeCaseKind)

		kinds = append(kinds, CloudResourceKindInfo{
			Kind:        snakeCaseKind,
			Provider:    provider,
			Description: description,
		})
	}

	log.Printf("Tool completed: list_cloud_resource_kinds, returned %d kinds", len(kinds))

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(kinds, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to marshal cloud resource kinds",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// getProviderByValue determines the provider based on enum value range.
func getProviderByValue(value int32) string {
	switch {
	case value >= 1 && value <= 49:
		return "test"
	case value >= 50 && value <= 199:
		return "saas"
	case value >= 200 && value <= 399:
		return "aws"
	case value >= 400 && value <= 599:
		return "azure"
	case value >= 600 && value <= 799:
		return "gcp"
	case value >= 800 && value <= 999:
		return "kubernetes"
	case value >= 1000 && value <= 1199:
		return "civo"
	case value >= 1200 && value <= 1499:
		return "digitalocean"
	case value >= 1500 && value <= 1799:
		return "civo"
	case value >= 1800 && value <= 2099:
		return "cloudflare"
	default:
		return "unknown"
	}
}

// getDescriptionByProvider provides a human-readable description.
func getDescriptionByProvider(provider, name string) string {
	descriptions := map[string]string{
		"test":         "Test/development resource",
		"saas":         "SaaS platform resource",
		"aws":          "Amazon Web Services resource",
		"azure":        "Microsoft Azure resource",
		"gcp":          "Google Cloud Platform resource",
		"kubernetes":   "Kubernetes workload or operator",
		"civo":         "Civo Cloud resource",
		"digitalocean": "DigitalOcean resource",
		"cloudflare":   "Cloudflare resource",
	}

	desc, ok := descriptions[provider]
	if !ok {
		desc = "Cloud resource"
	}

	return desc + ": " + name
}
