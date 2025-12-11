package apiresource

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode"

	apiresourcekind "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource/apiresourcekind"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"google.golang.org/protobuf/proto"
)

// ApiResourceKindInfo represents simplified API resource kind information for agents.
type ApiResourceKindInfo struct {
	Kind        string `json:"kind"`
	Group       string `json:"group"`
	IdPrefix    string `json:"id_prefix"`
	Description string `json:"description"`
	IsVersioned bool   `json:"is_versioned"`
}

// CreateListApiResourceKindsTool creates the MCP tool definition for listing API resource kinds.
func CreateListApiResourceKindsTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_api_resource_kinds",
		Description: "List all API resource kinds in the Planton Cloud system. " +
			"Returns the complete taxonomy of first-class platform resources including " +
			"organizations, teams, environments, services, credentials, and cloud resources. " +
			"Each kind is returned in snake_case format (e.g., 'organization', 'environment'). " +
			"API resource kinds are distinct from cloud resource kinds - they represent top-level " +
			"platform resources, while cloud resource kinds are the infrastructure resources " +
			"that exist within the cloud_resource API kind.",
		InputSchema: mcp.ToolInputSchema{
			Type:       "object",
			Properties: map[string]interface{}{},
			Required:   []string{},
		},
	}
}

// HandleListApiResourceKinds handles the MCP tool invocation for listing API resource kinds.
//
// This function:
//  1. Iterates through the ApiResourceKind enum values
//  2. Skips the 'unspecified' value (0)
//  3. Extracts metadata from enum options
//  4. Returns JSON array with kind info
func HandleListApiResourceKinds(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_api_resource_kinds")

	// Build list of API resource kinds from enum
	kinds := make([]ApiResourceKindInfo, 0)

	// Iterate through all enum values
	for name, value := range apiresourcekind.ApiResourceKind_value {
		// Skip unspecified
		if value == 0 {
			continue
		}

		kind := apiresourcekind.ApiResourceKind(value)

		// Get metadata for this kind
		meta, err := getKindMeta(kind)
		if err != nil {
			log.Printf("Warning: Could not get metadata for %s: %v", name, err)
			// Include it anyway with basic info
			snakeCaseKind := pascalToSnakeCase(name)
			kinds = append(kinds, ApiResourceKindInfo{
				Kind:        snakeCaseKind,
				Group:       "unknown",
				IdPrefix:    "",
				Description: fmt.Sprintf("API resource: %s", snakeCaseKind),
				IsVersioned: false,
			})
			continue
		}

		snakeCaseKind := pascalToSnakeCase(name)
		groupName := getGroupName(meta.Group)
		description := getDescription(snakeCaseKind, groupName, meta)

		kinds = append(kinds, ApiResourceKindInfo{
			Kind:        snakeCaseKind,
			Group:       groupName,
			IdPrefix:    meta.IdPrefix,
			Description: description,
			IsVersioned: meta.IsVersioned,
		})
	}

	log.Printf("Tool completed: list_api_resource_kinds, returned %d kinds", len(kinds))

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(kinds, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "Failed to marshal API resource kinds",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// getKindMeta extracts the ApiResourceKindMeta from the enum value's options.
func getKindMeta(kind apiresourcekind.ApiResourceKind) (*apiresourcekind.ApiResourceKindMeta, error) {
	// Get the descriptor for the enum value
	enumValueDescriptor := kind.Descriptor().Values().ByNumber(kind.Number())
	if enumValueDescriptor == nil {
		return nil, fmt.Errorf("no descriptor found for kind: %v", kind)
	}

	// Get the options from the enum value descriptor
	options := enumValueDescriptor.Options()
	if options == nil {
		return nil, fmt.Errorf("no options found for kind: %v", kind)
	}

	// Extract the kind_meta field from the options
	meta, ok := proto.GetExtension(options, apiresourcekind.E_KindMeta).(*apiresourcekind.ApiResourceKindMeta)
	if !ok || meta == nil {
		return nil, fmt.Errorf("no kind_meta found for kind: %v", kind)
	}

	return meta, nil
}

// getGroupName converts the ApiResourceGroup enum to a string.
func getGroupName(group apiresourcekind.ApiResourceGroup) string {
	return strings.ToLower(group.String())
}

// getDescription provides a human-readable description.
func getDescription(kindName, groupName string, meta *apiresourcekind.ApiResourceKindMeta) string {
	groupDescriptions := map[string]string{
		"test":             "Test/development resource",
		"iam":              "Identity and Access Management",
		"resource_manager": "Resource organization and lifecycle",
		"connect":          "Provider credentials and connections",
		"infra_hub":        "Infrastructure provisioning and management",
		"billing":          "Billing and cost management",
		"audit":            "Audit logs and compliance",
		"service_hub":      "Service catalog and management",
	}

	groupDesc, ok := groupDescriptions[groupName]
	if !ok {
		groupDesc = "Platform resource"
	}

	// Add display name if available
	if meta.DisplayName != "" {
		return fmt.Sprintf("%s: %s", groupDesc, meta.DisplayName)
	}

	return fmt.Sprintf("%s: %s", groupDesc, kindName)
}

// pascalToSnakeCase converts PascalCase to snake_case.
func pascalToSnakeCase(s string) string {
	var result []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// CreateApiResourceKindsResource creates an MCP resource definition for API resource kinds.
// This resource is automatically available to agents without requiring a tool call.
func CreateApiResourceKindsResource() mcp.Resource {
	return mcp.NewResource(
		"planton://api-resource-kinds",
		"API Resource Kinds",
		mcp.WithResourceDescription("Complete list of API resource kinds (platform resources like organizations, environments, teams, credentials, services, etc.) in snake_case format"),
		mcp.WithMIMEType("application/json"),
	)
}

// HandleReadApiResourceKinds handles reading the API resource kinds MCP resource.
// This provides the same information as list_api_resource_kinds tool but as a resource
// that agents can access automatically.
func HandleReadApiResourceKinds(request mcp.ReadResourceRequest) ([]interface{}, error) {
	log.Printf("Resource read: api-resource-kinds")

	// Build list of API resource kinds from enum
	kinds := make([]ApiResourceKindInfo, 0)

	// Iterate through all enum values
	for name, value := range apiresourcekind.ApiResourceKind_value {
		// Skip unspecified
		if value == 0 {
			continue
		}

		kind := apiresourcekind.ApiResourceKind(value)

		// Get metadata for this kind
		meta, err := getKindMeta(kind)
		if err != nil {
			log.Printf("Warning: Could not get metadata for %s: %v", name, err)
			// Include it anyway with basic info
			snakeCaseKind := pascalToSnakeCase(name)
			kinds = append(kinds, ApiResourceKindInfo{
				Kind:        snakeCaseKind,
				Group:       "unknown",
				IdPrefix:    "",
				Description: fmt.Sprintf("API resource: %s", snakeCaseKind),
				IsVersioned: false,
			})
			continue
		}

		snakeCaseKind := pascalToSnakeCase(name)
		groupName := getGroupName(meta.Group)
		description := getDescription(snakeCaseKind, groupName, meta)

		kinds = append(kinds, ApiResourceKindInfo{
			Kind:        snakeCaseKind,
			Group:       groupName,
			IdPrefix:    meta.IdPrefix,
			Description: description,
			IsVersioned: meta.IsVersioned,
		})
	}

	// Return as JSON
	jsonData, err := json.MarshalIndent(kinds, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal kinds: %w", err)
	}

	log.Printf("Resource read completed: api-resource-kinds, returned %d kinds", len(kinds))

	return []interface{}{
		mcp.TextResourceContents{
			ResourceContents: mcp.ResourceContents{
				URI:      request.Params.URI,
				MIMEType: "application/json",
			},
			Text: string(jsonData),
		},
	}, nil
}




