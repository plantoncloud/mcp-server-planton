# Planton Cloud MCP Server

MCP (Model Context Protocol) server for Planton Cloud that enables AI agents to query cloud resources using user permissions.

## Overview

The Planton Cloud MCP Server provides tools for LangGraph agents, Claude Desktop, and other MCP clients to interact with Planton Cloud resources. Unlike typical MCP servers that use API keys or machine accounts, this server uses **user JWT tokens**, ensuring that all resource queries respect Fine-Grained Authorization (FGA) based on the user's actual permissions.

### Key Features

- **User-scoped permissions** - Queries respect the user's actual permissions via JWT
- **Environment queries** - List and filter environments by organization
- **Extensible** - More resource types coming soon (organizations, projects, cloud resources)
- **MCP standard** - Works with any MCP client (LangGraph, Claude Desktop, Cursor, etc.)
- **Go implementation** - Fast, lightweight, and easy to distribute

## Installation

### Option 1: Docker (Recommended)

Pull and run from GitHub Container Registry:

```bash
docker run -i --rm \
  -e USER_JWT_TOKEN="your-jwt-token" \
  -e PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

### Option 2: Pre-built Binaries

Download the latest release for your platform from [GitHub Releases](https://github.com/plantoncloud-inc/mcp-server-planton/releases):

```bash
# macOS (ARM64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Darwin_arm64.tar.gz | tar xz

# macOS (Intel)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Darwin_x86_64.tar.gz | tar xz

# Linux (AMD64)
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Linux_x86_64.tar.gz | tar xz

# Move to PATH
sudo mv mcp-server-planton /usr/local/bin/
```

### Option 3: From Source

Requires Go 1.22+:

```bash
go install github.com/plantoncloud-inc/mcp-server-planton/cmd/mcp-server-planton@latest
```

Or clone and build:

```bash
git clone https://github.com/plantoncloud-inc/mcp-server-planton.git
cd mcp-server-planton
make build
```

## Quick Start

### Standalone Usage

Set required environment variables:

```bash
export USER_JWT_TOKEN="your-jwt-token"
export PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443"
```

Run the server:

```bash
mcp-server-planton
```

### Integration with LangGraph

Add to your `langgraph.json`:

```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "USER_JWT_TOKEN": "${USER_JWT_TOKEN}",
        "PLANTON_APIS_GRPC_ENDPOINT": "${PLANTON_APIS_GRPC_ENDPOINT}"
      }
    }
  }
}
```

Or using Docker:

```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "USER_JWT_TOKEN=${USER_JWT_TOKEN}",
        "-e", "PLANTON_APIS_GRPC_ENDPOINT=${PLANTON_APIS_GRPC_ENDPOINT}",
        "ghcr.io/plantoncloud-inc/mcp-server-planton:latest"
      ]
    }
  }
}
```

### Integration with Claude Desktop

Add to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "USER_JWT_TOKEN": "your-jwt-token",
        "PLANTON_APIS_GRPC_ENDPOINT": "apis.planton.cloud:443"
      }
    }
  }
}
```

Or using Docker:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "docker",
      "args": [
        "run", "-i", "--rm",
        "-e", "USER_JWT_TOKEN=your-jwt-token",
        "-e", "PLANTON_APIS_GRPC_ENDPOINT=apis.planton.cloud:443",
        "ghcr.io/plantoncloud-inc/mcp-server-planton:latest"
      ]
    }
  }
}
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `USER_JWT_TOKEN` | Yes | - | User's JWT token for authentication |
| `PLANTON_APIS_GRPC_ENDPOINT` | No | `localhost:8080` | Planton Cloud APIs gRPC endpoint |

### Getting a JWT Token

To obtain a JWT token:

1. Log in to Planton Cloud web console
2. Open browser developer tools (F12)
3. Go to Application/Storage → Local Storage
4. Find the authentication token
5. Copy the JWT value

For programmatic access, use the Planton Cloud CLI:

```bash
planton auth login
planton auth token
```

## Available Tools

### list_environments_for_org

List all environments available in an organization that the user has permission to view.

**Input:**
```json
{
  "org_id": "your-organization-id"
}
```

