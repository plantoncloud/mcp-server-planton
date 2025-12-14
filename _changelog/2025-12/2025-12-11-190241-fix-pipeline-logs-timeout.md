# Fix Pipeline Build Logs Timeout Issue

**Type:** Bug Fix  
**Component:** Service Hub MCP Tools  
**Impact:** Critical - Fixes agent execution freezing on large pipeline logs  
**Date:** 2025-12-11

## Problem

The `get_pipeline_build_logs` MCP tool was failing after ~4.5 minutes when streaming large Tekton pipeline logs, causing:

1. **Tool Timeout**: Log streaming exceeded typical agent/tool timeout limits
2. **Library Bug**: `langchain-mcp-adapters` v0.1.14 has an `UnboundLocalError` when operations are cancelled
3. **Frozen Conversations**: Frontend UI became unresponsive, unable to continue after tool failure
4. **Poor UX**: No feedback to users about what went wrong or how to proceed

### Error Flow

```
13:11:48 - get_pipeline_build_logs invoked
13:16:24 - Operation cancelled after 4.5 minutes
13:16:24 - UnboundLocalError: call_tool_result not defined
→ Agent execution fails, conversation freezes
```

## Root Cause

The MCP server was attempting to stream **all** pipeline logs without limits:
- No timeout protection
- No entry count limits
- No pagination support
- Large pipelines (10k+ log entries) exceeded timeouts

When the operation was cancelled, `langchain-mcp-adapters` tried to return an uninitialized variable, causing a crash.

## Solution

Added **timeout protection** and **entry limits** to prevent cancellation and provide graceful degradation.

### Changes Made

**File:** `internal/domains/servicehub/pipeline/get_logs.go`

#### 1. Added Safety Constants

```go
const (
    MaxLogStreamDuration = 2 * time.Minute  // Safely under agent timeout
    MaxLogEntries = 5000                     // Reasonable UI display limit
)
```

#### 2. Timeout-Protected Streaming

- Wrapped stream processing with `context.WithTimeout`
- Added entry counter with limit checking
- Early exit when limits are reached

#### 3. Structured Response Format

```go
type LogStreamResponse struct {
    LogEntries    []TektonTaskLogEntry `json:"log_entries"`
    TotalReturned int                  `json:"total_returned"`
    LimitReached  bool                 `json:"limit_reached,omitempty"`
    Message       string               `json:"message,omitempty"`
}
```

#### 4. Informative User Messages

When limits are hit, users receive clear messages:

- **Timeout**: "Log streaming timed out after 2 minutes. Showing first N log entries. Check pipeline status to see if it's still running."
- **Entry Limit**: "Log entry limit reached. Showing first 5000 log entries. The pipeline produced more logs than can be displayed."

#### 5. Enhanced Logging

```go
log.Printf("Tool completed: pipeline: %s, entries: %d, duration: %v, limited: %v", 
    pipelineID, len(logEntries), duration, limitReached)
```

## Benefits

1. **No More Timeouts**: Operations complete within 2 minutes or return partial results
2. **Graceful Degradation**: Users see first 5000 entries with clear message about limits
3. **Better Observability**: Enhanced logging helps diagnose issues
4. **Improved UX**: Clear feedback when limits are reached
5. **Prevents Crashes**: No more `UnboundLocalError` from library bug

## Testing

### Test Scenarios

1. **Small Logs (<100 entries)**: Completes normally, returns all entries
2. **Medium Logs (100-5000 entries)**: Completes within timeout, returns all entries
3. **Large Logs (>5000 entries)**: Returns first 5000 with `limit_reached: true` and helpful message
4. **Very Large Logs (>10k entries)**: May hit timeout, returns partial results with timeout message

### Expected Behavior

- Tool always completes within 2 minutes
- No crashes or frozen conversations
- Clear user feedback about what happened
- Agents can continue conversation after tool execution

## Known Limitations

1. **No Pagination**: Currently returns only first 5000 entries (pagination planned for future)
2. **Library Bug Remains**: `langchain-mcp-adapters` v0.1.14 still has the underlying bug, but we prevent it by avoiding cancellation
3. **Fixed Limits**: Timeout and entry limits are hardcoded (configuration support could be added)

## Related Issues

- `langchain-mcp-adapters` GitHub issue: "Issue: `UnboundLocalError` in `call_tool` on Client Disconnect"
- Current version: 0.1.14

## Next Steps

- [ ] Add pagination support for large log streams
- [ ] Monitor timeout rates in production
- [ ] Consider upgrading `langchain-mcp-adapters` if bug is fixed
- [ ] Add configuration for timeout/limit values if needed

## Migration Notes

**No breaking changes.** This is a backward-compatible bug fix.

Existing tool calls will:
- Complete faster
- Return structured responses with metadata
- Provide better error messages
- Never freeze conversations

## Verification

To verify the fix:

```bash
# Test with a service that has large pipeline logs
# The tool should complete within 2 minutes and return results
```

Expected response format:

```json
{
  "log_entries": [...],
  "total_returned": 5000,
  "limit_reached": true,
  "message": "Log entry limit reached. Showing first 5000 log entries..."
}
```

## Impact Assessment

**High Priority Fix:**
- Prevents agent execution failures
- Unblocks frozen conversations
- Significantly improves user experience with CI/CD debugging
- No downtime or migration required









