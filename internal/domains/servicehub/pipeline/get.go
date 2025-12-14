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

// PipelineSimple is a simplified representation of a Pipeline for JSON serialization.
type PipelineSimple struct {
	ID        string         `json:"id"`
	Slug      string         `json:"slug"`
	Name      string         `json:"name"`
	Org       string         `json:"org"`
	ServiceID string         `json:"service_id"`
	CommitSha string         `json:"commit_sha"`
	Branch    string         `json:"branch"`
	Status    PipelineStatus `json:"status"`
	CreatedAt string         `json:"created_at,omitempty"`
	UpdatedAt string         `json:"updated_at,omitempty"`
}

// PipelineStatus contains pipeline execution status information.
type PipelineStatus struct {
	ProgressStatus string             `json:"progress_status"`
	ProgressResult string             `json:"progress_result"`
	StatusReason   string             `json:"status_reason,omitempty"`
	StartTime      string             `json:"start_time,omitempty"`
	EndTime        string             `json:"end_time,omitempty"`
	BuildStage     PipelineBuildStage `json:"build_stage,omitempty"`
}

// PipelineBuildStage contains build stage execution details.
type PipelineBuildStage struct {
	Status       string `json:"status"`
	Result       string `json:"result"`
	StatusReason string `json:"status_reason,omitempty"`
	StartTime    string `json:"start_time,omitempty"`
	EndTime      string `json:"end_time,omitempty"`
}

// CreateGetPipelineByIdTool creates the MCP tool definition for getting pipeline by ID.
func CreateGetPipelineByIdTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_pipeline_by_id",
		Description: "Get detailed information about a pipeline execution by its ID. " +
			"Returns pipeline metadata, execution status, build/deploy progress, and commit details. " +
			"Use this to check pipeline status and investigate build/deploy issues.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline ID (e.g., 'pipe-abc123')",
				},
			},
			Required: []string{"pipeline_id"},
		},
	}
}

// HandleGetPipelineById handles the MCP tool invocation for getting pipeline by ID.
func HandleGetPipelineById(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_pipeline_by_id")

	// Extract pipeline_id from arguments
	pipelineID, ok := arguments["pipeline_id"].(string)
	if !ok || pipelineID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "pipeline_id is required",
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

	// Query pipeline
	pipeline, err := client.GetById(ctx, pipelineID)
	if err != nil {
		return errors.HandleGRPCError(err, pipelineID), nil
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

	log.Printf("Tool completed: get_pipeline_by_id, pipeline: %s", pipelineID)

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
