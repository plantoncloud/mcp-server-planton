# Contributing to Planton Cloud MCP Server

Thank you for your interest in contributing to the Planton Cloud MCP Server! This document provides guidelines and instructions for contributing.

## Getting Started

### Prerequisites

- Go 1.22 or higher
- Git
- Docker (optional, for container testing)

### Setting Up Development Environment

1. Clone the repository:
```bash
git clone https://github.com/plantoncloud-inc/mcp-server-planton.git
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

## Development Workflow

### Running the Server Locally

Set required environment variables:
```bash
export USER_JWT_TOKEN="your-jwt-token"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"
```

Run the server:
```bash
./bin/mcp-server-planton
```

Or build and run directly:
```bash
go run ./cmd/mcp-server-planton
```

### Code Quality

We use standard Go tools for code quality.

Run tests:
```bash
make test
# or
go test -v ./...
```

Run linter (requires golangci-lint):
```bash
make lint
# or
golangci-lint run
```

Format code:
```bash
go fmt ./...
```

Vet code:
```bash
go vet ./...
```

### Code Style Guidelines

- Follow standard Go conventions and idioms
- Use `gofmt` for code formatting
- Write descriptive variable and function names
- Add comments for exported functions and types
- Keep functions focused and single-purpose
- Handle errors explicitly

## Adding New Tools

To add a new MCP tool:

1. **Create or update gRPC client** (if needed):

```go
// internal/grpc/organization_client.go
package grpc

import (
    "context"
    orgv1 "github.com/plantoncloud-inc/planton-cloud/apis/project/planton/provider/blintora/cloud/v1/blintora/cloud/organization/v1"
    "google.golang.org/grpc"
)

type OrganizationClient struct {
    conn   *grpc.ClientConn
    client orgv1.OrganizationQueryControllerClient
}

func NewOrganizationClient(grpcEndpoint, userToken string) (*OrganizationClient, error) {
    // Initialize gRPC client with auth interceptor
    // ...
}
```

2. **Implement the tool**:

```go
// internal/mcp/tools/organization.go
package tools

import (
    "context"
    "github.com/mark3labs/mcp-go/mcp"
)

func CreateOrganizationTool() mcp.Tool {
    return mcp.Tool{
        Name:        "list_organizations",
        Description: "List all organizations the user has access to",
        InputSchema: mcp.ToolInputSchema{
            Type:       "object",
            Properties: map[string]interface{}{},
        },
    }
}

func HandleListOrganizations(ctx context.Context, arguments map[string]interface{}, client *grpc.OrganizationClient) ([]mcp.Content, error) {
    // Implement tool handler
    // ...
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
```

4. **Update documentation**:
   - Add tool description to README.md
   - Document input/output schema
   - Provide usage examples

## Submitting Changes

### Pull Request Process

1. Fork the repository
2. Create a feature branch:
```bash
git checkout -b feature/your-feature-name
```

3. Make your changes
4. Run tests and linting:
```bash
make test lint
```

5. Commit your changes with clear, descriptive messages:
```bash
git commit -m "feat: add new tool for querying organizations"
```

6. Push to your fork:
```bash
git push origin feature/your-feature-name
```

7. Open a Pull Request against the `main` branch

### Commit Message Format

We follow conventional commit format:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding or updating tests
- `chore:` - Maintenance tasks

Examples:
```
feat: add organization query tool
fix: handle timeout errors in environment client
docs: update installation instructions
refactor: simplify error handling in tools
```

### Pull Request Guidelines

- Keep PRs focused on a single feature or fix
- Update documentation for any user-facing changes
- Ensure all checks pass (tests, linting)
- Provide clear description of changes
- Reference related issues if applicable

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

- Clear description of the issue
- Steps to reproduce
- Expected behavior vs actual behavior
- Environment details (Go version, OS, etc.)
- Relevant logs or error messages

### Feature Requests

When requesting features, please include:

- Clear description of the feature
- Use case and motivation
- Example of how it would be used
- Any relevant context or alternatives considered

## Questions and Support

- **GitHub Issues**: For bug reports and feature requests
- **GitHub Discussions**: For questions and general discussion

## Code of Conduct

- Be respectful and inclusive
- Assume good intentions
- Give and accept constructive feedback gracefully
- Focus on what's best for the community

## License

By contributing, you agree that your contributions will be licensed under the Apache-2.0 License.
