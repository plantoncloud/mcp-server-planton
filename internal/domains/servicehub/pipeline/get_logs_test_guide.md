# Pipeline Build Logs - Testing Guide

## Overview

This guide provides instructions for testing the improved `get_pipeline_build_logs` MCP tool with:
- Reduced timeout (45 seconds instead of 2 minutes)
- Smart early return to prevent timeouts
- Enhanced crash protection with context checks and recovery
- Better error messages and pagination guidance
- gRPC keepalive for connection stability

## Test Scenarios

### 1. Small Pipeline (< 100 log entries)

**Objective**: Verify normal operation with small log files

**Test Steps**:
```json
{
  "pipeline_id": "pipe-small-logs-example"
}
```

**Expected Behavior**:
- Completes quickly (< 1 second)
- Returns all log entries
- `limit_reached: false`
- `has_more: false`
- No warning messages

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 45,
  "limit_reached": false,
  "has_more": false
}
```

### 2. Medium Pipeline (100-5000 log entries)

**Objective**: Verify normal operation within limits

**Test Steps**:
```json
{
  "pipeline_id": "pipe-medium-logs-example"
}
```

**Expected Behavior**:
- Completes within timeout (< 2 minutes)
- Returns all log entries
- `limit_reached: false`
- `has_more: false`
- No warning messages

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 2500,
  "limit_reached": false,
  "has_more": false
}
```

### 3. Large Pipeline (> 5000 log entries)

**Objective**: Verify entry limit handling and pagination

**Test Steps**:

**Request 1** - Get first page:
```json
{
  "pipeline_id": "pipe-large-logs-example"
}
```

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 5000,
  "limit_reached": true,
  "has_more": true,
  "next_offset": 5000,
  "message": "Log entry limit reached. Showing 5000 log entries (skipped 0). More logs are available. Use skip_entries=5000 to fetch the next page."
}
```

**Request 2** - Get second page:
```json
{
  "pipeline_id": "pipe-large-logs-example",
  "skip_entries": 5000
}
```

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 5000,
  "total_skipped": 5000,
  "limit_reached": true,
  "has_more": true,
  "next_offset": 10000,
  "message": "Log entry limit reached. Showing 5000 log entries (skipped 5000). More logs are available. Use skip_entries=10000 to fetch the next page."
}
```

**Request 3** - Get last page (assuming 12000 total entries):
```json
{
  "pipeline_id": "pipe-large-logs-example",
  "skip_entries": 10000
}
```

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 2000,
  "total_skipped": 10000,
  "limit_reached": true,
  "has_more": false,
  "message": "Showing 2000 log entries (skipped 10000). This is the last page of logs."
}
```

### 4. Custom Entry Limits

**Objective**: Verify custom max_entries parameter

**Test Steps**:
```json
{
  "pipeline_id": "pipe-large-logs-example",
  "max_entries": 1000
}
```

**Expected Behavior**:
- Returns exactly 1000 entries
- Provides pagination info for remaining entries

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 1000,
  "limit_reached": true,
  "has_more": true,
  "next_offset": 1000,
  "message": "Log entry limit reached. Showing 1000 log entries (skipped 0). More logs are available. Use skip_entries=1000 to fetch the next page."
}
```

### 5. Timeout Scenario (Very Large Pipeline)

**Objective**: Verify timeout protection with extremely large log files

**Test Steps**:
```json
{
  "pipeline_id": "pipe-very-large-logs-example"
}
```

**Expected Behavior**:
- Completes after 45 seconds or less (with early return)
- Returns partial results (at least 1000 entries if streaming is working)
- `limit_reached: true`
- Clear timeout or early return message

**Expected Response** (Early Return):
```json
{
  "log_entries": [...],
  "total_returned": 1500,
  "limit_reached": true,
  "has_more": true,
  "next_offset": 1500,
  "message": "Retrieved 1500 log entries in 23s (early return to prevent timeout). More logs are available. Use skip_entries=1500 to fetch the next page."
}
```

**Expected Response** (Timeout):
```json
{
  "log_entries": [...],
  "total_returned": 3500,
  "limit_reached": true,
  "has_more": false,
  "message": "Log streaming timed out after 45 seconds. Showing 3500 log entries (skipped 0). The pipeline may have produced more logs. Use skip_entries=3500 to fetch the next page."
}
```

### 6. Context Cancellation / Connection Loss Test

**Objective**: Verify server doesn't crash when client disconnects

