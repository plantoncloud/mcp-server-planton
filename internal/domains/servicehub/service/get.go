package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud/mcp-server-planton/internal/config"
	"github.com/plantoncloud/mcp-server-planton/internal/domains/servicehub/clients"
)

// CreateGetServiceByIdTool creates the MCP tool definition for getting service by ID.
func CreateGetServiceByIdTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_service_by_id",
		Description: "Get detailed information about a service by its ID. " +
			"Returns complete service configuration including Git repo, pipeline settings, and deployment status.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID (e.g., 'svc-abc123')",
				},
			},
			Required: []string{"service_id"},
		},
	}
}

// HandleGetServiceById handles the MCP tool invocation for getting service by ID.
func HandleGetServiceById(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_service_by_id")

	// Extract service_id from arguments
	serviceID, ok := arguments["service_id"].(string)
	if !ok || serviceID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "service_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client
	client, err := clients.NewServiceClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
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

	// Query service
	svc, err := client.GetById(ctx, serviceID)
	if err != nil {
		return errors.HandleGRPCError(err, serviceID), nil
	}

	// Convert to simple struct
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

	log.Printf("Tool completed: get_service_by_id, service: %s", serviceID)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(serviceSimple, "", "  ")
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

// CreateGetServiceByOrgBySlugTool creates the MCP tool definition for getting service by org and slug.
func CreateGetServiceByOrgBySlugTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_service_by_org_by_slug",
		Description: "Get service details by organization and service name/slug. " +
			"Useful when you know the service name but not the ID.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID",
				},
				"slug": map[string]interface{}{
					"type":        "string",
					"description": "Service slug/name",
				},
			},
			Required: []string{"org_id", "slug"},
		},
	}
}

// HandleGetServiceByOrgBySlug handles the MCP tool invocation for getting service by org and slug.
func HandleGetServiceByOrgBySlug(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_service_by_org_by_slug")

	// Extract arguments
	orgID, okOrg := arguments["org_id"].(string)
	slug, okSlug := arguments["slug"].(string)

	if !okOrg || orgID == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "org_id is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	if !okSlug || slug == "" {
		errResp := errors.ErrorResponse{
			Error:   "INVALID_ARGUMENT",
			Message: "slug is required",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create gRPC client
	client, err := clients.NewServiceClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
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

	// Query service
	svc, err := client.GetByOrgBySlug(ctx, orgID, slug)
	if err != nil {
		return errors.HandleGRPCError(err, fmt.Sprintf("%s/%s", orgID, slug)), nil
	}

	// Convert to simple struct
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

	log.Printf("Tool completed: get_service_by_org_by_slug, service: %s/%s", orgID, slug)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(serviceSimple, "", "  ")
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












