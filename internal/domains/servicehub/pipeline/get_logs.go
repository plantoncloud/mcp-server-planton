package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

// TektonTaskLogEntry is a simplified representation of a log entry for JSON serialization.
type TektonTaskLogEntry struct {
	Owner      string `json:"owner,omitempty"`
	TaskName   string `json:"task_name"`
	LogMessage string `json:"log_message"`
}

// CreateGetPipelineBuildLogsTool creates the MCP tool definition for streaming pipeline build logs.
func CreateGetPipelineBuildLogsTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_pipeline_build_logs",
		Description: "Stream and retrieve all build logs for a pipeline execution. " +
			"Returns complete Tekton task logs including build output, errors, and diagnostic messages. " +
			"Logs are fetched from Redis (for running pipelines) or R2 storage (for completed pipelines). " +
			"Use this to troubleshoot build failures and understand what happened during pipeline execution.",
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

// HandleGetPipelineBuildLogs handles the MCP tool invocation for streaming pipeline build logs.
func HandleGetPipelineBuildLogs(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_pipeline_build_logs")

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

	// Start log stream
	stream, err := client.GetLogStream(ctx, pipelineID)
	if err != nil {
		return errors.HandleGRPCError(err, pipelineID), nil
	}

	// Collect all log entries from the stream
	var logEntries []TektonTaskLogEntry
	for {
		logEntry, err := stream.Recv()
		if err == io.EOF {
			// Stream completed successfully
			break
		}
		if err != nil {
			// Stream error
			errResp := errors.ErrorResponse{
				Error:   "STREAM_ERROR",
				Message: fmt.Sprintf("Error receiving log entry: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}

		// Convert to simple struct
		logEntries = append(logEntries, TektonTaskLogEntry{
			Owner:      logEntry.GetOwner(),
			TaskName:   logEntry.GetTaskName(),
			LogMessage: logEntry.GetLogMessage(),
		})
	}

	log.Printf("Tool completed: get_pipeline_build_logs, pipeline: %s, entries: %d", pipelineID, len(logEntries))

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(logEntries, "", "  ")
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