**Test Steps**:
1. Start log streaming request for a large pipeline
2. After 5-10 seconds, close browser tab / kill connection
3. Check server logs
4. Verify server is still running
5. Make another request to confirm server is healthy

**Expected Behavior**:
- Server detects context cancellation
- Logs show: "Context cancelled at checkpoint" or "Context cancelled before building response"
- Server continues running (no panic)
- Partial results are logged (if any were collected)
- Next request succeeds normally

**Expected Server Logs**:
```
Checkpoint: pipeline=pipe-xxx, entries=500, processed=600, elapsed=8s
Context cancelled at checkpoint: pipeline=pipe-xxx, entries=500, processed=700
Context cancelled before building response: context canceled, entries collected: 500
Tool completed: get_pipeline_build_logs, pipeline: pipe-xxx, entries: 500, duration: 9s, rate: 55.56 entries/sec, limited: true, cancelled: true
```

**Not Expected**:
- ❌ Server panic/crash
- ❌ Nil pointer dereference
- ❌ SSE server error
- ❌ Server becomes unresponsive

### 7. Agent Integration Test

**Objective**: Verify the fix prevents frozen conversations and server crashes

**Test Steps**:
1. Start a conversation with an agent
2. Ask agent to get pipeline logs for a large pipeline
3. Wait for tool to complete
4. Verify agent continues conversation
5. Test with connection interruption (close browser/kill connection mid-request)

**Expected Behavior**:
- Tool completes within 45 seconds (15-30s for typical pipelines)
- Agent receives structured response with pagination guidance
- Agent provides summary to user
- Conversation remains responsive
- User can send follow-up messages
- **Connection interruption**: Server doesn't crash, logs show graceful handling

**Not Expected** (these were the bugs):
- ❌ Tool times out after 2+ minutes
- ❌ SSE server panic with nil pointer dereference
- ❌ Frozen conversation UI
- ❌ No response from agent
- ❌ Server crash on connection loss

## Performance Metrics to Monitor

### Timing Metrics

| Scenario | Expected Duration | Actual Duration | Status |
|----------|------------------|-----------------|--------|
| Small (< 100) | < 5 seconds | ___ | ___ |
| Medium (1000) | < 15 seconds | ___ | ___ |
| Large (5000) | < 30 seconds | ___ | ___ |
| Very Large (early return) | ~22-30 seconds | ___ | ___ |
| Very Large (hit timeout) | ~45 seconds | ___ | ___ |

### Entry Count Metrics

| Scenario | Expected Entries | Actual Entries | Status |
|----------|-----------------|----------------|--------|
| Small pipeline | All entries | ___ | ___ |
| Medium pipeline | All entries | ___ | ___ |
| Large pipeline (page 1) | 5000 | ___ | ___ |
| Large pipeline (page 2) | 5000 | ___ | ___ |
| Large pipeline (last page) | Remaining | ___ | ___ |

## Error Scenarios to Test

### 1. Invalid Pipeline ID

**Test**:
```json
{
  "pipeline_id": "pipe-does-not-exist"
}
```

**Expected**: Proper gRPC error response, no crash

### 2. Invalid Parameters

**Test**:
```json
{
  "pipeline_id": "pipe-test",
  "max_entries": -100,
  "skip_entries": -50
}
```

**Expected**: Parameters normalized to safe values, no crash

### 3. Network Issues

**Test**: Simulate network interruption during streaming

**Expected**: Timeout handling kicks in, returns partial results

## Verification Checklist

After running tests, verify:

- [ ] No timeouts exceed 45 seconds
- [ ] No SSE server panics or nil pointer dereferences
- [ ] Context cancellation is handled gracefully
- [ ] All pagination calculations are correct
- [ ] Response format is consistent across scenarios
- [ ] Agent conversations remain responsive
- [ ] Error messages are clear and actionable with retry guidance
- [ ] Logs contain useful debugging information (checkpoints, rates, etc.)
- [ ] Performance meets expectations (< 30s for typical pipelines)
- [ ] Early return triggers appropriately (>= 1000 entries at 22.5s+)
- [ ] gRPC keepalive prevents connection drops
- [ ] UI displays results properly
- [ ] Users can continue conversations after tool execution
- [ ] Server remains healthy after connection interruptions

## Testing in Production

### Gradual Rollout

1. Deploy to development environment first
2. Test with known problematic pipelines
3. Monitor metrics for 24 hours
4. Deploy to staging environment
5. Verify with actual user workflows
6. Deploy to production with monitoring

### Monitoring

Watch for:
- Timeout occurrence rate
- Average entries returned
- Pagination usage patterns
- Tool execution duration
- Error rates
- User feedback

