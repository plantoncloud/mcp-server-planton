package cloudresource

import (
	"fmt"
	"strings"
	"unicode"

	cloudresourcekind "buf.build/gen/go/project-planton/apis/protocolbuffers/go/org/project_planton/shared/cloudresourcekind"
)

// NormalizeCloudResourceKind converts various input formats to the canonical CloudResourceKind enum value.
//
// Accepts:
//   - snake_case: "aws_rds_instance", "kubernetes_deployment"
//   - PascalCase: "AwsRdsInstance", "KubernetesDeployment"
//   - Natural language with spaces: "AWS RDS Instance", "Kubernetes Deployment"
//   - Hyphenated: "aws-rds-instance", "kubernetes-deployment"
//
// Returns the CloudResourceKind enum value or error if not found.
func NormalizeCloudResourceKind(input string) (cloudresourcekind.CloudResourceKind, error) {
	if input == "" {
		return cloudresourcekind.CloudResourceKind_unspecified, fmt.Errorf("cloud resource kind cannot be empty")
	}

	// Try direct enum lookup first (exact match)
	if val, ok := cloudresourcekind.CloudResourceKind_value[input]; ok {
		return cloudresourcekind.CloudResourceKind(val), nil
	}

	// Normalize: lowercase and replace spaces/hyphens with underscores
	normalized := strings.ToLower(input)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	// Try normalized lookup
	if val, ok := cloudresourcekind.CloudResourceKind_value[normalized]; ok {
		return cloudresourcekind.CloudResourceKind(val), nil
	}

	// Try converting snake_case to PascalCase for enum lookup
	// This handles cases where list_cloud_resource_kinds returns "aws_rds_instance"
	// and we need to look up "AwsRdsInstance" in the enum
	pascalCase := snakeToPascalCase(normalized)
	if val, ok := cloudresourcekind.CloudResourceKind_value[pascalCase]; ok {
		return cloudresourcekind.CloudResourceKind(val), nil
	}

	// If still not found, return error with suggestions
	suggestions := FindSimilarKinds(input, 5)
	if len(suggestions) > 0 {
		return cloudresourcekind.CloudResourceKind_unspecified,
			fmt.Errorf("unknown cloud resource kind: %s. Did you mean: %s?", input, strings.Join(suggestions, ", "))
	}

	return cloudresourcekind.CloudResourceKind_unspecified,
		fmt.Errorf("unknown cloud resource kind: %s", input)
}

// FindSimilarKinds finds CloudResourceKind enum values similar to the input string.
// Uses simple string distance/matching heuristics to find likely matches.
//
// Args:
//   - input: The input string to match against
//   - limit: Maximum number of suggestions to return
//
// Returns a slice of similar kind names (enum string values).
func FindSimilarKinds(input string, limit int) []string {
	if input == "" || limit <= 0 {
		return nil
	}

	normalized := strings.ToLower(input)
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.ReplaceAll(normalized, "-", "_")

	var matches []kindMatch

	// Get all CloudResourceKind enum values
	for name := range cloudresourcekind.CloudResourceKind_value {
		// Skip unspecified
		if name == "unspecified" {
			continue
		}

		score := calculateSimilarity(normalized, name)
		if score > 0 {
			matches = append(matches, kindMatch{
				name:  name,
				score: score,
			})
		}
	}

	// Sort by score (descending)
	sortMatchesByScore(matches)

	// Return top N matches
	result := make([]string, 0, limit)
	for i := 0; i < len(matches) && i < limit; i++ {
		result = append(result, matches[i].name)
	}

	return result
}

// kindMatch represents a kind name with its similarity score
type kindMatch struct {
	name  string
	score int
}

// calculateSimilarity returns a simple similarity score between two strings.
// Higher score means more similar.
func calculateSimilarity(input, candidate string) int {
	score := 0

	// Exact match
	if input == candidate {
		return 1000
	}

	// Contains match
	if strings.Contains(candidate, input) {
		score += 500
	}
	if strings.Contains(input, candidate) {
		score += 400
	}

	// Prefix match
	if strings.HasPrefix(candidate, input) {
		score += 300
	}
	if strings.HasPrefix(input, candidate) {
		score += 250
	}

	// Word-level matching (split by underscore)
	inputWords := strings.Split(input, "_")
	candidateWords := strings.Split(candidate, "_")

	for _, inWord := range inputWords {
		for _, candWord := range candidateWords {
			if inWord == candWord {
				score += 100
			} else if strings.Contains(candWord, inWord) {
				score += 50
			} else if strings.Contains(inWord, candWord) {
				score += 25
			}
		}
	}

	return score
}

// sortMatchesByScore sorts kindMatch slice by score in descending order (simple bubble sort)
func sortMatchesByScore(matches []kindMatch) {
	n := len(matches)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if matches[j].score < matches[j+1].score {
				matches[j], matches[j+1] = matches[j+1], matches[j]
			}
		}
	}
}

// GetPopularKindsByCategory returns a map of popular CloudResourceKind values grouped by provider/category.
// This is useful for error messages and agent guidance.
func GetPopularKindsByCategory() map[string][]string {
	return map[string][]string{
		"kubernetes": {
			"kubernetes_deployment",
			"kubernetes_postgres",
			"kubernetes_redis",
			"kubernetes_mongodb",
			"kubernetes_kafka",
		},
		"aws": {
			"aws_eks_cluster",
			"aws_rds_instance",
			"aws_rds_cluster",
			"aws_lambda",
			"aws_s3_bucket",
			"aws_vpc",
		},
		"gcp": {
			"gcp_gke_cluster",
			"gcp_cloud_sql",
			"gcp_cloud_function",
			"gcp_vpc",
		},
		"azure": {
			"azure_aks_cluster",
			"azure_postgres",
			"azure_storage_account",
		},
	}
}

// snakeToPascalCase converts snake_case to PascalCase
// Examples: "aws_rds_instance" → "AwsRdsInstance", "gcp_gke_cluster" → "GcpGkeCluster"
func snakeToPascalCase(s string) string {
	var result strings.Builder
	capitalizeNext := true
	for _, r := range s {
		if r == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result.WriteRune(unicode.ToUpper(r))
			capitalizeNext = false
		} else {
			result.WriteRune(unicode.ToLower(r))
		}
	}
	return result.String()
}







