# Planton Cloud MCP Server

[![CI](https://github.com/plantoncloud-inc/mcp-server-planton/actions/workflows/ci.yml/badge.svg)](https://github.com/plantoncloud-inc/mcp-server-planton/actions/workflows/ci.yml)
[![CodeQL](https://github.com/plantoncloud-inc/mcp-server-planton/actions/workflows/codeql.yml/badge.svg)](https://github.com/plantoncloud-inc/mcp-server-planton/actions/workflows/codeql.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/plantoncloud-inc/mcp-server-planton)](https://goreportcard.com/report/github.com/plantoncloud-inc/mcp-server-planton)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Docker](https://img.shields.io/badge/Docker-ghcr.io-blue?logo=docker)](https://github.com/plantoncloud-inc/mcp-server-planton/pkgs/container/mcp-server-planton)

MCP (Model Context Protocol) server for Planton Cloud that enables AI agents to query and manage cloud resources using user permissions.

## Overview

The Planton Cloud MCP Server provides AI assistants like Cursor, Claude Desktop, and LangGraph agents with tools to interact with Planton Cloud resources. All queries respect your actual permissions through API key authentication.

**Key Features:**
- User-scoped permissions via API key authentication
- Query cloud resources, environments, organizations
- Create and manage cloud infrastructure
- Works with any MCP client (Cursor, Claude Desktop, LangGraph)
- Available as HTTP endpoint or local binary

## Quick Start

### Get Your API Key

1. Log in to [Planton Cloud Console](https://console.planton.cloud)
2. Click your profile icon → **API Keys**
3. Click **Create Key** and copy the generated key

### Integration with Cursor

Add to your Cursor MCP settings (`~/.cursor/mcp.json`):

#### Remote Endpoint (Recommended)

```json
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
```

#### Local Testing with Docker

```json
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
```

Run the Docker container:
```bash
docker run -p 8080:8080 \
  -e PLANTON_API_KEY="YOUR_PLANTON_API_KEY" \
  -e PLANTON_MCP_TRANSPORT="http" \
  -e PLANTON_MCP_HTTP_AUTH_ENABLED="true" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

#### Local Binary (STDIO Mode)

```json
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

Install the binary:
```bash
# macOS (ARM64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Darwin_arm64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Darwin_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/

# Linux (AMD64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/latest/download/mcp-server-planton_Linux_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
```

### Integration with Claude Desktop

Add to your Claude Desktop MCP settings:

```json
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

### Integration with LangGraph

Add to your `langgraph.json`:

```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "${PLANTON_API_KEY}",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```

## Available Tools

The MCP server provides tools for querying and managing Planton Cloud resources:

### Cloud Resources
- `list_cloud_resource_kinds` - List all available cloud resource types
- `get_cloud_resource_schema` - Get schema/spec for a resource type
- `search_cloud_resources` - Search and filter cloud resources
- `lookup_cloud_resource_by_name` - Find resource by exact name
- `get_cloud_resource_by_id` - Get complete resource details by ID
- `create_cloud_resource` - Create new cloud resources
- `update_cloud_resource` - Update existing resources
- `delete_cloud_resource` - Delete cloud resources

### Environments
- `list_environments_for_org` - List environments in an organization

All tools respect your user permissions - you can only access resources you have permission to view or manage.

## Configuration

### Essential Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PLANTON_API_KEY` | **(required)** | Your API key from Planton Cloud console |
| `PLANTON_CLOUD_ENVIRONMENT` | `live` | Target environment: `live`, `test`, or `local` |
| `PLANTON_MCP_TRANSPORT` | `stdio` | Transport mode: `stdio`, `http`, or `both` |
| `PLANTON_MCP_HTTP_PORT` | `8080` | HTTP server port (when using HTTP transport) |
| `PLANTON_MCP_HTTP_AUTH_ENABLED` | `true` | Enable bearer token authentication for HTTP |

**Note:** When HTTP authentication is enabled, your `PLANTON_API_KEY` is used as the bearer token.

For complete configuration options, see [Configuration Guide](docs/configuration.md).

## Security

This MCP server uses **user API keys** rather than machine accounts, ensuring that:

- All queries respect your actual permissions (Fine-Grained Authorization)
- No API key persistence - keys are held in memory only during execution
- Every API call is validated against your permissions
- Complete audit trail with your user identity

When using HTTP transport with authentication enabled, your `PLANTON_API_KEY` serves as the bearer token for accessing the MCP server endpoint.

## Documentation

- [HTTP Transport Guide](docs/http-transport.md) - Running the server locally and HTTP deployment
- [Configuration Guide](docs/configuration.md) - Complete environment variable reference
- [Development Guide](docs/development.md) - Contributing and local development setup
- [Installation Guide](docs/installation.md) - Detailed installation instructions

## Support

- **Issues**: [GitHub Issues](https://github.com/plantoncloud-inc/mcp-server-planton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/plantoncloud-inc/mcp-server-planton/discussions)
- **Documentation**: [Planton Cloud Docs](https://docs.planton.cloud)

## License

Apache-2.0 - see [LICENSE](LICENSE) for details.

---

Built with ❤️ by [Planton Cloud](https://planton.cloud)
