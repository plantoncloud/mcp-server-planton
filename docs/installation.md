# Installation Guide

This guide covers different ways to install and run the Planton Cloud MCP Server.

## Prerequisites

- Access to Planton Cloud (account and JWT token)
- For binary installation: No additional dependencies
- For Docker installation: Docker installed
- For source installation: Go 1.22 or higher

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

Download the latest release for your platform from [GitHub Releases](https://github.com/plantoncloud-inc/mcp-server-planton/releases):

**macOS (ARM64/Apple Silicon):**
```bash
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Darwin_arm64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
chmod +x /usr/local/bin/mcp-server-planton
```

**macOS (Intel):**
```bash
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Darwin_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
chmod +x /usr/local/bin/mcp-server-planton
```

**Linux (AMD64):**
```bash
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Linux_x86_64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
chmod +x /usr/local/bin/mcp-server-planton
```

**Linux (ARM64):**
```bash
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.1.0/mcp-server-planton_0.1.0_Linux_arm64.tar.gz | tar xz
sudo mv mcp-server-planton /usr/local/bin/
chmod +x /usr/local/bin/mcp-server-planton
```

**Windows (AMD64):**
```powershell
# Download from GitHub Releases
# Extract mcp-server-planton.exe
# Add to PATH or use full path
```

Verify installation:
```bash
mcp-server-planton --help
```

### Method 2: Docker (Recommended for Containers)

Pull and run from GitHub Container Registry:

```bash
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:latest

# Or pull specific version
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:v0.1.0
```

Run the container:

```bash
docker run -i --rm \
  -e USER_JWT_TOKEN="your-jwt-token" \
  -e PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443" \
  ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

### Method 3: Install with go install

If you have Go installed:

```bash
go install github.com/plantoncloud-inc/mcp-server-planton/cmd/mcp-server-planton@latest
```

The binary will be installed to `$GOPATH/bin` (typically `~/go/bin`).

### Method 4: Install from Source

For development or latest features:

```bash
# Clone repository
git clone https://github.com/plantoncloud-inc/mcp-server-planton.git
cd mcp-server-planton

# Build
make build

# Install to system
sudo cp bin/mcp-server-planton /usr/local/bin/
```

## Configuration

### 1. Obtain JWT Token

Get your Planton Cloud JWT token:

**Option A: From Web Console**
1. Log in to Planton Cloud web console
2. Open browser developer tools (F12)
3. Navigate to Application → Local Storage
4. Find and copy the authentication token

**Option B: From CLI**
```bash
planton auth login
planton auth token
```

### 2. Set Environment Variables

Create environment variables for configuration:

```bash
export USER_JWT_TOKEN="your-jwt-token-here"
export PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443"
```

For local development against a local Planton Cloud instance:

```bash
export USER_JWT_TOKEN="your-jwt-token-here"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"
```

### 3. Verify Installation

Test the server:

```bash
# With environment variables set
mcp-server-planton
```

You should see log output indicating the server started successfully.

## Integration Setup

### LangGraph Integration

Add to your `langgraph.json`:

**Using binary:**
```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "USER_JWT_TOKEN": "${USER_JWT_TOKEN}",
        "PLANTON_APIS_GRPC_ENDPOINT": "apis.planton.cloud:443"
      }
    }
  }
}
```

**Using Docker:**
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

### Claude Desktop Integration

Add to Claude Desktop MCP configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

**Using binary:**
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

**Using Docker:**
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

### Cursor Integration

Add to your Cursor MCP settings:

```json
{
  "mcp": {
    "servers": {
      "planton-cloud": {
        "command": "mcp-server-planton",
        "env": {
          "USER_JWT_TOKEN": "your-jwt-token",
          "PLANTON_APIS_GRPC_ENDPOINT": "apis.planton.cloud:443"
        }
      }
    }
  }
}
```

## Troubleshooting

### Binary Not Found

If you get "command not found" errors:

```bash
# Verify binary is in PATH
which mcp-server-planton

# If not, add to PATH or use full path
export PATH="$PATH:/usr/local/bin"

# Or use full path
/usr/local/bin/mcp-server-planton
```

### Connection Issues

If the server can't connect to Planton Cloud APIs:

1. Verify the endpoint is correct:
   - Production: `apis.planton.cloud:443`
   - Local: `localhost:8080`

2. Check network connectivity:
```bash
ping apis.planton.cloud
```

3. Verify JWT token is valid:
```bash
# Check token expiration
echo $USER_JWT_TOKEN | cut -d'.' -f2 | base64 -d
```

### Permission Errors

If you get permission denied errors:

1. Verify your JWT token is current (not expired)
2. Check that your user has permissions in the organization
3. Contact your Planton Cloud administrator

### Docker Issues

If Docker commands fail:

```bash
# Verify Docker is running
docker ps

# Check Docker logs
docker logs <container-id>

# Pull latest image
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

## Updating

### Update Binary

Download the latest release from GitHub:

```bash
# Download new version
curl -L https://github.com/plantoncloud-inc/mcp-server-planton/releases/download/v0.2.0/mcp-server-planton_0.2.0_Darwin_arm64.tar.gz | tar xz

# Replace existing binary
sudo mv mcp-server-planton /usr/local/bin/
```

### Update Docker Image

```bash
docker pull ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

### Update from Source

```bash
cd mcp-server-planton
git pull
make build
sudo cp bin/mcp-server-planton /usr/local/bin/
```

## Uninstallation

### Remove Binary

```bash
sudo rm /usr/local/bin/mcp-server-planton
```

### Remove Docker Image

```bash
docker rmi ghcr.io/plantoncloud-inc/mcp-server-planton:latest
```

### Remove Source Build

```bash
rm -rf mcp-server-planton/
```

## Next Steps

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Development Guide](development.md) - Contributing and development setup
- [README](../README.md) - Back to main documentation
