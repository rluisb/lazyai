// Package compiler provides fragment resolution for template compilation.
// Ported from the TypeScript fragment-resolver.ts.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// FragmentContext holds template variables for fragment resolution.
type FragmentContext struct {
	ProjectName         string
	PlanningDir         string
	PrimaryLanguage     string
	Framework           string
	WorkspaceType       string
	ProjectInstructions string
	ToolDescription     string
	ToolNotes           string
	TestFramework       string
	PackageManager      string
	TestCommand         string
	LintCommand         string
	BuildCommand        string
	DevCommand          string
	InstallCommand      string
	ProjectDescription  string
	Features            *FeatureFlags
}

// FeatureFlags controls which features are compiled into the output.
// Supports both camelCase and snake_case for backwards compatibility.
type FeatureFlags struct {
	ContextEngineering *bool
	RPIWorkflow        *bool
	ChainOfThought     *bool
	TreeOfThoughts     *bool
	ADREnforcement     *bool
	QualityGates       *bool
	AgentHarness       *bool
	BugResolution      *bool
	PivotHandling      *bool
	GitConventions     *bool

	// Legacy snake_case aliases
	ContextEngineering_ *bool
	RPIWorkflow_        *bool
	ChainOfThought_     *bool
	TreeOfThoughts_     *bool
	ADREnforcement_     *bool
	QualityGates_       *bool
	AgentHarness_       *bool
	BugResolution_      *bool
	PivotHandling_      *bool
	GitConventions_     *bool
}

// FragmentResolver resolves XML fragments and variable interpolation in templates.
type FragmentResolver struct {
	libraryDir    string
	fragmentCache map[string]string
}

// NewFragmentResolver creates a new FragmentResolver for the given library directory.
func NewFragmentResolver(libraryDir string) *FragmentResolver {
	return &FragmentResolver{
		libraryDir:    libraryDir,
		fragmentCache: make(map[string]string),
	}
}

// Resolve resolves all fragments, conditionals, and variables in the template content.
func (r *FragmentResolver) Resolve(content string, ctx FragmentContext) string {
	// First pass: resolve conditionals.
	result := r.resolveConditionals(content, ctx)
	// Second pass: resolve includes.
	result = r.resolveIncludes(result)
	// Third pass: resolve variables.
	result = r.resolveVariables(result, ctx)
	return result
}

// resolveConditionals handles {{#if features.name}}...{{/if}} blocks.
var conditionalRe = regexp.MustCompile(`\{\{#if\s+([\w.]+)\}\}([\s\S]*?)\{\{\/if\}\}`)

func (r *FragmentResolver) resolveConditionals(content string, ctx FragmentContext) string {
	return conditionalRe.ReplaceAllStringFunc(content, func(match string) string {
		submatch := conditionalRe.FindStringSubmatch(match)
		if len(submatch) < 3 {
			return ""
		}
		condition := submatch[1]
		body := submatch[2]
		if r.evaluateCondition(condition, ctx) {
			return body
		}
		return ""
	})
}

func (r *FragmentResolver) evaluateCondition(condition string, ctx FragmentContext) bool {
	parts := strings.Split(condition, ".")
	if len(parts) == 0 {
		return false
	}

	if parts[0] == "features" && ctx.Features != nil {
		if len(parts) < 2 {
			return false
		}
		name := parts[1]
		return r.getFeatureFlag(name, ctx.Features)
	}

	return false
}

