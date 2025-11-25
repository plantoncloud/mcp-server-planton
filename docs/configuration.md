# Configuration Guide

Comprehensive configuration options for the Planton Cloud MCP Server.

## Environment Variables

### Required Variables

#### PLANTON_API_KEY

User's API key for authentication with Planton Cloud APIs. This can be either a JWT token or an API key obtained from the Planton Cloud console.

```bash
export PLANTON_API_KEY="your-api-key-or-jwt-token"
```

**How to obtain:**

**From Web Console:**
1. Log in to Planton Cloud web console
2. Click on your profile icon in the top-right corner
3. Select **API Keys** from the menu
4. Click **Create Key** to generate a new API key
5. Copy the generated key

**Note:** Existing API keys may not be visible in the console for security reasons, so it's recommended to create a new key.

**Important:** This key represents the user's identity and permissions. Keep it secure and never commit it to version control.

### Optional Variables

#### PLANTON_CLOUD_ENVIRONMENT

Target environment for Planton Cloud APIs.

```bash
export PLANTON_CLOUD_ENVIRONMENT="live"
```

**Default:** `live`

**Valid values:**
- `live`: Production environment (`api.live.planton.cloud:443`)
- `test`: Test environment (`api.test.planton.cloud:443`)
- `local`: Local development (`localhost:8080`)

#### PLANTON_APIS_GRPC_ENDPOINT

Override gRPC endpoint for Planton Cloud APIs. This takes precedence over `PLANTON_CLOUD_ENVIRONMENT`.

```bash
export PLANTON_APIS_GRPC_ENDPOINT="custom-endpoint:443"
```

**Default:** Based on `PLANTON_CLOUD_ENVIRONMENT` setting

**When to use:**
- Custom or private Planton Cloud installations
- Non-standard endpoints
- Testing with specific backend instances

## Configuration Loading

The MCP server loads configuration from environment variables on startup using the Go standard library.

**Configuration struct:**

```go
type Config struct {
    PlantonAPIKey           string
    PlantonAPIsGRPCEndpoint string
}
```

**Loading process:**

```go
func LoadFromEnv() (*Config, error) {
    apiKey := os.Getenv("PLANTON_API_KEY")
    if apiKey == "" {
        return nil, fmt.Errorf(
            "PLANTON_API_KEY environment variable required. " +
            "This should be set by LangGraph when spawning MCP server",
        )
    }
    
    // Endpoint selection with priority:
    // 1. PLANTON_APIS_GRPC_ENDPOINT (explicit override)
    // 2. Based on PLANTON_CLOUD_ENVIRONMENT
    // 3. Default to "live" (api.live.planton.cloud:443)
    endpoint := getEndpoint()
    
    return &Config{
        PlantonAPIKey:           apiKey,
        PlantonAPIsGRPCEndpoint: endpoint,
    }, nil
}
```

## Configuration Files

### Environment Files

Create a `.env` file in your project root for local development:

```env
# Required
PLANTON_API_KEY=your-api-key-or-jwt-token

# Optional: Target environment (defaults to 'live')
PLANTON_CLOUD_ENVIRONMENT=live  # or 'test', 'local'

# Optional: Override endpoint (not needed for standard environments)
# PLANTON_APIS_GRPC_ENDPOINT=custom-endpoint:443
```

**Note:** The Go server doesn't automatically load `.env` files. You'll need to source them manually or use a tool like `direnv`:

```bash
# Source manually
export $(cat .env | xargs)

# Or use direnv
echo 'export PLANTON_API_KEY="..."' > .envrc
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
        "PLANTON_API_KEY": "${PLANTON_API_KEY}",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```

**Note:** LangGraph supports environment variable expansion with `${VAR}` syntax.

For local development:

```json
{
  "mcp_servers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "${PLANTON_API_KEY}",
        "PLANTON_CLOUD_ENVIRONMENT": "local"
      }
    }
  }
}
```

### Claude Desktop Configuration

In `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "your-api-key",
        "PLANTON_CLOUD_ENVIRONMENT": "live"
      }
    }
  }
}
```

**Note:** Claude Desktop requires actual values, not environment variable references.

For local development:

```json
{
  "mcpServers": {
    "planton-cloud": {
      "command": "mcp-server-planton",
      "env": {
        "PLANTON_API_KEY": "your-api-key",
        "PLANTON_CLOUD_ENVIRONMENT": "local"
      }
    }
  }
}
```

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

### API Key Management

1. **Never commit keys** - Use environment variables or secret managers
2. **Rotate regularly** - API keys should be rotated periodically
3. **Scope appropriately** - Use keys with minimum required permissions
4. **Monitor usage** - Track key usage for audit purposes

### Environment-Specific API Keys

Use different keys for different environments:

```bash
# Development
export PLANTON_API_KEY_DEV="dev-key..."

# Staging
export PLANTON_API_KEY_STAGING="staging-key..."

# Production
export PLANTON_API_KEY_PROD="prod-key..."
```

### Secret Management

For production deployments, use secret managers:

**AWS Secrets Manager:**
```bash
export PLANTON_API_KEY=$(aws secretsmanager get-secret-value \
  --secret-id planton/api-key \
  --query SecretString \
  --output text)
```

**HashiCorp Vault:**
```bash
export PLANTON_API_KEY=$(vault kv get -field=key secret/planton/api)
```

**Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: planton-mcp-config
type: Opaque
stringData:
  PLANTON_API_KEY: "your-api-key"
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
    fmt.Printf("API key present: %t\n", cfg.PlantonAPIKey != "")
}
```

## Troubleshooting

### Missing API Key Error

```
Configuration error: PLANTON_API_KEY environment variable required
```

**Solution:** Set the `PLANTON_API_KEY` environment variable.

```bash
export PLANTON_API_KEY="your-api-key-here"
```

### Invalid API Key Error

```
rpc error: code = Unauthenticated desc = Invalid authentication credentials
```

**Solutions:**
1. Verify API key is not expired
2. Ensure API key is complete (no truncation)
3. Generate a new API key from the Planton Cloud console (Profile → API Keys → Create Key)

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
export PLANTON_API_KEY="dev-key-from-local-planton"
export PLANTON_CLOUD_ENVIRONMENT="local"
mcp-server-planton
```

### Production

```bash
export PLANTON_API_KEY="$(vault kv get -field=key secret/planton/api)"
export PLANTON_CLOUD_ENVIRONMENT="live"
mcp-server-planton
```

### Test Environment

```bash
export PLANTON_API_KEY="test-api-key"
export PLANTON_CLOUD_ENVIRONMENT="test"
mcp-server-planton
```

### Docker Compose

```yaml
version: '3.8'
services:
  mcp-server:
    image: ghcr.io/plantoncloud-inc/mcp-server-planton:latest
    environment:
      PLANTON_API_KEY: ${PLANTON_API_KEY}
      PLANTON_CLOUD_ENVIRONMENT: live
    stdin_open: true
    tty: true
```

## Next Steps

- [Installation Guide](installation.md) - Installation instructions
- [Development Guide](development.md) - Development setup
- [README](../README.md) - Back to main documentation
