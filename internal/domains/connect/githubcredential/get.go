package githubcredential

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/common/errors"
	"github.com/plantoncloud-inc/mcp-server-planton/internal/config"
	connectclients "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/connect/clients"
	servicehubclients "github.com/plantoncloud-inc/mcp-server-planton/internal/domains/servicehub/clients"
)

// GithubCredentialInfo contains GitHub credential information (metadata only, no secrets).
type GithubCredentialInfo struct {
	ID             string `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Org            string `json:"org"`
	InstallationID int64  `json:"installation_id"`
	AccountID      string `json:"account_id"`
	AccountType    string `json:"account_type"`
	ConnectionHost string `json:"connection_host"`
}

// CreateGetGithubCredentialForServiceTool creates the MCP tool definition for getting GitHub credential for a service.
func CreateGetGithubCredentialForServiceTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_github_credential_for_service",
		Description: "Get GitHub credential details associated with a service. " +
			"Returns credential metadata but not the actual access token. " +
			"Useful for understanding which GitHub account is connected to a service.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"service_id": map[string]interface{}{
					"type":        "string",
					"description": "Service ID",
				},
			},
			Required: []string{"service_id"},
		},
	}
}

// HandleGetGithubCredentialForService handles the MCP tool invocation for getting GitHub credential for a service.
func HandleGetGithubCredentialForService(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_github_credential_for_service")

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

	// Create Service client to get service details
	serviceClient, err := servicehubclients.NewServiceClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		serviceClient, err = servicehubclients.NewServiceClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create service gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer serviceClient.Close()

	// Get service to extract GitHub credential ID
	service, err := serviceClient.GetById(ctx, serviceID)
	if err != nil {
		return errors.HandleGRPCError(err, serviceID), nil
	}

	// Extract GitHub credential ID from service spec
	githubRepo := service.GetSpec().GetGitRepo().GetGithubRepo()
	if githubRepo == nil {
		errResp := errors.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Service is not connected to a GitHub repository",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	credentialID := githubRepo.GetGithubCredentialId()
	if credentialID == "" {
		errResp := errors.ErrorResponse{
			Error:   "NOT_FOUND",
			Message: "Service does not have a GitHub credential configured",
		}
		errJSON, _ := json.MarshalIndent(errResp, "", "  ")
		return mcp.NewToolResultText(string(errJSON)), nil
	}

	// Create GitHub credential client
	credClient, err := connectclients.NewGithubCredentialClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		credClient, err = connectclients.NewGithubCredentialClient(
			cfg.PlantonAPIsGRPCEndpoint,
			cfg.PlantonAPIKey,
		)
		if err != nil {
			errResp := errors.ErrorResponse{
				Error:   "CLIENT_ERROR",
				Message: fmt.Sprintf("Failed to create GitHub credential gRPC client: %v", err),
			}
			errJSON, _ := json.MarshalIndent(errResp, "", "  ")
			return mcp.NewToolResultText(string(errJSON)), nil
		}
	}
	defer credClient.Close()

	// Get GitHub credential
	credential, err := credClient.GetById(ctx, credentialID)
	if err != nil {
		return errors.HandleGRPCError(err, credentialID), nil
	}

	// Convert to info struct (metadata only, no secrets)
	appInstallInfo := credential.GetSpec().GetAppInstallInfo()
	credInfo := GithubCredentialInfo{
		ID:             credential.GetMetadata().GetId(),
		Slug:           credential.GetMetadata().GetSlug(),
		Name:           credential.GetMetadata().GetName(),
		Org:            credential.GetMetadata().GetOrg(),
		InstallationID: appInstallInfo.GetInstallationId(),
		AccountID:      appInstallInfo.GetAccountId(),
		AccountType:    appInstallInfo.GetAccountType().String(),
		ConnectionHost: credential.GetSpec().GetGithubConnectionHost(),
	}

	log.Printf("Tool completed: get_github_credential_for_service, credential: %s", credentialID)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(credInfo, "", "  ")
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

// CreateGetGithubCredentialByOrgBySlugTool creates the MCP tool definition for getting GitHub credential by org and slug.
func CreateGetGithubCredentialByOrgBySlugTool() mcp.Tool {
	return mcp.Tool{
		Name: "get_github_credential_by_org_by_slug",
		Description: "Get GitHub credential by organization and credential name/slug. " +
			"Use this when you know the credential name. " +
			"Returns credential metadata but not the actual access token.",
		InputSchema: mcp.ToolInputSchema{
			Type: "object",
			Properties: map[string]interface{}{
				"org_id": map[string]interface{}{
					"type":        "string",
					"description": "Organization ID",
				},
				"slug": map[string]interface{}{
					"type":        "string",
					"description": "Credential slug/name",
				},
			},
			Required: []string{"org_id", "slug"},
		},
	}
}

// HandleGetGithubCredentialByOrgBySlug handles the MCP tool invocation for getting GitHub credential by org and slug.
func HandleGetGithubCredentialByOrgBySlug(
	ctx context.Context,
	arguments map[string]interface{},
	cfg *config.Config,
) (*mcp.CallToolResult, error) {
	log.Printf("Tool invoked: get_github_credential_by_org_by_slug")

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

	// Create GitHub credential client
	client, err := connectclients.NewGithubCredentialClientFromContext(ctx, cfg.PlantonAPIsGRPCEndpoint)
	if err != nil {
		client, err = connectclients.NewGithubCredentialClient(
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

	// Get GitHub credential
	credential, err := client.GetByOrgBySlug(ctx, orgID, slug)
	if err != nil {
		return errors.HandleGRPCError(err, fmt.Sprintf("%s/%s", orgID, slug)), nil
	}

	// Convert to info struct (metadata only, no secrets)
	appInstallInfo := credential.GetSpec().GetAppInstallInfo()
	credInfo := GithubCredentialInfo{
		ID:             credential.GetMetadata().GetId(),
		Slug:           credential.GetMetadata().GetSlug(),
		Name:           credential.GetMetadata().GetName(),
		Org:            credential.GetMetadata().GetOrg(),
		InstallationID: appInstallInfo.GetInstallationId(),
		AccountID:      appInstallInfo.GetAccountId(),
		AccountType:    appInstallInfo.GetAccountType().String(),
		ConnectionHost: credential.GetSpec().GetGithubConnectionHost(),
	}

	log.Printf("Tool completed: get_github_credential_by_org_by_slug, credential: %s/%s", orgID, slug)

	// Return formatted JSON response
	resultJSON, err := json.MarshalIndent(credInfo, "", "  ")
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
