package migration

import (
	"fmt"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
)

// parseCopilotSetup reads a project's GitHub Copilot configuration
// (.github/copilot-instructions.md + `.github/copilot-chat-modes/` tree +
// related chatmode / instruction files) and returns a canonical ParsedSetup.
// Mirrors TS's `CopilotParser.parse()` in
// packages/ai-setup-ts/src/migration/parsers/copilot-parser.ts.
func parseCopilotSetup(sourcePath string) (*ParsedSetup, error) {
	parsed := &ParsedSetup{
		Agents:         []AgentDefinition{},
		Rules:          []RuleDefinition{},
		Commands:       []CommandDefinition{},
		Templates:      []TemplateDefinition{},
		CustomSections: []CustomSection{},
		Files:          []ParsedFile{},
		Metadata: map[string]string{
			"adapter": "copilot",
		},
	}

	if err := parseCopilotInstructions(sourcePath, parsed); err != nil {
		return nil, err
	}

	// .github/copilot-chat-modes/*.chatmode.md — VS Code chat modes
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".github", "copilot-chat-modes", "*.chatmode.md"), func(id, name, content, rel string) {
		parsed.Agents = append(parsed.Agents, AgentDefinition{
			ID: id, Name: name, Content: content, SourcePath: rel,
		})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "agent"})
	}); err != nil {
		return nil, err
	}

	// .github/instructions/*.instructions.md — per-path Copilot instructions
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".github", "instructions", "*.instructions.md"), func(id, name, content, rel string) {
		parsed.Rules = append(parsed.Rules, RuleDefinition{
			ID:         id,
			Category:   "copilot-instruction",
			Content:    content,
			SourcePath: rel,
			Priority:   50,
		})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "rule"})
	}); err != nil {
		return nil, err
	}

	if parsed.ProjectName == "" {
		parsed.ProjectName = filepath.Base(sourcePath)
	}

	if len(parsed.Agents) == 0 && len(parsed.Rules) == 0 && len(parsed.CustomSections) == 0 {
		return nil, fmt.Errorf("detected Copilot setup did not contain any parseable migration content")
	}

	return parsed, nil
}

func parseCopilotInstructions(sourcePath string, parsed *ParsedSetup) error {
	rootPath := filepath.Join(sourcePath, ".github", "copilot-instructions.md")
	if !files.FileExists(rootPath) {
		return nil
	}

	data, err := files.ReadFile(rootPath)
	if err != nil {
		return fmt.Errorf("parse copilot-instructions.md: %w", err)
	}
	content := string(data)
	rel := filepath.Join(".github", "copilot-instructions.md")
	parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "config"})

	if name := extractFirstHeading(content); name != "" {
		parsed.ProjectName = name
	}
	if description := extractFirstParagraph(content); description != "" {
		parsed.Description = description
	}

	parsed.CustomSections = append(parsed.CustomSections, CustomSection{
		ID:         "copilot-instructions",
		Title:      "Imported copilot-instructions.md",
		Content:    content,
		SourcePath: rel,
	})

	return nil
}
