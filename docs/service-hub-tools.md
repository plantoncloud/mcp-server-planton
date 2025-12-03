# Service Hub Tools

This document provides detailed information about Service Hub tools available in the Planton Cloud MCP Server.

## Overview

Service Hub tools enable AI agents to query and understand services, pipelines, and Git integrations in Planton Cloud. These tools are essential for helping users create and modify Tekton pipelines by providing context about their services and repositories.

## Service Tools

### list_services_for_org

Lists all services in an organization with Git repository information and pipeline configuration.

**Input:**
```json
{
  "org_id": "org-abc123"
}
```

**Output:**
```json
[
  {
    "id": "svc-xyz789",
    "slug": "backend-api",
    "name": "Backend API",
    "description": "Main backend service",
    "org": "org-abc123",
    "git_repo": {
      "owner_name": "acmecorp",
      "name": "backend-api",
      "default_branch": "main",
      "browser_url": "https://github.com/acmecorp/backend-api",
      "clone_url": "https://github.com/acmecorp/backend-api.git",
      "provider": "GIT_REPO_PROVIDER_GITHUB",
      "project_root": ""
    },
    "pipeline_config": {
      "pipeline_provider": "PIPELINE_PROVIDER_PLATFORM",
      "image_build_method": "IMAGE_BUILD_METHOD_DOCKERFILE",
      "image_repository_path": "backend-api",
      "disable_pipelines": false
    }
  }
]
```

**Use Cases:**
- List all services to help users choose which service to work with
- Understand service configuration before creating pipelines
- Discover services in an organization

### get_service_by_id

Gets detailed information about a service by its ID.

**Input:**
```json
{
  "service_id": "svc-xyz789"
}
```

**Output:**
Same structure as single service in `list_services_for_org`.

**Use Cases:**
- Get complete service details when you have the service ID
- Understand Git repository configuration for pipeline creation
- Check pipeline provider settings

### get_service_by_org_by_slug

Gets service details by organization and service name/slug.

**Input:**
```json
{
  "org_id": "org-abc123",
  "slug": "backend-api"
}
```

**Output:**
Same structure as `get_service_by_id`.

**Use Cases:**
- Look up service when you know the name but not the ID
- Useful for CLI-style interactions where users provide service names

### list_service_branches

Lists all Git branches for a service's repository.

**Input:**
```json
{
  "service_id": "svc-xyz789"
}
```

**Output:**
```json
{
  "branches": [
    "main",
    "develop",
    "feature/new-api",
    "hotfix/critical-bug"
  ]
}
```

**Use Cases:**
- Help users select which branch to configure pipelines for
- Validate that a specified branch exists
- Display available branches in conversational context

## Tekton Pipeline Tools

### list_tekton_pipelines

Lists available Tekton pipeline templates (platform-provided and organization-specific).

**Input:**
```json
{
  "org_id": "org-abc123"  // Optional - if empty, shows platform-provided pipelines
}
```

**Output:**
```json
[
  {
    "id": "tknpipe-platform-01",
    "slug": "nodejs-docker",
    "name": "Node.js Docker Build",
    "description": "Build Node.js applications using Docker",
    "org": "",
    "tags": ["nodejs", "docker", "platform"]
  },
  {
    "id": "tknpipe-custom-01",
    "slug": "custom-build",
    "name": "Custom Build Pipeline",
    "description": "Organization-specific build process",
    "org": "org-abc123",
    "tags": ["custom", "golang"]
  }
]
```

**Use Cases:**
- Browse available pipeline templates before creating custom pipelines
- Discover platform-provided pipelines
- Find organization-specific pipeline templates
- Filter pipelines by tags (language, build method, compliance)

### get_tekton_pipeline

Gets complete Tekton pipeline definition including YAML content.

**Input (by ID):**
```json
{
  "pipeline_id": "tknpipe-platform-01"
}
```

**Input (by org and slug):**
```json
{
  "org_id": "org-abc123",
  "slug": "nodejs-docker"
}
```

**Output:**
```json
{
  "id": "tknpipe-platform-01",
  "slug": "nodejs-docker",
  "name": "Node.js Docker Build",
  "description": "Build Node.js applications using Docker",
  "org": "",
  "tags": ["nodejs", "docker", "platform"],
  "yaml_content": "apiVersion: tekton.dev/v1beta1\nkind: Pipeline\nmetadata:\n  name: nodejs-docker\nspec:\n  ..."
}
```

**Use Cases:**
- Get pipeline YAML to customize for a service
- Understand pipeline structure before modification
- Copy platform-provided pipelines as starting templates
- View complete pipeline definitions

## GitHub Credential Tools

### get_github_credential_for_service

Gets GitHub credential details associated with a service (metadata only, no access tokens).

