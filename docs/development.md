# Development Guide

Guide for contributing to and developing the Planton Cloud MCP Server.

## Development Setup

### Prerequisites

- Go 1.22+
- Git
- Docker (optional, for container testing)
- golangci-lint (optional, for linting)
- Access to Planton Cloud APIs (local or remote)

### Initial Setup

1. Fork and clone the repository:

```bash
git clone https://github.com/YOUR_USERNAME/mcp-server-planton.git
cd mcp-server-planton
```

2. Install dependencies:

```bash
go mod download
```

3. Build the project:

```bash
make build
```

4. Set up environment variables:

```bash
export USER_JWT_TOKEN="your-jwt-token"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"
```

## Development Workflow

### Running the Server

```bash
# Set environment variables
export USER_JWT_TOKEN="your-jwt-token"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"

# Run server from binary
./bin/mcp-server-planton

# Or build and run directly
go run ./cmd/mcp-server-planton
```

### Code Quality Tools

#### Running Tests

```bash
# Run all tests
make test

# Or use go test directly
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run tests with coverage report
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Linting

```bash
# Run linter (requires golangci-lint)
make lint

# Or run golangci-lint directly
golangci-lint run

# Install golangci-lint if needed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

#### Code Formatting

```bash
# Format code
go fmt ./...

# Check for common issues
go vet ./...
```

#### Running All Checks

```bash
# Format, vet, test, and lint
go fmt ./... && go vet ./... && go test ./... && golangci-lint run
```

### Testing

The project uses standard Go testing patterns.

**Test structure:**

```
internal/
├── config/
│   └── config_test.go
├── grpc/
│   ├── client_test.go
│   └── interceptor_test.go
└── mcp/
    ├── server_test.go
    └── tools/
        └── environment_test.go
```

**Writing tests:**

```go
package config_test

import (
    "os"
    "testing"
    
    "github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

func TestLoadFromEnv(t *testing.T) {
    // Set up test environment
    os.Setenv("USER_JWT_TOKEN", "test-token")
    defer os.Unsetenv("USER_JWT_TOKEN")
    
    cfg, err := config.LoadFromEnv()
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    
    if cfg.UserJWTToken != "test-token" {
        t.Errorf("Expected token 'test-token', got: %s", cfg.UserJWTToken)
    }
}
```

## Project Structure

```
mcp-server-planton/
├── cmd/
│   └── mcp-server-planton/
│       └── main.go                 # Entry point
├── internal/
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── grpc/
│   │   ├── interceptor.go          # gRPC auth interceptor
│   │   └── client.go               # Environment gRPC client
│   └── mcp/
│       ├── server.go               # MCP server setup
│       └── tools/
│           └── environment.go      # Environment query tools
├── archive/
│   └── python/                     # Archived Python implementation
├── docs/                           # Documentation
│   ├── installation.md
│   ├── configuration.md
│   └── development.md
├── .github/
│   └── workflows/                  # CI/CD pipelines
│       └── release.yml
├── .goreleaser.yaml                # Multi-platform build config
├── Dockerfile                      # Multi-stage container build
├── Makefile                        # Build commands
├── go.mod                          # Go dependencies
├── go.sum                          # Dependency checksums
├── README.md                       # Main documentation
├── LICENSE                         # Apache-2.0 license
├── CONTRIBUTING.md                 # Contribution guidelines
└── .gitignore                      # Git ignore rules
```

## Adding New Features

### Adding a New MCP Tool

1. **Create or update gRPC client** (if needed):

```go
// internal/grpc/organization_client.go
package grpc

import (
    "context"
    
    orgv1 "github.com/plantoncloud-inc/planton-cloud/apis/project/planton/provider/blintora/cloud/v1/blintora/cloud/organization/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type OrganizationClient struct {
    conn   *grpc.ClientConn
    client orgv1.OrganizationQueryControllerClient
}

func NewOrganizationClient(grpcEndpoint, userToken string) (*OrganizationClient, error) {
    opts := []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithUnaryInterceptor(UserTokenAuthInterceptor(userToken)),
    }
    
    conn, err := grpc.NewClient(grpcEndpoint, opts...)
    if err != nil {
        return nil, err
    }
    
    client := orgv1.NewOrganizationQueryControllerClient(conn)
    
    return &OrganizationClient{
        conn:   conn,
        client: client,
    }, nil
}

func (c *OrganizationClient) ListOrganizations(ctx context.Context) ([]*orgv1.Organization, error) {
    resp, err := c.client.List(ctx, &orgv1.ListRequest{})
    if err != nil {
        return nil, err
    }
    return resp.Organizations, nil
}

func (c *OrganizationClient) Close() error {
    return c.conn.Close()
}
```

2. **Implement the tool**:

