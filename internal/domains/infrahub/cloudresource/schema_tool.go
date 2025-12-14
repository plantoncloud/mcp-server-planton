package cloudresource

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	crinternal "github.com/plantoncloud/mcp-server-planton/internal/domains/infrahub/cloudresource/internal"
)

// CreateGetCloudResourceSchemaTool creates the MCP tool definition for getting cloud resource schema.
func CreateGetCloudResourceSchemaTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_cloud_resource_schema",
		Description: `Get the complete schema/specification for a cloud resource type.
        
Returns all required and optional fields, their types, validation rules, and descriptions.
Use this before creating a resource to understand what information needs to be collected.

Common cloud_resource_kind values:
- Kubernetes: kubernetes_deployment, kubernetes_postgres, kubernetes_redis, kubernetes_mongodb
- AWS: aws_eks_cluster, aws_rds_instance, aws_rds_cluster, aws_lambda, aws_s3_bucket, aws_vpc
- GCP: gcp_gke_cluster, gcp_cloud_sql, gcp_cloud_function, gcp_vpc

The tool accepts multiple formats and normalizes them:
- "aws_rds_instance" (snake_case)
- "AwsRdsInstance" (PascalCase)
- "AWS RDS Instance" (natural language)

For the complete list of 150+ resource types, use 'list_cloud_resource_kinds' tool.`,
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"cloud_resource_kind": map[string]interface{}{
					"type":        "string",
					"description": "Cloud resource kind enum value (e.g., aws_rds_instance, gcp_gke_cluster)",
				},
			},
			Required: []string{"cloud_resource_kind"},
		},
	}
}

// HandleGetCloudResourceSchema handles the MCP tool invocation for getting cloud resource schema.
//
// This function:
//  1. Extracts and normalizes the cloud_resource_kind argument
//  2. Returns helpful error with suggestions if kind is invalid
//  3. Extracts the schema using protobuf reflection
//  4. Returns rich JSON schema with field information
func HandleGetCloudResourceSchema(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	// Extract cloud_resource_kind from arguments
	kindStr, ok := arguments["cloud_resource_kind"].(string)
	if !ok || kindStr == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "cloud_resource_kind is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool invoked: get_cloud_resource_schema, kind=%s", kindStr)

	// Normalize the kind (handles multiple formats)
	kind, err := crinternal.NormalizeCloudResourceKind(kindStr)
	if err != nil {
		// Kind not found - return error with helpful suggestions
		errResp := map[string]interface{}{
			"error":                     "INVALID_CLOUD_RESOURCE_KIND",
			"message":                   err.Error(),
			"input":                     kindStr,
			"popular_kinds_by_category": crinternal.GetPopularKindsByCategory(),
			"hint":                      "Enable 'list_cloud_resource_kinds' tool to discover all 150+ available types",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Normalized kind: %s -> %s", kindStr, kind.String())

	// Extract schema using protobuf reflection
	schema, err := crinternal.ExtractCloudResourceSchema(kind)
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "SCHEMA_EXTRACTION_ERROR",
			Message: fmt.Sprintf("Failed to extract schema for %s: %v", kind.String(), err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	log.Printf("Tool completed: get_cloud_resource_schema, kind=%s, fields=%d", kind.String(), len(schema.Fields))

	// Return the schema as JSON
	schemaJSON, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal schema: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(schemaJSON)), nil
}
