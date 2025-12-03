package tektonpipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	tektonpipelinev1 "buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/servicehub/tektonpipeline/v1"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

// TektonPipelineDetails contains detailed pipeline information including YAML content.
type TektonPipelineDetails struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Org         string `json:"org,omitempty"`
	YAMLContent string `json:"yaml_content"`
}

// CreateGetTektonPipelineTool creates the MCP tool definition for getting Tekton pipeline details.
func CreateGetTektonPipelineTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_tekton_pipeline",
		Description: "Get complete Tekton pipeline definition including YAML content. " +
			"Use this to see the pipeline structure before customizing it for your service. " +
			"Provide either pipeline_id OR both org_id and name.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline ID (e.g., 'tknpipe-abc123')",
				},
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID (required if using name)",
				},
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline name (required if using org_id, will be converted to slug)",
				},
			},
		},
	}
}

// HandleGetTektonPipeline handles the MCP tool invocation for getting Tekton pipeline details.
func HandleGetTektonPipeline(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_tekton_pipeline")

	// Extract arguments
	pipelineID, hasPipelineID := arguments["pipeline_id"].(string)
	orgID, hasOrgID := arguments["org_id"].(string)
	name, hasName := arguments["name"].(string)

	// Validate arguments - need either pipeline_id OR (org_id + name)
	if !hasPipelineID && (!hasOrgID || !hasName) {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "Provide either pipeline_id OR both org_id and name",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client
	client, err := clients.NewTektonPipelineClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		client, err = clients.NewTektonPipelineClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer client.Close()

	// Query pipeline - either by ID or by org+name
	var pipeline *tektonpipelinev1.TektonPipeline

	if hasPipelineID && pipelineID != "" {
		pipeline, err = client.GetById(ctx, pipelineID)
		if err != nil {
			return errors.HandleGRPCError(err, pipelineID), nil
		}
		log.Printf("Retrieved pipeline by ID: %s", pipelineID)
	} else {
		pipeline, err = client.GetByOrgAndName(ctx, orgID, name)
		if err != nil {
			return errors.HandleGRPCError(err, fmt.Sprintf("%s/%s", orgID, name)), nil
		}
		log.Printf("Retrieved pipeline by org/name: %s/%s", orgID, name)
	}

	// Convert to detailed struct with YAML content
	pipelineDetails := TektonPipelineDetails{
		ID:          pipeline.GetMetadata().GetId(),
		Slug:        pipeline.GetMetadata().GetSlug(),
		Name:        pipeline.GetMetadata().GetName(),
		Description: pipeline.GetSpec().GetDescription(),
		Org:         pipeline.GetMetadata().GetOrg(),
		YAMLContent: pipeline.GetSpec().GetYamlContent(),
	}

	log.Printf("Tool completed: get_tekton_pipeline")

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(pipelineDetails, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}