```go
// internal/mcp/tools/organization.go
package tools

import (
    "context"
    "encoding/json"
    
    "github.com/mark3labs/mcp-go/mcp"
    "github.com/plantoncloud-inc/mcp-server-planton/internal/grpc"
)

func CreateOrganizationTool() mcp.Tool {
    return mcp.Tool{
        Name: "list_organizations",
        Description: "List all organizations the user has access to",
        InputSchema: mcp.ToolInputSchema{
            Type:       "object",
            Properties: map[string]interface{}{},
        },
    }
}

func HandleListOrganizations(ctx context.Context, arguments map[string]interface{}, client *grpc.OrganizationClient) ([]mcp.Content, error) {
    orgs, err := client.ListOrganizations(ctx)
    if err != nil {
        return nil, err
    }
    
    jsonData, err := json.MarshalIndent(orgs, "", "  ")
    if err != nil {
        return nil, err
    }
    
    return []mcp.Content{
        {
            Type: "text",
            Text: string(jsonData),
        },
    }, nil
}
```

3. **Register in server**:

```go
// internal/mcp/server.go
func (s *Server) registerTools() {
    // Register environment tools
    s.registerEnvironmentTools()
    
    // Register organization tools
    s.registerOrganizationTools()
}

func (s *Server) registerOrganizationTools() {
    // Register list_organizations tool
    s.mcpServer.AddTool(
        tools.CreateOrganizationTool(),
        func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
            return s.handleListOrganizations(ctx, request)
        },
    )
}
```

4. **Update documentation**:
   - Add tool description to README.md
   - Document input/output schema
   - Provide usage examples

### Code Style Guidelines

- **Error handling**: Always handle errors explicitly, never ignore them
- **Context propagation**: Pass `context.Context` as first parameter for network calls
- **Documentation**: Add godoc comments for all exported types and functions
- **Naming**: Use descriptive names following Go conventions
- **Package organization**: Keep packages focused and cohesive

**Example:**

```go
// FetchResourceByID retrieves a cloud resource by its unique identifier.
//
// The function respects the user's permissions via the JWT token in the context.
// Returns an error if the resource doesn't exist or the user lacks permissions.
func FetchResourceByID(ctx context.Context, resourceID string, client ResourceClient) (*Resource, error) {
    if resourceID == "" {
        return nil, fmt.Errorf("resource ID cannot be empty")
    }
    
    resource, err := client.GetResource(ctx, resourceID)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch resource %s: %w", resourceID, err)
    }
    
    return resource, nil
}
```

## Debugging

### Enable Debug Logging

```go
// In main.go or any package
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)
```

### Debugging gRPC Calls

Enable gRPC debug logging:

```bash
export GRPC_GO_LOG_VERBOSITY_LEVEL=99
export GRPC_GO_LOG_SEVERITY_LEVEL=info
```

### Using Delve Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Debug the application
dlv debug ./cmd/mcp-server-planton

# Or attach to running process
dlv attach <pid>
```

### Using IDE Debuggers

Most Go IDEs (VS Code, GoLand, etc.) have excellent debugging support:

**VS Code**: Use the Go extension and create a launch configuration
**GoLand**: Built-in debugger with breakpoints and variable inspection

## Building and Distribution

### Local Build

```bash
# Build for current architecture
make build

# Build for specific architecture
GOOS=linux GOARCH=amd64 go build -o bin/mcp-server-planton ./cmd/mcp-server-planton
```

### Docker Build

```bash
# Build Docker image
make docker-build

# Run Docker image
make docker-run
```

### Multi-platform Build with GoReleaser

```bash
# Install GoReleaser
go install github.com/goreleaser/goreleaser@latest

# Build without publishing (snapshot)
goreleaser build --snapshot --clean

# Full release (requires tag)
git tag v0.2.0
goreleaser release --clean
```

## Releasing

### Version Bumping

Version is set via Git tags following semantic versioning.

### Creating a Release

1. Create a git tag:

```bash
git tag -a v0.2.0 -m "Release v0.2.0

- Add organization query tools
- Improve error handling
- Update documentation"

git push origin v0.2.0
```

2. GitHub Actions will automatically:
   - Build binaries for all platforms via GoReleaser
   - Build multi-arch Docker images
   - Push images to GitHub Container Registry
   - Create GitHub release with artifacts

## Continuous Integration

The project uses GitHub Actions for CI/CD:

- **release.yml**: Runs on tag push
  - Builds binaries for multiple platforms
  - Builds and publishes Docker images
  - Creates GitHub release

## Common Issues

### Module Cache Issues

If you get module-related errors:

```bash
go clean -modcache
go mod download
```

### gRPC Connection Issues

Test gRPC connection:

```go
package main

import (
    "context"
    "log"
    "time"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    conn, err := grpc.DialContext(ctx, "localhost:8080",
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithBlock(),
    )
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()
    
    log.Println("Connection successful!")
}
```

### Build Issues

```bash
# Clean build cache
go clean -cache

# Rebuild everything
go build -a ./cmd/mcp-server-planton
```

## Resources

- [MCP Protocol Documentation](https://modelcontextprotocol.io)
- [gRPC Go Guide](https://grpc.io/docs/languages/go/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Planton Cloud Documentation](https://docs.planton.cloud)

## Getting Help

- **Issues**: [GitHub Issues](https://github.com/plantoncloud-inc/mcp-server-planton/issues)
- **Discussions**: [GitHub Discussions](https://github.com/plantoncloud-inc/mcp-server-planton/discussions)
- **Contributing**: See [CONTRIBUTING.md](../CONTRIBUTING.md)
