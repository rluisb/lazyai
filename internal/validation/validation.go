// Package validation provides name validation utilities for the ai-setup project.
// Ported from the TypeScript utilities in src/utils/validation.ts.
package validation

import (
	"fmt"
	"regexp"
	"strings"

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