func (r *FragmentResolver) getFeatureFlag(name string, f *FeatureFlags) bool {
	switch name {
	case "contextEngineering", "context_engineering":
		return boolPtr(f.ContextEngineering) || boolPtr(f.ContextEngineering_)
	case "rpiWorkflow", "rpi_workflow":
		return boolPtr(f.RPIWorkflow) || boolPtr(f.RPIWorkflow_)
	case "chainOfThought", "chain_of_thought":
		return boolPtr(f.ChainOfThought) || boolPtr(f.ChainOfThought_)
	case "treeOfThoughts", "tree_of_thoughts":
		return boolPtr(f.TreeOfThoughts) || boolPtr(f.TreeOfThoughts_)
	case "adrEnforcement", "adr_enforcement":
		return boolPtr(f.ADREnforcement) || boolPtr(f.ADREnforcement_)
	case "qualityGates", "quality_gates":
		return boolPtr(f.QualityGates) || boolPtr(f.QualityGates_)
	case "agentHarness", "agent_harness":
		return boolPtr(f.AgentHarness) || boolPtr(f.AgentHarness_)
	case "bugResolution", "bug_resolution":
		return boolPtr(f.BugResolution) || boolPtr(f.BugResolution_)
	case "pivotHandling", "pivot_handling":
		return boolPtr(f.PivotHandling) || boolPtr(f.PivotHandling_)
	case "gitConventions", "git_conventions":
		return boolPtr(f.GitConventions) || boolPtr(f.GitConventions_)
	default:
		return false
	}
}

func boolPtr(p *bool) bool {
	return p != nil && *p
}

// resolveIncludes handles {{#include fragments/name.xml}} directives.
var includeRe = regexp.MustCompile(`\{\{#include\s+([\w/.-]+)\}\}`)

func (r *FragmentResolver) resolveIncludes(content string) string {
	return includeRe.ReplaceAllStringFunc(content, func(match string) string {
		submatch := includeRe.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		return r.loadFragment(submatch[1])
	})
}

func (r *FragmentResolver) loadFragment(fragmentPath string) string {
	if cached, ok := r.fragmentCache[fragmentPath]; ok {
		return cached
	}

	fullPath := filepath.Join(r.libraryDir, fragmentPath)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fragment not found: %s\n", fragmentPath)
		return fmt.Sprintf("<!-- Fragment not found: %s -->", fragmentPath)
	}

	content := string(data)
	r.fragmentCache[fragmentPath] = content
	return content
}

// resolveVariables handles {{VARIABLE_NAME}} substitution.
var variableRe = regexp.MustCompile(`\{\{(\w+)\}\}`)

func (r *FragmentResolver) resolveVariables(content string, ctx FragmentContext) string {
	variables := map[string]string{
		"PROJECT_NAME":         ctx.ProjectName,
		"PLANNING_DIR":         ctx.PlanningDir,
		"PRIMARY_LANGUAGE":     defaultStr(ctx.PrimaryLanguage, "TypeScript"),
		"FRAMEWORK":            ctx.Framework,
		"WORKSPACE_TYPE":       defaultStr(ctx.WorkspaceType, "project"),
		"TOOL_DESCRIPTION":     ctx.ToolDescription,
		"TOOL_NOTES":           ctx.ToolNotes,
		"PROJECT_INSTRUCTIONS": ctx.ProjectInstructions,
		"TEST_FRAMEWORK":       ctx.TestFramework,
		"PACKAGE_MANAGER":      ctx.PackageManager,
		"TEST_COMMAND":         ctx.TestCommand,
		"LINT_COMMAND":         ctx.LintCommand,
		"BUILD_COMMAND":        ctx.BuildCommand,
		"DEV_COMMAND":          ctx.DevCommand,
		"INSTALL_COMMAND":      ctx.InstallCommand,
		"PROJECT_DESCRIPTION":  ctx.ProjectDescription,
	}

	return variableRe.ReplaceAllStringFunc(content, func(match string) string {
		submatch := variableRe.FindStringSubmatch(match)
		if len(submatch) < 2 {
			return match
		}
		varName := submatch[1]
		if val, ok := variables[varName]; ok && val != "" {
			return val
		}
		return match // Leave unresolved variables as-is.
	})
}

func defaultStr(val, def string) string {
	if val != "" {
		return val
	}
	return def
}
