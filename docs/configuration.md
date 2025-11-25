# Configuration Guide

Comprehensive configuration options for the Planton Cloud MCP Server.

## Environment Variables

### Required Variables

#### USER_JWT_TOKEN

User's JWT token for authentication with Planton Cloud APIs.

```bash
export USER_JWT_TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**How to obtain:**
- Web Console: Developer Tools → Application → Local Storage
- CLI: `planton auth token`

**Important:** This token represents the user's identity and permissions. Keep it secure and never commit it to version control.

### Optional Variables

#### PLANTON_APIS_GRPC_ENDPOINT

gRPC endpoint for Planton Cloud APIs.

```bash
export PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443"
```

**Default:** `localhost:8080`

**Common values:**
- Production: `apis.planton.cloud:443`
- Staging: `staging.apis.planton.cloud:443`
- Local development: `localhost:8080`

## Configuration Loading

The MCP server loads configuration from environment variables on startup using the Go standard library.

**Configuration struct:**

```go
type Config struct {
    UserJWTToken            string
    PlantonAPIsGRPCEndpoint string
}
```

**Loading process:**

```go
func LoadFromEnv() (*Config, error) {
    userJWT := os.Getenv("USER_JWT_TOKEN")
    if userJWT == "" {
        return nil, fmt.Errorf(
            "USER_JWT_TOKEN environment variable required. " +
            "This should be set by LangGraph when spawning MCP server",
        )
    }
    
    endpoint := os.Getenv("PLANTON_APIS_GRPC_ENDPOINT")
    if endpoint == "" {
        endpoint = "localhost:8080"
    }
    
    return &Config{
        UserJWTToken:            userJWT,
        PlantonAPIsGRPCEndpoint: endpoint,
    }, nil
}
```

## Configuration Files

### Environment Files

Create a `.env` file in your project root for local development:

```env
# Required
USER_JWT_TOKEN=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

# Optional (defaults to localhost:8080)
PLANTON_APIS_GRPC_ENDPOINT=apis.planton.cloud:443
```

**Note:** The Go server doesn't automatically load `.env` files. You'll need to source them manually or use a tool like `direnv`:

```bash
# Source manually
export $(cat .env | xargs)

# Or use direnv
echo 'export USER_JWT_TOKEN="..."' > .envrc
direnv allow
```

**Security note:** Add `.env` to your `.gitignore`:

```gitignore
.env
.env.local
.env.*.local
```

## Integration-Specific Configuration

### LangGraph Configuration

In `langgraph.json`:

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

**Note:** LangGraph supports environment variable expansion with `${VAR}` syntax.

### Claude Desktop Configuration

In `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "USER_JWT_TOKEN": "actual-token-here",
        "PLANTON_APIS_GRPC_ENDPOINT": "apis.planton.cloud:443"
      }
    }
  }
}
```

**Note:** Claude Desktop requires actual values, not environment variable references.

## Advanced Configuration

### Logging

The server uses Go's standard `log` package. Configure logging in the application:

```go
import "log"

// Set logging flags
log.SetFlags(log.LstdFlags | log.Lshortfile)

// Log with standard methods
log.Println("Server starting...")
log.Printf("Connecting to: %s", endpoint)
log.Fatalf("Fatal error: %v", err)
```

**Logging levels:**

The standard Go approach doesn't have log levels, but you can control output:

```bash
# Redirect to file
mcp-server-planton 2> server.log

# Suppress logs (not recommended)
mcp-server-planton 2>/dev/null
```

For structured logging in production, consider adding a logging library like `zerolog` or `zap`.

### TLS/SSL Configuration

For production deployments with TLS:

```bash
# Use secure endpoint (port 443)
export PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443"
```

The gRPC client automatically detects and uses TLS for standard HTTPS ports.

For custom TLS certificates:

```go
import (
    "crypto/tls"
    "crypto/x509"
    "io/ioutil"
    
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials"
)

// Load custom certificates
cert, err := ioutil.ReadFile("ca.pem")
if err != nil {
    log.Fatal(err)
}

certPool := x509.NewCertPool()
certPool.AppendCertsFromPEM(cert)

tlsConfig := &tls.Config{
    RootCAs: certPool,
}

creds := credentials.NewTLS(tlsConfig)
opts := []grpc.DialOption{
    grpc.WithTransportCredentials(creds),
}

conn, err := grpc.Dial(endpoint, opts...)
```

### Timeout Configuration

Default gRPC timeout is context-based. To customize timeouts:

```go
import (
    "context"
    "time"
)