**Output:**
```json
[
  {
    "id": "env-123",
    "slug": "production",
    "name": "Production Environment",
    "description": "Production deployment environment"
  },
  {
    "id": "env-456",
    "slug": "staging",
    "name": "Staging Environment",
    "description": "Pre-production staging"
  }
]
```

**Error Handling:**

The tool returns user-friendly error messages for common issues:

- `UNAUTHENTICATED` - JWT token invalid or expired
- `PERMISSION_DENIED` - User lacks permission to view environments in the organization
- `NOT_FOUND` - Organization doesn't exist
- `UNAVAILABLE` - Planton Cloud APIs are temporarily unavailable

## Security Architecture

### User JWT Propagation

This MCP server follows a unique security pattern:

```
User → LangGraph/MCP Client → MCP Server (with JWT) → Planton Cloud APIs
```

**Key security properties:**

1. **No JWT persistence** - JWT is only held in memory during execution
2. **User permissions enforced** - APIs validate JWT and check FGA on every call
3. **Short-lived process** - MCP server exits when agent execution completes
4. **Audit trail** - All API calls are logged with user identity

This is different from machine account patterns where a service account has broad permissions. With user JWT propagation, the MCP server can only access what the user can access.

## Development

### Prerequisites

- Go 1.22+
- Access to Planton Cloud APIs (local or remote)

### Setup

```bash
# Clone repository
git clone https://github.com/plantoncloud-inc/mcp-server-planton.git
cd mcp-server-planton

# Install dependencies
go mod download

# Build
make build
```

### Running Locally

```bash
export USER_JWT_TOKEN="your-jwt-token"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"
./bin/mcp-server-planton
```

### Make Targets

```bash
make build          # Build binary
make install        # Install to GOPATH/bin
make test           # Run tests
make lint           # Run linter
make docker-build   # Build Docker image
make docker-run     # Run Docker container
make clean          # Remove build artifacts
```

### Project Structure

```
mcp-server-planton/
├── cmd/
│   └── mcp-server-planton/
│       └── main.go              # Entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── grpc/
│   │   ├── interceptor.go       # Auth interceptor
│   │   └── client.go            # Environment gRPC client
│   └── mcp/
│       ├── server.go            # MCP server setup
│       └── tools/
│           └── environment.go   # Environment query tools
├── .github/
│   └── workflows/
│       └── release.yml          # Release automation
├── .goreleaser.yaml             # GoReleaser config
├── Dockerfile                   # Multi-stage Docker build
├── Makefile                     # Build commands
└── README.md                    # This file
```

## Distribution

### GitHub Releases

Pre-built binaries for multiple platforms are available on [GitHub Releases](https://github.com/plantoncloud-inc/mcp-server-planton/releases):

- macOS (Intel and ARM64)
- Linux (AMD64 and ARM64)
- Windows (AMD64)

### Docker Images

Docker images are published to GitHub Container Registry:

```bash
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:latest
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:v0.1.0
```

Images are available for:
- `linux/amd64`
- `linux/arm64`

## Roadmap

- [x] Environment query tools
- [x] Go implementation
- [x] Docker distribution
- [x] GitHub Releases
- [ ] Organization query tools
- [ ] Project query tools
- [ ] Cloud resource query tools
- [ ] Resource mutation tools (create, update, delete)
- [ ] Caching and performance optimization
- [ ] Comprehensive test suite

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Contribution Guide

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linting (`make test lint`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to your fork (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Support

- **Documentation**: [docs/](docs/)
- **Issues**: [GitHub Issues](https://github.com/plantoncloud-inc/mcp-server-planton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/plantoncloud-inc/mcp-server-planton/discussions)

## License

Apache-2.0 - see [LICENSE](LICENSE) for details.

## Related Projects

- [Planton Cloud](https://planton.cloud) - Cloud infrastructure management platform
- [Project Planton](https://github.com/project-planton/project-planton) - Open-source deployment components
- [LangGraph](https://langchain-ai.github.io/langgraph/) - Framework for building stateful AI agents
- [Model Context Protocol](https://modelcontextprotocol.io) - Protocol for connecting AI models to data sources

---

Built with ❤️ by [Planton Cloud](https://planton.cloud)
