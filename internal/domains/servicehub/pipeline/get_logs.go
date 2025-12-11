package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

const (
	// MaxLogStreamDuration is the maximum time allowed for streaming logs
	// Set to 2 minutes to be safely under typical agent/tool timeouts
	MaxLogStreamDuration = 2 * time.Minute

	// MaxLogEntries is the maximum number of log entries to return
	// Prevents overwhelming the client and hitting timeout limits
	MaxLogEntries = 5000
)

// TektonTaskLogEntry is a simplified representation of a log entry for JSON serialization.
type TektonTaskLogEntry struct {
	Owner      string `json:"owner,omitempty"`
	TaskName   string `json:"task_name"`
	LogMessage string `json:"log_message"`
}

// LogStreamResponse wraps log entries with metadata about streaming limits and pagination
type LogStreamResponse struct {
	LogEntries    []TektonTaskLogEntry `json:"log_entries"`
	TotalReturned int                  `json:"total_returned"`
	TotalSkipped  int                  `json:"total_skipped,omitempty"`
	LimitReached  bool                 `json:"limit_reached,omitempty"`
	HasMore       bool                 `json:"has_more,omitempty"`
	NextOffset    int                  `json:"next_offset,omitempty"`
	Message       string               `json:"message,omitempty"`
}

// CreateGetPipelineBuildLogsTool creates the MCP tool definition for streaming pipeline build logs.
func CreateGetPipelineBuildLogsTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_pipeline_build_logs",
		Description: "Stream and retrieve build logs for a pipeline execution. " +
			"Returns Tekton task logs including build output, errors, and diagnostic messages. " +
			"Logs are fetched from Redis (for running pipelines) or R2 storage (for completed pipelines). " +
			"Use this to troubleshoot build failures and understand what happened during pipeline execution. " +
			fmt.Sprintf("Note: Returns up to %d log entries per request with a %d minute timeout. ", MaxLogEntries, int(MaxLogStreamDuration.Minutes())) +
			"For large log files, use 'max_entries' and 'skip_entries' parameters for pagination. " +
			"If limits are reached, partial results are returned with a message indicating more logs are available.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"pipeline_id": map[string]interface{}{
					"type":        "string",
					"description": "Pipeline ID (e.g., 'pipe-abc123')",
				},
				"max_entries": map[string]interface{}{
					"type":        "number",
					"description": fmt.Sprintf("Maximum number of log entries to return (default: %d, max: %d)", MaxLogEntries, MaxLogEntries),
				},
				"skip_entries": map[string]interface{}{
					"type":        "number",
					"description": "Number of log entries to skip for pagination (default: 0). Use this to fetch subsequent pages of logs.",
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

	// Extract optional max_entries parameter
	maxEntries := MaxLogEntries
	if maxEntriesArg, ok := arguments["max_entries"].(float64); ok {
		requestedMax := int(maxEntriesArg)
		if requestedMax > 0 && requestedMax <= MaxLogEntries {
			maxEntries = requestedMax
		} else if requestedMax > MaxLogEntries {
			log.Printf("Requested max_entries %d exceeds limit %d, using limit", requestedMax, MaxLogEntries)
			maxEntries = MaxLogEntries
		}
	}

	// Extract optional skip_entries parameter for pagination
	skipEntries := 0
	if skipEntriesArg, ok := arguments["skip_entries"].(float64); ok {
		skipEntries = int(skipEntriesArg)
		if skipEntries < 0 {
			skipEntries = 0
		}
	}

	log.Printf("Pipeline logs request: pipeline=%s, max_entries=%d, skip_entries=%d",
		pipelineID, maxEntries, skipEntries)

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

	// Create timeout context for streaming
	streamCtx, cancel := context.WithTimeout(ctx, MaxLogStreamDuration)
	defer cancel()

	// Start log stream with timeout context
	stream, err := client.GetLogStream(streamCtx, pipelineID)
	if err != nil {
		return errors.HandleGRPCError(err, pipelineID), nil
	}

	// Track streaming state
	var logEntries []TektonTaskLogEntry
	limitReached := false
	timeoutReached := false
	hasMore := false
	startTime := time.Now()

	// Track total entries processed (including skipped)
	totalProcessed := 0
	entriesSkipped := 0

	// Collect log entries with limits and pagination
	for len(logEntries) < maxEntries {
		logEntry, err := stream.Recv()
		if err == io.EOF {
			// Stream completed successfully
			break
		}
		if err != nil {
			// Check if this is a timeout error
			if streamCtx.Err() == context.DeadlineExceeded {
				timeoutReached = true
				log.Printf("Log stream timeout reached for pipeline: %s, duration: %v, entries collected: %d, skipped: %d",
					pipelineID, time.Since(startTime), len(logEntries), entriesSkipped)
				break
			}
			// Other stream error
			errResp := errors.ErrorResponse{
				Error:   "STREAM_ERROR",
				Message: fmt.Sprintf("Error receiving log entry: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}

		totalProcessed++

		// Skip entries for pagination
		if entriesSkipped < skipEntries {
			entriesSkipped++
			continue
		}

		// Convert to simple struct
		logEntries = append(logEntries, TektonTaskLogEntry{
			Owner:      logEntry.GetOwner(),
			TaskName:   logEntry.GetTaskName(),
			LogMessage: logEntry.GetLogMessage(),
		})
	}

	// Check if we hit the entry limit (meaning there might be more)
	if len(logEntries) >= maxEntries {
		limitReached = true
		// Try to peek ahead to see if there are more entries
		_, err := stream.Recv()
		if err != io.EOF && streamCtx.Err() != context.DeadlineExceeded {
			hasMore = true
		}
		log.Printf("Log entry limit reached for pipeline: %s, limit: %d, has_more: %v",
			pipelineID, maxEntries, hasMore)
	}

	// Build response with metadata
	response := LogStreamResponse{
		LogEntries:    logEntries,
		TotalReturned: len(logEntries),
		TotalSkipped:  entriesSkipped,
		LimitReached:  limitReached || timeoutReached,
		HasMore:       hasMore,
	}

	// Add next offset if there are more entries
	if hasMore {
		response.NextOffset = skipEntries + len(logEntries)
	}

	// Add informative message if limits were hit
	if timeoutReached {
		response.Message = fmt.Sprintf(
			"Log streaming timed out after %d minutes. Showing %d log entries (skipped %d). "+
				"The pipeline may have produced more logs. Check the pipeline status to see if it's still running.",
			int(MaxLogStreamDuration.Minutes()), len(logEntries), entriesSkipped)
	} else if limitReached && hasMore {
		response.Message = fmt.Sprintf(
			"Log entry limit reached. Showing %d log entries (skipped %d). "+
				"More logs are available. Use skip_entries=%d to fetch the next page.",
			len(logEntries), entriesSkipped, response.NextOffset)
	} else if limitReached && !hasMore {
		response.Message = fmt.Sprintf(
			"Showing %d log entries (skipped %d). This is the last page of logs.",
			len(logEntries), entriesSkipped)
	}

	duration := time.Since(startTime)
	log.Printf("Tool completed: get_pipeline_build_logs, pipeline: %s, entries: %d, duration: %v, limited: %v",
		pipelineID, len(logEntries), duration, limitReached || timeoutReached)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(response, "", "  ")
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