// Set timeout on context
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Use context in gRPC calls
resp, err := client.ListEnvironments(ctx, req)
```

### Connection Pooling

gRPC connections in Go are designed to be long-lived and multiplexed. The default behavior is optimal for most use cases:

```go
// Single connection handles multiple concurrent RPCs
conn, err := grpc.NewClient(endpoint, opts...)

// Connection is reused across multiple calls
client := environmentv1.NewEnvironmentQueryControllerClient(conn)
```

## Security Best Practices

### JWT Token Management

1. **Never commit tokens** - Use environment variables or secret managers
2. **Rotate regularly** - JWT tokens should have expiration times
3. **Scope appropriately** - Use tokens with minimum required permissions
4. **Monitor usage** - Track token usage for audit purposes

### Environment-Specific Tokens

Use different tokens for different environments:

```bash
# Development
export USER_JWT_TOKEN_DEV="dev-token..."

# Staging
export USER_JWT_TOKEN_STAGING="staging-token..."

# Production
export USER_JWT_TOKEN_PROD="prod-token..."
```

### Secret Management

For production deployments, use secret managers:

**AWS Secrets Manager:**
```bash
export USER_JWT_TOKEN=$(aws secretsmanager get-secret-value \
  --secret-id planton/jwt-token \
  --query SecretString \
  --output text)
```

**HashiCorp Vault:**
```bash
export USER_JWT_TOKEN=$(vault kv get -field=token secret/planton/jwt)
```

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: planton-mcp-config
type: Opaque
stringData:
  USER_JWT_TOKEN: "your-jwt-token"
  PLANTON_APIS_GRPC_ENDPOINT: "apis.planton.cloud:443"
```

**Using secrets in Pod:**
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: mcp-server
spec:
  containers:
  - name: mcp-server
    image: ghcr.io/plantoncloud-inc/mcp-server-planton:latest
    envFrom:
    - secretRef:
        name: planton-mcp-config
```

## Validation

Verify your configuration programmatically:

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/plantoncloud-inc/mcp-server-planton/internal/config"
)

func main() {
    cfg, err := config.LoadFromEnv()
    if err != nil {
        log.Fatalf("Configuration error: %v", err)
    }
    
    fmt.Println("Configuration valid!")
    fmt.Printf("Endpoint: %s\n", cfg.PlantonAPIsGRPCEndpoint)
    fmt.Printf("Token present: %t\n", cfg.UserJWTToken != "")
}
```

## Troubleshooting

### Missing Token Error

```
Configuration error: USER_JWT_TOKEN environment variable required
```

**Solution:** Set the `USER_JWT_TOKEN` environment variable.

```bash
export USER_JWT_TOKEN="your-token-here"
```

### Invalid Token Error

```
rpc error: code = Unauthenticated desc = Invalid authentication credentials
```

**Solutions:**
1. Verify token is not expired
2. Ensure token is complete (no truncation)
3. Re-authenticate and get a new token

### Connection Refused

```
rpc error: code = Unavailable desc = connection error: desc = "transport: Error while dialing dial tcp: connect: connection refused"
```

**Solutions:**
1. Verify endpoint is correct
2. Check network connectivity
3. Ensure firewall allows gRPC traffic
4. Try pinging the endpoint

### Context Deadline Exceeded

```
rpc error: code = DeadlineExceeded desc = context deadline exceeded
```

**Solutions:**
1. Check network latency
2. Increase timeout in context
3. Verify API server is responding

## Configuration Examples

### Local Development

```bash
export USER_JWT_TOKEN="dev-token-from-local-planton"
export PLANTON_APIS_GRPC_ENDPOINT="localhost:8080"
mcp-server-planton
```

### Production

```bash
export USER_JWT_TOKEN="$(vault kv get -field=token secret/planton/jwt)"
export PLANTON_APIS_GRPC_ENDPOINT="apis.planton.cloud:443"
mcp-server-planton
```

### Docker Compose

```yaml
version: '3.8'
services:
  mcp-server:
    image: ghcr.io/plantoncloud-inc/mcp-server-planton:latest
    environment:
      USER_JWT_TOKEN: ${USER_JWT_TOKEN}
      PLANTON_APIS_GRPC_ENDPOINT: apis.planton.cloud:443
    stdin_open: true
    tty: true
```

## Next Steps

- [Installation Guide](installation.md) - Installation instructions
- [Development Guide](development.md) - Development setup
- [README](../README.md) - Back to main documentation
