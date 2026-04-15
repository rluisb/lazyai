// Package compiler provides the template compiler that generates tool-native
// output files from shared templates.
// Ported from the TypeScript template-compiler.ts.
package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CompilerConfig holds the configuration for template compilation.
type CompilerConfig struct {
	LibraryDir string
	OutputDir  string
	Tool       string
	Context    FragmentContext
}

// TemplateCompiler compiles templates for a specific tool by resolving
// fragments and generating tool-native output files.
type TemplateCompiler struct {
	config   CompilerConfig
	resolver *FragmentResolver
}

// NewTemplateCompiler creates a new TemplateCompiler for the given configuration.
func NewTemplateCompiler(config CompilerConfig) *TemplateCompiler {
	return &TemplateCompiler{
		config:   config,
		resolver: NewFragmentResolver(config.LibraryDir),
	}
}

// Compile compiles all templates for the configured tool and returns the output.
func (tc *TemplateCompiler) Compile() (*CompiledOutput, error) {
	toolTemplateDir := filepath.Join(tc.config.LibraryDir, "tool-templates", tc.config.Tool)
	sharedRootTemplate := filepath.Join(tc.config.LibraryDir, SharedRootTemplatePath)

	hasSharedRootTemplate := fileExists(sharedRootTemplate)
	hasToolTemplateDir := dirExists(toolTemplateDir)

	if !hasSharedRootTemplate && !hasToolTemplateDir {
		return nil, fmt.Errorf("tool template directory not found: %s", toolTemplateDir)
	}

	var files []CompiledFile
	context := tc.getContextWithToolOverrides()

	// Process shared root template.
	if hasSharedRootTemplate {
		content, err := os.ReadFile(sharedRootTemplate)
		if err != nil {
			return nil, fmt.Errorf("read shared root template: %w", err)
		}
		resolved := tc.resolver.Resolve(string(content), context)
		files = append(files, CompiledFile{RelativePath: "root.md", Content: resolved})
	}

	// Process tool-specific templates.
	if hasToolTemplateDir {
		templateFiles := findTemplateFiles(toolTemplateDir)
		for _, templateFile := range templateFiles {
			// Skip root template if shared root template was already processed.
			baseName := filepath.Base(templateFile)
			if hasSharedRootTemplate && strings.HasPrefix(baseName, "root.template.") {
				continue
			}

			content, err := os.ReadFile(templateFile)
			if err != nil {
				return nil, fmt.Errorf("read template %s: %w", templateFile, err)
			}
			resolved := tc.resolver.Resolve(string(content), context)

			// Convert template path to output path.
			relPath, err := filepath.Rel(toolTemplateDir, templateFile)
			if err != nil {
				relPath = filepath.Base(templateFile)
			}
			outputPath := strings.ReplaceAll(relPath, ".template.md", ".md")
			outputPath = strings.ReplaceAll(outputPath, ".template.", ".")

			files = append(files, CompiledFile{RelativePath: outputPath, Content: resolved})
		}
	}

	return &CompiledOutput{Tool: tc.config.Tool, Files: files}, nil
}

// CompileAndWrite compiles templates and writes the output files to disk.
func (tc *TemplateCompiler) CompileAndWrite() error {
	output, err := tc.Compile()
	if err != nil {
		return err
	}

	for _, file := range output.Files {
		fullPath := filepath.Join(tc.config.OutputDir, file.RelativePath)
		dir := filepath.Dir(fullPath)

		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}

		if err := os.WriteFile(fullPath, []byte(file.Content), 0o644); err != nil {
			return fmt.Errorf("write file %s: %w", fullPath, err)
		}
		fmt.Fprintf(os.Stderr, "  -> %s\n", file.RelativePath)
	}
	return nil
}

func (tc *TemplateCompiler) getContextWithToolOverrides() FragmentContext {
	ctx := tc.config.Context
	overrides, ok := ToolOverrideMap[tc.config.Tool]
	if ok {
		ctx.ToolDescription = overrides.Description
		ctx.ToolNotes = overrides.Notes
	}
	return ctx
}

// CompileForTools compiles templates for multiple tools and returns a map of results.
func CompileForTools(tools []string, libraryDir, outputDir string, context FragmentContext) (map[string]*CompiledOutput, error) {
	results := make(map[string]*CompiledOutput)

	for _, tool := range tools {
		compiler := NewTemplateCompiler(CompilerConfig{
			LibraryDir: libraryDir,
			OutputDir:  outputDir,
			Tool:       tool,
			Context:    context,
		})

		output, err := compiler.Compile()
		if err != nil {
			return nil, fmt.Errorf("compile for tool %s: %w", tool, err)
		}
		results[tool] = output
	}

	return results, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
