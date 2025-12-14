package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/servicehub/clients"
)

// CreateGetLatestPipelineByServiceIdTool creates the MCP tool definition for getting latest pipeline by service ID.
func CreateGetLatestPipelineByServiceIdTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_latest_pipeline_by_service_id",
		Description: "Get the most recent pipeline execution for a service by service ID. " +
			"Returns the latest pipeline metadata, execution status, and commit details. " +
			"Use this when you have a service ID and want to check the most recent build/deploy status.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (e.g., 'svc-abc123')",
				},
			},
			Required: []string{"service_id"},
		},
	}
}

// HandleGetLatestPipelineByServiceId handles the MCP tool invocation for getting latest pipeline by service ID.
func HandleGetLatestPipelineByServiceId(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_latest_pipeline_by_service_id")

	// Extract service_id from arguments
	serviceID, ok := arguments["service_id"].(string)
	if !ok || serviceID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "service_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client
	client, err := clients.NewPipelineClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		client, err = clients.NewPipelineClient(
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

	// Query latest pipeline for service
	pipeline, err := client.GetLastPipelineByServiceId(ctx, serviceID)
	if err != nil {
		return errors.HandleGRPCError(err, serviceID), nil
	}

	// Convert to simple struct
	pipelineSimple := PipelineSimple{
		ID:        pipeline.GetMetadata().GetId(),
		Slug:      pipeline.GetMetadata().GetSlug(),
		Name:      pipeline.GetMetadata().GetName(),
		Org:       pipeline.GetMetadata().GetOrg(),
		ServiceID: pipeline.GetSpec().GetServiceId(),
	}

	// Add commit info if available
	if gitCommit := pipeline.GetSpec().GetGitCommit(); gitCommit != nil {
		pipelineSimple.CommitSha = gitCommit.GetSha()
		pipelineSimple.Branch = gitCommit.GetBranch()
	}

	// Add timestamps if available
	if audit := pipeline.GetStatus().GetAudit(); audit != nil {
		if statusAudit := audit.GetStatusAudit(); statusAudit != nil {
			if statusAudit.GetCreatedAt() != nil {
				pipelineSimple.CreatedAt = statusAudit.GetCreatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00")
			}
			if statusAudit.GetUpdatedAt() != nil {
				pipelineSimple.UpdatedAt = statusAudit.GetUpdatedAt().AsTime().Format("2006-01-02T15:04:05Z07:00")
			}
		}
	}

	// Add status information
	status := pipeline.GetStatus()
	pipelineSimple.Status = PipelineStatus{
		ProgressStatus: status.GetProgressStatus().String(),
		ProgressResult: status.GetProgressResult().String(),
		StatusReason:   status.GetStatusReason(),
	}

	if status.GetStartTime() != nil {
		pipelineSimple.Status.StartTime = status.GetStartTime().AsTime().Format("2006-01-02T15:04:05Z07:00")
	}
	if status.GetEndTime() != nil {
		pipelineSimple.Status.EndTime = status.GetEndTime().AsTime().Format("2006-01-02T15:04:05Z07:00")
	}

	// Add build stage information if present
	if buildStage := status.GetBuildStage(); buildStage != nil {
		pipelineSimple.Status.BuildStage = PipelineBuildStage{
			Status:       buildStage.GetStatus().String(),
			Result:       buildStage.GetResult().String(),
			StatusReason: buildStage.GetStatusReason(),
		}
		if buildStage.GetStartTime() != nil {
			pipelineSimple.Status.BuildStage.StartTime = buildStage.GetStartTime().AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
		if buildStage.GetEndTime() != nil {
			pipelineSimple.Status.BuildStage.EndTime = buildStage.GetEndTime().AsTime().Format("2006-01-02T15:04:05Z07:00")
		}
	}

	log.Printf("Tool completed: get_latest_pipeline_by_service_id, service: %s, pipeline: %s", serviceID, pipelineSimple.ID)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(pipelineSimple, "", "  ")
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













