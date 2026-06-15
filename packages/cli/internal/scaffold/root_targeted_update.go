package scaffold

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/compiler"
)

// TargetedUpdatePatch is the audit contract emitted by targeted AGENTS.md
// updates. It records only the slots that were safely patched and any slots
// skipped to preserve hand-authored content.
type TargetedUpdatePatch struct {
	File                         string                      `json:"file"`
	Replacements                 []TargetedUpdateReplacement `json:"replacements"`
	Warnings                     []string                    `json:"warnings"`
	PreservedUnrecognizedContent bool                        `json:"preservedUnrecognizedContent"`
}

type TargetedUpdateReplacement struct {
	Field    string                 `json:"field"`
	OldText  string                 `json:"oldText"`
	NewText  string                 `json:"newText"`
	Location TargetedUpdateLocation `json:"location"`
}

type TargetedUpdateLocation struct {
	Section   *string `json:"section"`
	LineStart *int    `json:"lineStart"`
	LineEnd   *int    `json:"lineEnd"`
}

type targetedFieldSpec struct {
	field        string
	newText      string
	section      string
	placeholders []string
	linePrefixes []string
}

// BuildTargetedAgentsUpdatePatch applies the W1.A targeted update policy for an
// existing AGENTS.md: exact fallback placeholders are replaced, simple generated
// value slots are patched only when still safely recognizable, and all other
// content is preserved byte-for-byte with warnings for unsafe known slots.
func BuildTargetedAgentsUpdatePatch(file, existing string, ctx compiler.FragmentContext) (string, TargetedUpdatePatch) {
	patch := TargetedUpdatePatch{
		File:                         file,
		Replacements:                 []TargetedUpdateReplacement{},
		Warnings:                     []string{},
		PreservedUnrecognizedContent: true,
	}
	content := existing

	for _, spec := range targetedAgentsFieldSpecs(ctx) {
		if spec.newText == "" {
			continue
		}
		for _, placeholder := range spec.placeholders {
			content = replaceTargetedExact(content, placeholder, spec, &patch)
		}
	}

	for _, spec := range targetedAgentsFieldSpecs(ctx) {
		if spec.newText == "" || len(spec.linePrefixes) == 0 {
			continue
		}
		content = replaceTargetedLineSlots(content, spec, &patch)
	}
	warnUnsafeProjectOverview(content, ctx, &patch)

	return content, patch
}

func targetedAgentsFieldSpecs(ctx compiler.FragmentContext) []targetedFieldSpec {
	c := ctx.Constitution
	if c == nil {
		c = &compiler.ConstitutionContext{}
	}
	coverage := ""
	if c.CoverageThreshold != nil {
		coverage = strconv.Itoa(*c.CoverageThreshold)
	}
	return []targetedFieldSpec{
		{field: "PROJECT_OVERVIEW", newText: strings.TrimSpace(c.ProjectOverview), section: "Project Overview", placeholders: []string{"[YOUR_PROJECT_OVERVIEW]"}},
		{field: "LANGUAGE", newText: strings.TrimSpace(c.Stack.Language), section: "Project Overview", placeholders: []string{"[YOUR_LANGUAGE]"}, linePrefixes: []string{"- Language: "}},
		{field: "FRAMEWORK", newText: strings.TrimSpace(c.Stack.Framework), section: "Project Overview", placeholders: []string{"[YOUR_FRAMEWORK]"}, linePrefixes: []string{"- Framework: "}},
		{field: "DATABASE", newText: strings.TrimSpace(c.Stack.Database), section: "Project Overview", placeholders: []string{"[YOUR_DATABASE]"}, linePrefixes: []string{"- Database: "}},
		{field: "ORM", newText: strings.TrimSpace(c.Stack.ORM), section: "Project Overview", placeholders: []string{"[YOUR_ORM]"}, linePrefixes: []string{"- ORM/Query: "}},
		{field: "TEST_FRAMEWORK", newText: strings.TrimSpace(c.Stack.Testing), section: "Project Overview", placeholders: []string{"[YOUR_TEST_FRAMEWORK]"}, linePrefixes: []string{"- Testing: "}},
		{field: "PACKAGE_MANAGER", newText: strings.TrimSpace(c.Stack.PackageManager), section: "Project Overview", placeholders: []string{"[YOUR_PACKAGE_MANAGER]"}, linePrefixes: []string{"- Package manager: "}},
		{field: "NAMING_CONVENTIONS", newText: strings.TrimSpace(c.Conventions.Naming), section: "Conventions", placeholders: []string{"[YOUR_NAMING_CONVENTION]"}},
		{field: "ERROR_HANDLING", newText: strings.TrimSpace(c.Conventions.ErrorHandling), section: "Conventions", placeholders: []string{"[YOUR_ERROR_PATTERN]"}},
		{field: "API_CONVENTIONS", newText: strings.TrimSpace(c.Conventions.APIResponses), section: "Conventions", placeholders: []string{"[YOUR_API_CONVENTION]"}},
		{field: "IMPORT_ORDER", newText: strings.TrimSpace(c.Conventions.ImportOrder), section: "Conventions", placeholders: []string{"[YOUR_IMPORT_ORDER]"}},
		{field: "PROTECTED_BRANCH", newText: strings.TrimSpace(c.ProtectedBranch), section: "Do NOT", placeholders: []string{"[YOUR_PROTECTED_BRANCH]"}},
		{field: "TEST_COMMAND", newText: strings.TrimSpace(c.Commands.Test), section: "Key Commands", placeholders: []string{"<!-- fill-in: test command -->"}},
		{field: "LINT_COMMAND", newText: strings.TrimSpace(c.Commands.Lint), section: "Key Commands", placeholders: []string{"[YOUR_LINT_COMMAND]"}},
		{field: "BUILD_COMMAND", newText: strings.TrimSpace(c.Commands.Build), section: "Key Commands", placeholders: []string{"<!-- fill-in: build command -->"}},
		{field: "COVERAGE_THRESHOLD", newText: coverage, section: "Testing", placeholders: []string{"[YOUR_COVERAGE_THRESHOLD]"}, linePrefixes: []string{"- Minimum coverage: ", "- Minimum coverage threshold: "}},
	}
}

