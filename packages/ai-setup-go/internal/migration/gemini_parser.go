package migration

import (
	"fmt"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
)

// parseGeminiSetup reads a project's Gemini CLI configuration (GEMINI.md
// + .gemini/ tree) and returns a canonical ParsedSetup. Mirrors TS's
// `GeminiParser.parse()` in packages/ai-setup-ts/src/migration/parsers/gemini-parser.ts.
func parseGeminiSetup(sourcePath string) (*ParsedSetup, error) {
	parsed := &ParsedSetup{
		Agents:         []AgentDefinition{},
		Rules:          []RuleDefinition{},
		Commands:       []CommandDefinition{},
		Templates:      []TemplateDefinition{},
		CustomSections: []CustomSection{},
		Files:          []ParsedFile{},
		Metadata: map[string]string{
			"adapter": "gemini",
		},
	}

	if err := parseGeminiRoot(sourcePath, parsed); err != nil {
		return nil, err
	}

	// .gemini/commands/*.toml — Gemini custom slash commands (TOML)
	if err := parseGeminiCommandsDir(sourcePath, parsed); err != nil {
		return nil, err
	}

	// .gemini/*.md — any extra markdown context files
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".gemini", "*.md"), func(id, name, content, rel string) {
		parsed.CustomSections = append(parsed.CustomSections, CustomSection{
			ID:         id,
			Title:      name,
			Content:    content,
			SourcePath: rel,
		})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "config"})
	}); err != nil {
		return nil, err
	}

	if parsed.ProjectName == "" {
		parsed.ProjectName = filepath.Base(sourcePath)
	}

	if len(parsed.Commands) == 0 && len(parsed.CustomSections) == 0 {
		return nil, fmt.Errorf("detected Gemini setup did not contain any parseable migration content")
	}

	return parsed, nil
}

func parseGeminiRoot(sourcePath string, parsed *ParsedSetup) error {
	rootPath := filepath.Join(sourcePath, "GEMINI.md")
	if !files.FileExists(rootPath) {
		return nil
	}

	data, err := files.ReadFile(rootPath)
	if err != nil {
		return fmt.Errorf("parse GEMINI.md: %w", err)
	}
	content := string(data)
	parsed.Files = append(parsed.Files, ParsedFile{Path: "GEMINI.md", Content: content, Type: "config"})

	if name := extractFirstHeading(content); name != "" {
		parsed.ProjectName = name
	}
	if description := extractFirstParagraph(content); description != "" {
		parsed.Description = description
	}

	parsed.CustomSections = append(parsed.CustomSections, CustomSection{
		ID:         "gemini-root",
		Title:      "Imported GEMINI.md",
		Content:    content,
		SourcePath: "GEMINI.md",
	})

	return nil
}

// parseGeminiCommandsDir reads `.gemini/commands/*.toml` — Gemini custom slash
// command templates are TOML, not Markdown.
func parseGeminiCommandsDir(sourcePath string, parsed *ParsedSetup) error {
	matches, err := filepath.Glob(filepath.Join(sourcePath, ".gemini", "commands", "*.toml"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		data, err := files.ReadFile(match)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filepath.Base(match), err)
		}

		rel, relErr := filepath.Rel(sourcePath, match)
		if relErr != nil {
			rel = match
		}

		base := filepath.Base(match)
		id := base[:len(base)-len(filepath.Ext(base))]
		content := string(data)

		parsed.Commands = append(parsed.Commands, CommandDefinition{
			ID:         id,
			Name:       id,
			Content:    content,
			SourcePath: rel,
		})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "command"})
	}

	return nil
}
