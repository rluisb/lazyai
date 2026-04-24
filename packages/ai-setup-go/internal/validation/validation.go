// Package validation provides name validation utilities for the ai-setup project.
// Ported from the TypeScript utilities in src/utils/validation.ts.
package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	aierror "github.com/ricardoborges-teachable/ai-setup/internal/error"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// invalidFilesystemChars matches characters that are invalid in filesystem paths
// across Windows, macOS, and Linux (matching the TS regex).
var invalidFilesystemCharsRe = regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F]`)

// artifactNameRe validates artifact names: lowercase, hyphens, no spaces.
var artifactNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*[a-z0-9]$`)

// simpleArtifactNameRe allows single-character names too.
var simpleArtifactNameRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

// ValidateArtifactName validates that name is a valid artifact name:
// lowercase alphanumeric and hyphens, no spaces, at least 2 characters.
func ValidateArtifactName(name string) error {
	if name == "" {
		return aierror.InvalidInput("artifact name cannot be empty", nil)
	}
	if !artifactNameRe.MatchString(name) && !simpleArtifactNameRe.MatchString(name) {
		return aierror.InvalidInput(fmt.Sprintf("invalid artifact name: %q (must be lowercase, hyphens, no spaces)", name), nil)
	}
	if len(name) < 2 {
		return aierror.InvalidInput(fmt.Sprintf("artifact name too short: %q (minimum 2 characters)", name), nil)
	}
	return nil
}

// ValidateToolId validates that id is a recognized tool identifier.
func ValidateToolId(id string) error {
	if !types.IsValidToolId(types.ToolId(id)) {
		return aierror.InvalidInput(fmt.Sprintf("invalid tool id: %q", id), nil)
	}
	return nil
}

// IsValidArtifactType reports whether t is one of the valid artifact type strings.
func IsValidArtifactType(t string) bool {
	switch types.ArtifactType(t) {
	case types.ArtifactTypeAgent,
		types.ArtifactTypeSkill,
		types.ArtifactTypeCommand,
		types.ArtifactTypePrompt,
		types.ArtifactTypeTemplate,
		types.ArtifactTypeWorkflow,
		types.ArtifactTypeDomain,
		types.ArtifactTypeMode:
		return true
	default:
		return false
	}
}

// SanitizeName converts a name to a valid artifact name by lowercasing,
// replacing spaces and underscores with hyphens, and stripping invalid chars.
func SanitizeName(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	s = invalidFilesystemCharsRe.ReplaceAllString(s, "")

	// Strip leading/trailing hyphens.
	s = strings.Trim(s, "-")

	// Collapse consecutive hyphens.
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}

	return s
}

// ValidateRequiredText validates that a string value is non-empty after trimming.
// Returns nil on success, or an AiSetupError on failure.
func ValidateRequiredText(value string, fieldLabel string) error {
	if strings.TrimSpace(value) == "" {
		return aierror.InvalidInput(fmt.Sprintf("%s cannot be empty", fieldLabel), nil)
	}
	return nil
}

// ValidateFilesystemSafeName validates that a name doesn't contain invalid
// filesystem characters.
func ValidateFilesystemSafeName(value string, fieldLabel string) error {
	if err := ValidateRequiredText(value, fieldLabel); err != nil {
		return err
	}
	if invalidFilesystemCharsRe.MatchString(value) {
		return aierror.InvalidInput(fmt.Sprintf("%s contains invalid filesystem characters", fieldLabel), nil)
	}
	return nil
}

// ValidateSpec006Metadata validates Spec 006 frontmatter fields.
// It returns human-readable issues; callers decide whether they are warnings or errors.
func ValidateSpec006Metadata(fm map[string]any) []string {
	issues := []string{}
	required := []string{"schema_version", "artifact_type", "id", "created_at", "updated_at", "created_by", "updated_by"}

	for _, field := range required {
		if isZeroFrontmatterValue(fm[field]) {
			issues = append(issues, fmt.Sprintf("missing required field: %s", field))
		}
	}

	if title, ok := fm["title"]; ok {
		if _, ok := title.(string); !ok {
			issues = append(issues, "field title must be a string")
		}
	}

	createdAt, createdOk := asFrontmatterString(fm["created_at"])
	updatedAt, updatedOk := asFrontmatterString(fm["updated_at"])
	workflowID, workflowIDOk := asFrontmatterString(fm["workflow_id"])
	workflowRunID, workflowRunIDOk := asFrontmatterString(fm["workflow_run_id"])

	var createdTime time.Time
	var updatedTime time.Time
	if createdOk {
		parsed, err := time.Parse(time.RFC3339, createdAt)
		if err != nil {
			issues = append(issues, "field created_at must be an ISO-8601 UTC datetime")
		} else {
			createdTime = parsed
		}
	}
	if updatedOk {
		parsed, err := time.Parse(time.RFC3339, updatedAt)
		if err != nil {
			issues = append(issues, "field updated_at must be an ISO-8601 UTC datetime")
		} else {
			updatedTime = parsed
		}
	}
	if !createdTime.IsZero() && !updatedTime.IsZero() && updatedTime.Before(createdTime) {
		issues = append(issues, "updated_at must be greater than or equal to created_at")
	}

	if workflowRunIDOk && workflowRunID != "" && (!workflowIDOk || workflowID == "") {
		issues = append(issues, "workflow_run_id requires workflow_id")
	}

	if refs, ok := fm["related_document_refs"]; ok {
		if invalid := validateRelativeRefs(refs); len(invalid) > 0 {
			issues = append(issues, invalid...)
		}
	}

	return issues
}

func isZeroFrontmatterValue(value any) bool {
	if value == nil {
		return true
	}
	if s, ok := value.(string); ok {
		return strings.TrimSpace(s) == ""
	}
	return false
}

func asFrontmatterString(value any) (string, bool) {
	s, ok := value.(string)
	if !ok {
		return "", false
	}
	return strings.TrimSpace(s), true
}

func validateRelativeRefs(value any) []string {
	items, ok := value.([]any)
	if !ok {
		if refs, ok := value.([]string); ok {
			items = make([]any, 0, len(refs))
			for _, ref := range refs {
				items = append(items, ref)
			}
		} else {
			return []string{"field related_document_refs must be an array of repo-relative paths"}
		}
	}

	issues := []string{}
	for _, item := range items {
		ref, ok := item.(string)
		if !ok || strings.TrimSpace(ref) == "" {
			issues = append(issues, "field related_document_refs must contain only strings")
			continue
		}
		if strings.HasPrefix(ref, "/") || strings.Contains(ref, `:\\`) {
			issues = append(issues, fmt.Sprintf("related_document_refs must be repo-relative: %s", ref))
		}
	}
	return issues
}
