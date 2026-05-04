package migration

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

var headingPattern = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)

var reservedContextMarkdownFiles = map[string]bool{
	"AGENTS.md":               true,
	"CLAUDE.md":               true,
	"copilot-instructions.md": true,
}

// ParseDetectedSetups converts detected setups into parsed migration inputs.
func ParseDetectedSetups(ctx *MigrationContext, detections []DetectionResult) ([]ParsedSetup, error) {
	parsedSetups := make([]ParsedSetup, 0, len(detections))

	for _, detection := range detections {
		parsed, err := parseDetectedSetup(ctx, detection)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", detection.AdapterID, err)
		}
		parsedSetups = append(parsedSetups, *parsed)
	}

	if len(parsedSetups) == 0 {
		return nil, fmt.Errorf("no supported detected setup could be parsed")
	}

	return parsedSetups, nil
}

func parseDetectedSetup(ctx *MigrationContext, detection DetectionResult) (*ParsedSetup, error) {
	switch detection.AdapterID {
	case "opencode":
		return parseOpenCodeSetup(ctx.SourcePath)
	default:
		return nil, fmt.Errorf("unsupported detected setup")
	}
}

func parseOpenCodeSetup(sourcePath string) (*ParsedSetup, error) {
	parsed := &ParsedSetup{
		Agents:         []AgentDefinition{},
		Rules:          []RuleDefinition{},
		Commands:       []CommandDefinition{},
		Templates:      []TemplateDefinition{},
		CustomSections: []CustomSection{},
		Files:          []ParsedFile{},
		Metadata: map[string]string{
			"adapter": "opencode",
		},
	}

	if err := parseOpenCodeRoot(sourcePath, parsed); err != nil {
		return nil, err
	}
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".opencode", "agents", "*.md"), func(id, name, content, rel string) {
		parsed.Agents = append(parsed.Agents, AgentDefinition{ID: id, Name: name, Content: content, SourcePath: rel})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "agent"})
	}); err != nil {
		return nil, err
	}
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".opencode", "commands", "*.md"), func(id, name, content, rel string) {
		parsed.Commands = append(parsed.Commands, CommandDefinition{ID: id, Name: name, Content: content, SourcePath: rel})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "command"})
	}); err != nil {
		return nil, err
	}
	if err := parseOpenCodeMarkdownDir(sourcePath, filepath.Join(".opencode", "templates", "*.md"), func(id, name, content, rel string) {
		parsed.Templates = append(parsed.Templates, TemplateDefinition{ID: id, Name: name, Content: content, SourcePath: rel})
		parsed.Files = append(parsed.Files, ParsedFile{Path: rel, Content: content, Type: "template"})
	}); err != nil {
		return nil, err
	}

	if parsed.ProjectName == "" {
		parsed.ProjectName = filepath.Base(sourcePath)
	}

	if len(parsed.Agents) == 0 && len(parsed.Commands) == 0 && len(parsed.Templates) == 0 && len(parsed.CustomSections) == 0 {
		return nil, fmt.Errorf("detected setup did not contain any parseable migration content")
	}

	return parsed, nil
}

func parseOpenCodeRoot(sourcePath string, parsed *ParsedSetup) error {
	rootPath := filepath.Join(sourcePath, "AGENTS.md")
	if !files.FileExists(rootPath) {
		return nil
	}

	data, err := files.ReadFile(rootPath)
	if err != nil {
		return fmt.Errorf("parse AGENTS.md: %w", err)
	}
	content := string(data)
	parsed.Files = append(parsed.Files, ParsedFile{Path: "AGENTS.md", Content: content, Type: "config"})

	if name := extractFirstHeading(content); name != "" {
		parsed.ProjectName = name
	}
	if description := extractFirstParagraph(content); description != "" {
		parsed.Description = description
	}

	return nil
}

func parseOpenCodeMarkdownDir(sourcePath, pattern string, add func(id, name, content, rel string)) error {
	matches, err := filepath.Glob(filepath.Join(sourcePath, pattern))
	if err != nil {
		return err
	}

	for _, match := range matches {
		base := filepath.Base(match)
		if strings.HasPrefix(base, "_") || reservedContextMarkdownFiles[base] {
			continue
		}

		data, err := files.ReadFile(match)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filepath.Base(match), err)
		}

		rel, relErr := filepath.Rel(sourcePath, match)
		if relErr != nil {
			rel = match
		}

		content := string(data)
		id := strings.TrimSuffix(base, filepath.Ext(base))
		name := extractFirstHeading(content)
		if name == "" {
			name = id
		}

		add(id, name, content, rel)
	}

	return nil
}

func extractFirstHeading(content string) string {
	match := headingPattern.FindStringSubmatch(content)
	if len(match) < 2 {
		return ""
	}
	return strings.TrimSpace(match[1])
}

func extractFirstParagraph(content string) string {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return ""
	}

	parts := strings.Split(trimmed, "\n\n")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || strings.HasPrefix(part, "#") {
			continue
		}
		return part
	}

	return ""
}