func replaceTargetedExact(content, oldText string, spec targetedFieldSpec, patch *TargetedUpdatePatch) string {
	if oldText == "" || spec.newText == "" || oldText == spec.newText {
		return content
	}
	searchStart := 0
	for searchStart <= len(content) {
		relativeIdx := strings.Index(content[searchStart:], oldText)
		if relativeIdx < 0 {
			return content
		}
		idx := searchStart + relativeIdx
		line := lineNumberAt(content, idx)
		patch.Replacements = append(patch.Replacements, TargetedUpdateReplacement{
			Field:    spec.field,
			OldText:  oldText,
			NewText:  spec.newText,
			Location: targetedLocation(spec.section, line, line),
		})
		content = content[:idx] + spec.newText + content[idx+len(oldText):]
		searchStart = idx + len(spec.newText)
	}
	return content
}

func replaceTargetedLineSlots(content string, spec targetedFieldSpec, patch *TargetedUpdatePatch) string {
	lines := strings.SplitAfter(content, "\n")
	changed := false
	for idx, line := range lines {
		body, ending := splitLineEnding(line)
		for _, prefix := range spec.linePrefixes {
			if !strings.HasPrefix(body, prefix) {
				continue
			}
			oldValue := strings.TrimSpace(strings.TrimPrefix(body, prefix))
			if normalizeSlotValue(oldValue) == spec.newText {
				continue
			}
			if !isSafeTargetedSlot(oldValue, spec) {
				patch.Warnings = append(patch.Warnings, fmt.Sprintf("left %s unchanged at line %d because existing value is not a recognized placeholder/value slot", spec.field, idx+1))
				continue
			}
			newBody := prefix + preserveSlotDelimiters(oldValue, spec.newText)
			patch.Replacements = append(patch.Replacements, TargetedUpdateReplacement{
				Field:    spec.field,
				OldText:  oldValue,
				NewText:  spec.newText,
				Location: targetedLocation(spec.section, idx+1, idx+1),
			})
			lines[idx] = newBody + ending
			changed = true
		}
	}
	if !changed {
		return content
	}
	return strings.Join(lines, "")
}

func warnUnsafeProjectOverview(content string, ctx compiler.FragmentContext, patch *TargetedUpdatePatch) {
	if ctx.Constitution == nil || strings.TrimSpace(ctx.Constitution.ProjectOverview) == "" {
		return
	}
	if strings.Contains(content, ctx.Constitution.ProjectOverview) || strings.Contains(content, "[YOUR_PROJECT_OVERVIEW]") {
		return
	}
	lines := strings.Split(content, "\n")
	for idx, line := range lines {
		if strings.TrimSpace(line) != "## Project Overview" {
			continue
		}
		for next := idx + 1; next < len(lines); next++ {
			trimmed := strings.TrimSpace(lines[next])
			if trimmed == "" || strings.HasPrefix(trimmed, "<!--") {
				continue
			}
			if strings.HasPrefix(trimmed, "## ") || strings.HasPrefix(trimmed, "**Stack:**") {
				return
			}
			patch.Warnings = append(patch.Warnings, fmt.Sprintf("left PROJECT_OVERVIEW unchanged at line %d because existing value is not a recognized placeholder/value slot", next+1))
			return
		}
	}
}

func splitLineEnding(line string) (body string, ending string) {
	if strings.HasSuffix(line, "\n") {
		ending = "\n"
		body = strings.TrimSuffix(line, "\n")
		if strings.HasSuffix(body, "\r") {
			body = strings.TrimSuffix(body, "\r")
			ending = "\r\n"
		}
		return body, ending
	}
	return line, ""
}

func isSafeTargetedSlot(oldValue string, spec targetedFieldSpec) bool {
	normalized := normalizeSlotValue(oldValue)
	if normalized == "" || strings.Contains(normalized, "[YOUR_") || strings.Contains(normalized, "{{") || strings.Contains(normalized, "fill-in:") {
		return true
	}
	for _, placeholder := range spec.placeholders {
		if normalized == placeholder {
			return true
		}
	}
	return spec.field == "COVERAGE_THRESHOLD" && normalized == "80"
}

func normalizeSlotValue(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, "%")
	trimmed = strings.TrimPrefix(trimmed, "`")
	trimmed = strings.TrimSuffix(trimmed, "`")
	return strings.TrimSpace(trimmed)
}

func preserveSlotDelimiters(oldValue, newText string) string {
	trimmed := strings.TrimSpace(oldValue)
	if strings.HasPrefix(trimmed, "`") && strings.HasSuffix(trimmed, "`") {
		return "`" + newText + "`"
	}
	return newText
}

func targetedLocation(section string, lineStart, lineEnd int) TargetedUpdateLocation {
	sectionCopy := section
	return TargetedUpdateLocation{Section: &sectionCopy, LineStart: &lineStart, LineEnd: &lineEnd}
}

func lineNumberAt(content string, idx int) int {
	if idx <= 0 {
		return 1
	}
	return strings.Count(content[:idx], "\n") + 1
}
