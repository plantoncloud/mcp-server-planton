package errors

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse represents an error response for MCP tool calls.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	OrgID   string `json:"org_id,omitempty"`
}

// HandleGRPCError converts gRPC errors to user-friendly error responses.
// This is exported so it can be reused by all domains.
func HandleGRPCError(err error, orgID string) *mcp.CallToolResult {
	st, ok := status.FromError(err)
	if !ok {
		errResp := ErrorResponse{
			Error:   "UNKNOWN_ERROR",
			Message: fmt.Sprintf("An unexpected error occurred: %v", err),
			OrgID:   orgID,
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON))
	}

	var errResp ErrorResponse
	errResp.OrgID = orgID

	switch st.Code() {
	case codes.Unauthenticated:
		errResp.Error = "UNAUTHENTICATED"
		errResp.Message = "Authentication failed. Your session may have expired. Please refresh and try again."

	case codes.PermissionDenied:
		errResp.Error = "PERMISSION_DENIED"
		errResp.Message = fmt.Sprintf(
			"You don't have permission to access this resource for organization '%s'. "+
				"Please contact your organization administrator.",
			orgID,
		)

	case codes.Unavailable:
		errResp.Error = "UNAVAILABLE"
		errResp.Message = "Planton Cloud APIs are currently unavailable. Please try again in a moment."

	case codes.NotFound:
		errResp.Error = "NOT_FOUND"
		errResp.Message = fmt.Sprintf("Resource not found for organization '%s'.", orgID)

	default:
		errResp.Error = st.Code().String()
		errResp.Message = st.Message()
		if errResp.Message == "" {
			errResp.Message = "An unexpected error occurred."
		}
	}

	log.Printf(
		"Tool error: org_id=%s, code=%s, message=%s",
		orgID, errResp.Error, errResp.Message,
	)

	errJSON, _ := json.MarshalIndent(errResp, "", "  ")
	return mcp.NewToolResultText(string(errJSON))
}












