<!-- 8a211760-b637-4ef3-9118-df09caf17af7 22042cad-acec-4fd3-9d16-12ca873a3aaf -->
# Reorganize MCP Server Documentation

## Goals

1. **README.md**: Focus on how users connect and use the MCP server (following GitHub MCP server structure)
2. **Cursor Integration**: Clear `mcp.json` examples for remote and local HTTP endpoints
3. **Simplify Auth**: Use `PLANTON_API_KEY` as bearer token (eliminate separate `PLANTON_MCP_HTTP_BEARER_TOKEN`)
4. **Move Details**: Server startup/running details go to `docs/http-transport.md` and `docs/development.md`

## Key Changes

### 1. README.md Restructure

**Keep concise and user-focused:**

- Quick integration examples (Cursor, Claude Desktop, LangGraph)
- HTTP endpoint configuration for `mcp.planton.ai` (remote)
- HTTP endpoint configuration for `localhost:8080` (local testing)
- Brief installation section with links to detailed docs
- Remove verbose HTTP transport details (move to docs/)

**New Cursor Integration Section:**

```json
// Remote deployment (when available)
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "https://mcp.planton.ai/",
      "headers": {
        "Authorization": "Bearer YOUR_PLANTON_API_KEY"
      }
    }
  }
}

// Local Docker testing
{
  "mcpServers": {
    "planton-cloud": {
      "type": "http",
      "url": "http://localhost:8080/",
      "headers": {
        "Authorization": "Bearer YOUR_PLANTON_API_KEY"
      }
    }
  }
}

// Local binary (stdio mode)
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "YOUR_PLANTON_API_KEY",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```

### 2. Simplify Authentication

**Current (complex):**

- `PLANTON_API_KEY` for Planton Cloud APIs
- `PLANTON_MCP_HTTP_BEARER_TOKEN` for HTTP transport

**New (simplified):**

- `PLANTON_API_KEY` serves both purposes
- When HTTP auth enabled, bearer token = PLANTON_API_KEY
- Update code in `internal/mcp/http_server.go`

### 3. Update docs/http-transport.md

**Add "Running Locally" section:**

- Docker command to run on localhost:8080
- Binary command to run on localhost:8080
- How to test with curl
- How to configure Cursor mcp.json for local endpoint

**Add "Remote Deployment" section:**

- Expected endpoint: `https://mcp.planton.ai/`
- How users configure Cursor for remote access
- Production considerations

### 4. Files to Modify

1. **README.md** - Restructure to user-focused integration guide
2. **docs/http-transport.md** - Add local running + Cursor integration details
3. **internal/mcp/http_server.go** - Use PLANTON_API_KEY as bearer token
4. **internal/config/config.go** - Remove PLANTON_MCP_HTTP_BEARER_TOKEN
5. **docs/configuration.md** - Update environment variables table

## Structure Inspiration

Follow GitHub MCP server pattern:

- **Overview** - What it does
- **Quick Start** - Integration examples first
- **Installation** - Brief, link to details
- **Configuration** - Essential vars only
- **Tools** - What's available
- **Links** - Detailed docs

## Implementation Order

1. Review GitHub MCP server README structure (if accessible)
2. Update authentication code to use PLANTON_API_KEY
3. Restructure README.md with Cursor examples
4. Update docs/http-transport.md with local setup
5. Update configuration documentation

### To-dos

- [ ] Modify HTTP server to use PLANTON_API_KEY as bearer token instead of separate token
- [ ] Restructure README.md to focus on user integration with clear Cursor mcp.json examples
- [ ] Update docs/http-transport.md with local setup instructions and Cursor integration
- [ ] Update configuration documentation to remove separate bearer token variable