**Input:**
```json
{
  "service_id": "svc-xyz789"
}
```

**Output:**
```json
{
  "id": "ghcred-abc123",
  "slug": "acmecorp-github",
  "name": "Acme Corp GitHub",
  "org": "org-abc123",
  "installation_id": 12345678,
  "account_id": "acmecorp",
  "account_type": "GITHUB_APP_ACCOUNT_TYPE_ORGANIZATION",
  "connection_host": "https://github.com"
}
```

**Use Cases:**
- Understand which GitHub account is connected to a service
- Get credential information for repository operations
- Verify GitHub integration is properly configured

### get_github_credential_by_org_by_slug

Gets GitHub credential by organization and credential name/slug.

**Input:**
```json
{
  "org_id": "org-abc123",
  "slug": "acmecorp-github"
}
```

**Output:**
Same structure as `get_github_credential_for_service`.

**Use Cases:**
- Look up credential when you know the credential name
- Useful when users want to work with a specific GitHub connection

### list_github_repositories

Lists all GitHub repositories accessible via a GitHub credential.

**Input:**
```json
{
  "credential_id": "ghcred-abc123"
}
```

**Output:**
```json
[
  {
    "owner": "acmecorp",
    "name": "backend-api",
    "browser_url": "https://github.com/acmecorp/backend-api",
    "clone_url": "https://github.com/acmecorp/backend-api.git"
  },
  {
    "owner": "acmecorp",
    "name": "frontend-app",
    "browser_url": "https://github.com/acmecorp/frontend-app",
    "clone_url": "https://github.com/acmecorp/frontend-app.git"
  }
]
```

**Use Cases:**
- Discover available repositories when onboarding a new service
- List all repositories accessible with a credential
- Help users choose which repository to connect

## Agent Workflow Examples

### Example 1: Creating a Tekton Pipeline for a Service

**User:** "I want to add a custom Tekton pipeline to my backend-api service"

**Agent workflow:**

1. **List services to identify the service:**
   ```json
   list_services_for_org({ "org_id": "org-abc123" })
   ```

2. **Get service details:**
   ```json
   get_service_by_org_by_slug({ "org_id": "org-abc123", "slug": "backend-api" })
   ```

3. **List available Tekton pipeline templates:**
   ```json
   list_tekton_pipelines({ "org_id": "org-abc123" })
   ```

4. **Get a template pipeline to use as base:**
   ```json
   get_tekton_pipeline({ "pipeline_id": "tknpipe-platform-01" })
   ```

5. **List Git branches to configure pipeline for:**
   ```json
   list_service_branches({ "service_id": "svc-xyz789" })
   ```

6. **Agent can now:**
   - Help user customize the pipeline YAML
   - Suggest appropriate configuration based on service details
   - Guide user on where to commit the pipeline file

### Example 2: Understanding Service Git Integration

**User:** "Which GitHub account is my backend-api connected to?"

**Agent workflow:**

1. **Get service details:**
   ```json
   get_service_by_org_by_slug({ "org_id": "org-abc123", "slug": "backend-api" })
   ```

2. **Get GitHub credential for the service:**
   ```json
   get_github_credential_for_service({ "service_id": "svc-xyz789" })
   ```

3. **Agent responds with:**
   - GitHub account name
   - Installation ID
   - Connection host

### Example 3: Exploring Available Repositories

**User:** "What repositories can I access with my GitHub connection?"

**Agent workflow:**

1. **Get GitHub credential by name:**
   ```json
   get_github_credential_by_org_by_slug({ "org_id": "org-abc123", "slug": "acmecorp-github" })
   ```

2. **List repositories:**
   ```json
   list_github_repositories({ "credential_id": "ghcred-abc123" })
   ```

3. **Agent presents:**
   - List of accessible repositories
   - Browser URLs for each repository
   - Clone URLs for Git operations

## Security and Permissions

All Service Hub tools respect Fine-Grained Authorization (FGA):

- **Service Access**: Users can only view services in organizations they're members of
- **Credential Access**: Users can only view credentials they have permission to access
- **Repository Listing**: Only repositories accessible via the user's credentials are shown
- **Pipeline Access**: Platform pipelines are public; organization pipelines require membership

Every API call is authenticated with the user's API key and validated against their actual permissions.

## Next Steps

After querying services and pipelines with these tools, the next phase will add:
- Git operations tools (clone, commit, create PR)
- Sandbox management for safe Git operations
- Pipeline deployment and testing tools

These capabilities are coming soon to enable complete end-to-end pipeline creation workflows.

## Related Documentation

- [MCP Server README](../README.md) - General MCP server documentation
- [HTTP Transport Guide](http-transport.md) - HTTP deployment and configuration
- [Configuration Guide](configuration.md) - Environment variables and settings

