package migration

import (
	"fmt"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
)

// parseClaudeSetup reads a project's Claude Code configuration (CLAUDE.md
// + .claude/ tree) and returns a canonical ParsedSetup. Mirrors TS's
// `ClaudeCodeParser.parse()` in packages/ai-setup-ts/src/migration/parsers/claude-parser.ts.
func parseClaudeSetup(sourcePath string) (*ParsedSetup, error) {
	parsed := &ParsedSetup{
		Agents:         []AgentDefinition{},
		Rules:          []RuleDefinition{},
		Commands:       []CommandDefinition{},
		Templates:      []TemplateDefinition{},
		CustomSections: []CustomSection{},
		Files:          []ParsedFile{},
		Metadata: map[string]string{
			"adapter": "claude-code",
		},
	}

	if err := parseClaudeRoot(sourcePath, parsed); err != nil {
		return nil, err
	}

	// .claude/*.md — agent definitions at the root of .claude/
	if err := parseClaudeRootAgents(sourcePath, parsed); err != nil {
		return nil, err
	}

	// .claude/commands/*.md — user-defined commands
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".claude", "commands", "*.md"), func(id, name, content, rel string) {
		parsed.Commands = append(parsed.Commands, CommandDefinition{ID: id, Name: name, Content: content, SourcePath: rel})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "command"})
	}); err != nil {
		return nil, err
	}

	// .claude/rules/*.md — guardrails
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".claude", "rules", "*.md"), func(id, name, content, rel string) {
		parsed.Rules = append(parsed.Rules, RuleDefinition{ID: id, Category: "claude-rule", Content: content, SourcePath: rel, Priority: 50})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "rule"})
	}); err != nil {
		return nil, err
	}

	// .claude/templates/*.md
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".claude", "templates", "*.md"), func(id, name, content, rel string) {
		parsed.Templates = append(parsed.Templates, TemplateDefinition{ID: id, Name: name, Content: content, SourcePath: rel})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "template"})
	}); err != nil {
		return nil, err
	}

	if parsed.ProjectName == "" {
		parsed.ProjectName = filepath.Base(sourcePath)
	}

	if len(parsed.Agents) == 0 && len(parsed.Commands) == 0 && len(parsed.Rules) == 0 &&
		len(parsed.Templates) == 0 && len(parsed.CustomSections) == 0 {
		return nil, fmt.Errorf("detected Claude Code setup did not contain any parseable migration content")
	}

	return parsed, nil
}

// parseClaudeRoot reads CLAUDE.md and extracts the project name / description
// from its first heading and paragraph.
func parseClaudeRoot(sourcePath string, parsed *ParsedSetup) error {
	rootPath := filepath.Join(sourcePath, "CLAUDE.md")
	if !files.FileExists(rootPath) {
		return nil
	}

	data, err := files.ReadFile(rootPath)
	if err != nil {
		return fmt.Errorf("parse CLAUDE.md: %w", err)
	}
	content := string(data)
	parsed.Files = append(parsed.Files, ParsedFile{Path: "CLAUDE.md", Content: content, Type: "config"})

	if name := extractFirstHeading(content); name != "" {
		parsed.ProjectName = name
	}
	if description := extractFirstParagraph(content); description != "" {
		parsed.Description = description
	}

	parsed.CustomSections = append(parsed.CustomSections, CustomSection{
		ID:         "claude-root",
		Title:      "Imported CLAUDE.md",
		Content:    content,
		SourcePath: "CLAUDE.md",
	})

	return nil
}

// parseClaudeRootAgents walks `.claude/*.md` (top-level only, not recursing
// into subdirs) and treats each .md file as an agent definition.
func parseClaudeRootAgents(sourcePath string, parsed *ParsedSetup) error {
	matches, err := filepath.Glob(filepath.Join(sourcePath, ".claude", "*.md"))
	if err != nil {
		return err
	}

	for _, match := range matches {
		base := filepath.Base(match)
		if base == "CLAUDE.md" {
			continue
		}

		data, err := files.ReadFile(match)
		if err != nil {
			return fmt.Errorf("parse %s: %w", base, err)
		}

		rel, relErr := filepath.Rel(sourcePath, match)
		if relErr != nil {
			rel = match
		}

		content := string(data)
		id := base[:len(base)-len(filepath.Ext(base))]
		name := extractFirstHeading(content)
		if name == "" {
			name = id
		}

		parsed.Agents = append(parsed.Agents, AgentDefinition{
			ID:         id,
			Name:       name,
			Content:    content,
			SourcePath: rel,
		})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "agent"})
	}

	return nil
}
