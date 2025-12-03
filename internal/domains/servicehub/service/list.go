package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/apiresource/apiresourcekind"
	"buf.build/gen/go/blintora/apis/protocolbuffers/go/ai/planton/commons/rpc"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

// ServiceSimple is a simplified representation of a Service for JSON serialization.
type ServiceSimple struct {
	ID             string             `json:"id"`
	Slug           string             `json:"slug"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Org            string             `json:"org"`
	GitRepo        GitRepoInfo        `json:"git_repo"`
	PipelineConfig PipelineConfigInfo `json:"pipeline_config,omitempty"`
}

// GitRepoInfo contains Git repository information.
type GitRepoInfo struct {
	OwnerName     string `json:"owner_name"`
	Name          string `json:"name"`
	DefaultBranch string `json:"default_branch"`
	BrowserURL    string `json:"browser_url"`
	CloneURL      string `json:"clone_url"`
	Provider      string `json:"provider"`
	ProjectRoot   string `json:"project_root,omitempty"`
}

// PipelineConfigInfo contains pipeline configuration information.
type PipelineConfigInfo struct {
	PipelineProvider    string `json:"pipeline_provider"`
	ImageBuildMethod    string `json:"image_build_method,omitempty"`
	ImageRepositoryPath string `json:"image_repository_path,omitempty"`
	DisablePipelines    bool   `json:"disable_pipelines,omitempty"`
}

// CreateListServicesForOrgTool creates the MCP tool definition for listing services.
func CreateListServicesForOrgTool() mcp.Tool {
	return mcp.Tool{
		Name: "list_services_for_org",
		Description: "List all services in an organization. " +
			"Returns service metadata including Git repository information and pipeline configuration. " +
			"Only returns services the user has permission to view.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID to list services for",
				},
			},
			Required: []string{"org_id"},
		},
	}
}

// HandleListServicesForOrg handles the MCP tool invocation for listing services.
//
// This function:
//  1. Creates ServiceClient with user's API key
//  2. Queries Planton Cloud Service Hub APIs for services
//  3. Converts protobuf responses to JSON-serializable structs
//  4. Returns formatted response or error message
func HandleListServicesForOrg(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: list_services_for_org")

	// Extract org_id from arguments
	orgID, ok := arguments["org_id"].(string)
	if !ok || orgID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "org_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client with per-user API key from context
	// For HTTP transport: API key extracted from Authorization header
	// For STDIO transport: API key from environment variable (fallback to config)
	client, err := clients.NewServiceClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		// Fallback to config API key for STDIO mode
		client, err = clients.NewServiceClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer client.Close()

	// Build find request with org filter
	findRequest := &apiresource.FindApiResourcesRequest{
		Page: &rpc.PageInfo{
			Num:  0,
			Size: 1000, // Get all services (reasonable limit)
		},
		Kind: apiresourcekind.ApiResourceKind_service,
		Org:  orgID,
	}

	// Query services
	serviceList, err := client.Find(ctx, findRequest)
	if err != nil {
		return errors.HandleGRPCError(err, ""), nil
	}

	// Convert protobuf objects to JSON-serializable structs
	services := make([]ServiceSimple, 0, len(serviceList.GetEntries()))
	for _, svc := range serviceList.GetEntries() {
		gitRepo := svc.GetSpec().GetGitRepo()
		pipelineCfg := svc.GetSpec().GetPipelineConfiguration()

		serviceSimple := ServiceSimple{
			ID:          svc.GetMetadata().GetId(),
			Slug:        svc.GetMetadata().GetSlug(),
			Name:        svc.GetMetadata().GetName(),
			Description: svc.GetSpec().GetDescription(),
			Org:         svc.GetMetadata().GetOrg(),
			GitRepo: GitRepoInfo{
				OwnerName:     gitRepo.GetOwnerName(),
				Name:          gitRepo.GetName(),
				DefaultBranch: gitRepo.GetDefaultBranch(),
				BrowserURL:    gitRepo.GetBrowserUrl(),
				CloneURL:      gitRepo.GetCloneUrl(),
				Provider:      gitRepo.GetGitRepoProvider().String(),
				ProjectRoot:   gitRepo.GetProjectRoot(),
			},
		}

		// Add pipeline config if present
		if pipelineCfg != nil {
			serviceSimple.PipelineConfig = PipelineConfigInfo{
				PipelineProvider:    pipelineCfg.GetPipelineProvider().String(),
				ImageBuildMethod:    pipelineCfg.GetImageBuildMethod().String(),
				ImageRepositoryPath: pipelineCfg.GetImageRepositoryPath(),
				DisablePipelines:    pipelineCfg.GetDisablePipelines(),
			}
		}

		services = append(services, serviceSimple)
	}

	log.Printf(
		"Tool completed: list_services_for_org, returned %d services for org %s",
		len(services),
		orgID,
	)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		errResp := errors.ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: fmt.Sprintf("Failed to marshal response: %v", err),
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}
