// Package compiler provides the template compiler that generates tool-native
// output files from shared templates.
// Ported from the TypeScript template-compiler.ts.
package compiler

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SharedRootTemplatePath is the path to the shared root template relative to the library dir.
var SharedRootTemplatePath = path.Join("tool-templates", "shared", "root.template.md")

// CompilerConfig holds the configuration for template compilation.
type CompilerConfig struct {
	LibraryDir string
	LibraryFS  fs.FS // Optional: when set, takes precedence over LibraryDir for reads
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
		resolver: NewFragmentResolver(config.LibraryDir, config.LibraryFS),
	}
}

// Compile compiles all templates for the configured tool and returns the output.
func (tc *TemplateCompiler) Compile() (*CompiledOutput, error) {
	toolTemplateDir := path.Join("tool-templates", tc.config.Tool)
	sharedRootTemplate := SharedRootTemplatePath

	hasSharedRootTemplate := tc.fileExistsFS(sharedRootTemplate)
	hasToolTemplateDir := tc.dirExistsFS(toolTemplateDir)

	if !hasSharedRootTemplate && !hasToolTemplateDir {
		return nil, fmt.Errorf("tool template directory not found: %s", toolTemplateDir)
	}

	var files []CompiledFile
	context := tc.getContextWithToolOverrides()

	// Process shared root template.
	if hasSharedRootTemplate {
		content, err := tc.readFileFS(sharedRootTemplate)
		if err != nil {
			return nil, fmt.Errorf("read shared root template: %w", err)
		}
		resolved := tc.resolver.Resolve(string(content), context)
		files = append(files, CompiledFile{RelativePath: "root.md", Content: resolved})
	}

	// Process tool-specific templates.
	if hasToolTemplateDir {
		templateFiles := tc.findTemplateFilesFS(toolTemplateDir)
		for _, templateFile := range templateFiles {
			// Skip root template if shared root template was already processed.
			baseName := path.Base(templateFile)
			if hasSharedRootTemplate && strings.HasPrefix(baseName, "root.template.") {
				continue
			}

			content, err := tc.readFileFS(templateFile)
			if err != nil {
				return nil, fmt.Errorf("read template %s: %w", templateFile, err)
			}
			resolved := tc.resolver.Resolve(string(content), context)

			// Convert template path to output path.
			relPath := templateFile
			if tc.config.LibraryFS != nil {
				relPath = strings.TrimPrefix(templateFile, toolTemplateDir+"/")
			} else if diskRel, err := filepath.Rel(filepath.Join(tc.config.LibraryDir, toolTemplateDir), templateFile); err == nil {
				relPath = diskRel
			} else {
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
		compilerLog.Info("wrote compiled file", "path", file.RelativePath)
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
func CompileForTools(tools []string, libraryDir string, libFS fs.FS, outputDir string, context FragmentContext) (map[string]*CompiledOutput, error) {
	results := make(map[string]*CompiledOutput)

	for _, tool := range tools {
		compiler := NewTemplateCompiler(CompilerConfig{
			LibraryDir: libraryDir,
			LibraryFS:  libFS,
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

// --- fs.FS helper methods ---

// fileExistsFS checks if a file exists in the library FS (or disk).
func (tc *TemplateCompiler) fileExistsFS(path string) bool {
	if tc.config.LibraryFS != nil {
		_, err := fs.Stat(tc.config.LibraryFS, path)
		return err == nil
	}
	fullPath := filepath.Join(tc.config.LibraryDir, path)
	_, err := os.Stat(fullPath)
	return err == nil
}

// dirExistsFS checks if a directory exists in the library FS (or disk).
func (tc *TemplateCompiler) dirExistsFS(path string) bool {
	if tc.config.LibraryFS != nil {
		info, err := fs.Stat(tc.config.LibraryFS, path)
		return err == nil && info.IsDir()
	}
	fullPath := filepath.Join(tc.config.LibraryDir, path)
	info, err := os.Stat(fullPath)
	return err == nil && info.IsDir()
}

// readFileFS reads a file from the library FS (or disk).
func (tc *TemplateCompiler) readFileFS(path string) ([]byte, error) {
	if tc.config.LibraryFS != nil {
		return fs.ReadFile(tc.config.LibraryFS, path)
	}
	fullPath := filepath.Join(tc.config.LibraryDir, path)
	return os.ReadFile(fullPath)
}

// findTemplateFilesFS recursively finds all template files in a directory.
func (tc *TemplateCompiler) findTemplateFilesFS(dir string) []string {
	if tc.config.LibraryFS != nil {
		return findTemplateFilesInFS(tc.config.LibraryFS, dir)
	}
	return findTemplateFiles(filepath.Join(tc.config.LibraryDir, dir))
}

// findTemplateFilesInFS recursively finds template files in an fs.FS.
func findTemplateFilesInFS(fsys fs.FS, dir string) []string {
	var result []string
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		fullPath := dir + "/" + entry.Name()
		if entry.IsDir() {
			result = append(result, findTemplateFilesInFS(fsys, fullPath)...)
		} else if containsTemplate(entry.Name()) {
			result = append(result, fullPath)
		}
	}
	return result
}

// findTemplateFiles recursively finds all files containing ".template." in their name on disk.
func findTemplateFiles(dir string) []string {
	var result []string
	entries, err := os.ReadDir(dir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dir, entry.Name())
		if entry.IsDir() {
			result = append(result, findTemplateFiles(fullPath)...)
		} else if containsTemplate(entry.Name()) {
			result = append(result, fullPath)
		}
	}
	return result
}

func containsTemplate(name string) bool {
	return filepath.Ext(name) != "" && len(name) > 10 &&
		(name[len(name)-len(".template.md"):] == ".template.md" ||
			containsDotTemplate(name))
}

func containsDotTemplate(name string) bool {
	for i := 0; i < len(name)-len(".template."); i++ {
		if name[i:i+len(".template.")] == ".template." {
			return true
		}
	}
	return false
}