### Rollback Criteria

Roll back if:
- Timeout rate increases
- Error rate > 5%
- User complaints about missing logs
- Performance degradation
- New crashes or errors

## Success Criteria

The fix is successful if:

1. **No Server Crashes**: Zero SSE server panics on connection loss
2. **No Frozen Conversations**: All agent executions complete successfully
3. **Fast Response**: 95% of requests complete in < 30 seconds
4. **Controlled Timeouts**: All requests complete within 45 seconds
5. **Proper Pagination**: Users can fetch all logs through pagination
6. **Clear Messages**: Users understand when limits are hit and how to continue
7. **Graceful Degradation**: Partial results returned on any error
8. **Connection Stability**: gRPC keepalive prevents drops during streaming
9. **Positive UX**: Users report improved debugging experience

## Debugging Failed Tests

If tests fail, check:

1. **Server Logs**: Look for timeout messages, entry counts, errors
2. **Client Logs**: Check for exception traces, timing information
3. **Network**: Verify no connectivity issues
4. **Pipeline State**: Confirm pipeline actually has logs
5. **Configuration**: Verify timeout and limit constants
6. **gRPC Connection**: Check for connection pool exhaustion

## Known Limitations

Document any limitations discovered during testing:

- Maximum practical pipeline size that can be fully retrieved
- Performance with concurrent requests
- Behavior with very slow log sources
- Edge cases in pagination calculations

## Future Improvements

Based on testing, consider:

- [ ] Dynamic timeout adjustment based on log volume
- [ ] Compressed log streaming for better performance
- [ ] Caching frequently accessed logs
- [ ] Background pre-fetching for large pipelines
- [ ] Progress indicators for long-running streams
- [ ] Configurable limits per organization

## Implementation Summary

### Changes Made

**Phase 1: Crash Fix (Critical)**
- ✅ Added panic recovery with defer/recover
- ✅ Context cancellation detection at multiple points
- ✅ Checkpoints every 100 entries to detect connection loss
- ✅ Pre-marshal context validation to prevent writing to dead connections

**Phase 2: Performance Improvements (Critical)**
- ✅ Reduced timeout from 120s to 45s (62.5% reduction)
- ✅ Pre-allocated slice capacity for better memory efficiency
- ✅ Smart early return: >= 1000 entries after 22.5s
- ✅ Checkpoint logging every 100 entries

**Phase 3: Enhanced Error Handling**
- ✅ Distinguish between timeout, cancellation, and stream errors
- ✅ Partial results with retry guidance on errors
- ✅ Clear pagination messages with specific skip_entries values
- ✅ Context-aware error messages

**Phase 4: Client-Side Optimizations**
- ✅ gRPC keepalive (30s ping, 10s timeout)
- ✅ PermitWithoutStream to maintain connection health

**Phase 5: Monitoring and Logging**
- ✅ Performance metrics (entries/sec rate)
- ✅ Checkpoint logging with elapsed time
- ✅ Error categorization in logs
- ✅ Connection state tracking

### Key Files Modified

1. `get_logs.go` (main implementation)
   - Added constants: CheckpointInterval, EarlyReturnThreshold, EarlyReturnTimeRatio
   - Enhanced HandleGetPipelineBuildLogs with defensive programming
   - Improved error messages and pagination guidance

2. `pipeline_client.go` (gRPC client)
   - Added keepalive.ClientParameters
   - Improved connection stability for long-running operations

3. `get_logs_test_guide.md` (testing documentation)
   - Updated test scenarios for 45s timeout
   - Added context cancellation test
   - Enhanced success criteria

### Expected Performance

| Pipeline Size | Before | After | Improvement |
|--------------|--------|-------|-------------|
| Small (<100) | ~2 min | <5s | 96% faster |
| Medium (1000) | ~2 min | <15s | 87.5% faster |
| Large (5000) | ~2 min | <30s | 75% faster |
| Very Large | ~2 min | 22-45s | 62-81% faster |

### Reliability Improvements

- **Zero crashes**: Context checks prevent nil pointer panics
- **Graceful degradation**: Partial results always returned
- **Clear guidance**: Users know exactly how to continue
- **Connection stability**: Keepalive prevents drops

### Testing Priority

1. **Critical**: Context cancellation / connection loss (prevent crashes)
2. **High**: Performance timing (verify < 45s)
3. **High**: Early return logic (verify triggers at 1000+ entries)
4. **Medium**: Error message clarity
5. **Medium**: Pagination accuracy


