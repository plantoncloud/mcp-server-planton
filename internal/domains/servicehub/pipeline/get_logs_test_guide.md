# Pipeline Build Logs - Testing Guide

## Overview

This guide provides instructions for testing the timeout protection and pagination features of the `get_pipeline_build_logs` MCP tool.

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
- Completes after exactly 2 minutes
- Returns partial results
- `limit_reached: true`
- Clear timeout message

**Expected Response**:
```json
{
  "log_entries": [...],
  "total_returned": 3500,
  "limit_reached": true,
  "message": "Log streaming timed out after 2 minutes. Showing 3500 log entries (skipped 0). The pipeline may have produced more logs. Check the pipeline status to see if it's still running."
}
```

### 6. Agent Integration Test

**Objective**: Verify the fix prevents frozen conversations

**Test Steps**:
1. Start a conversation with an agent
2. Ask agent to get pipeline logs for a large pipeline
3. Wait for tool to complete
4. Verify agent continues conversation

**Expected Behavior**:
- Tool completes within 2 minutes
- Agent receives structured response
- Agent provides summary to user
- Conversation remains responsive
- User can send follow-up messages

**Not Expected** (these were the bugs):
- ❌ Tool times out after 4+ minutes
- ❌ `UnboundLocalError` crash
- ❌ Frozen conversation UI
- ❌ No response from agent

## Performance Metrics to Monitor

### Timing Metrics

| Scenario | Expected Duration | Actual Duration | Status |
|----------|------------------|-----------------|--------|
| Small (< 100) | < 1 second | ___ | ___ |
| Medium (1000) | < 10 seconds | ___ | ___ |
| Large (5000) | < 30 seconds | ___ | ___ |
| Very Large (hit timeout) | ~120 seconds | ___ | ___ |

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

- [ ] No timeouts exceed 2 minutes
- [ ] No `UnboundLocalError` exceptions
- [ ] All pagination calculations are correct
- [ ] Response format is consistent across scenarios
- [ ] Agent conversations remain responsive
- [ ] Error messages are clear and actionable
- [ ] Logs contain useful debugging information
- [ ] Performance meets expectations
- [ ] UI displays results properly
- [ ] Users can continue conversations after tool execution

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

1. **No Frozen Conversations**: All agent executions complete successfully
2. **Fast Response**: 95% of requests complete in < 30 seconds
3. **No Timeouts**: No requests exceed 2 minutes
4. **Proper Pagination**: Users can fetch all logs through pagination
5. **Clear Messages**: Users understand when limits are hit
6. **No Crashes**: Zero `UnboundLocalError` occurrences
7. **Positive UX**: Users report improved debugging experience

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

