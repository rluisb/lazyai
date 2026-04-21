package migration

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// CanonicalWriteResult holds the result of writing files in canonical format.
type CanonicalWriteResult struct {
	Agents     []string
	Skills     []string
	Prompts    []string
	Rules      []string
	RootConfig string
	Skipped    []string
}

type canonicalTarget struct {
	TargetPath string
	SourcePath string
	Content    string
	Kind       ActionType
}

var nonAlphaNumRe = regexp.MustCompile(`[^a-z0-9]+`)
var leadingTrailingDashRe = regexp.MustCompile(`^-+|-+$`)

// WriteCanonical writes parsed data into the .ai/ canonical format.
// Ported from writeToCanonical in src/migration/canonical-writer.ts.
func WriteCanonical(targetDir string, setup *ParsedSetup, fileRecords *[]types.TrackedFile, dryRun bool) (*CanonicalWriteResult, error) {
	aiDir := filepath.Join(targetDir, ".ai")
	result := &CanonicalWriteResult{
		Agents:  []string{},
		Skills:  []string{},
		Prompts: []string{},
		Rules:   []string{},
		Skipped: []string{},
	}

	if !dryRun {
		if err := files.EnsureDir(aiDir); err != nil {
			return nil, err
		}
	}

	// Write agents.
	agentsDir := filepath.Join(aiDir, "agents")
	if !dryRun {
		if err := files.EnsureDir(agentsDir); err != nil {
			return nil, err
		}
	}
	for _, agent := range setup.Agents {
		fileName := normalizeName(agent.ID) + ".md"
		if fileName == ".md" {
			fileName = normalizeName(agent.Name) + ".md"
		}
		destination := filepath.Join(agentsDir, fileName)
		wrote, err := writeCanonicalFile(destination, []byte(agent.Content), targetDir, "migrated:"+agent.SourcePath, fileRecords, dryRun, result)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", destination, err)
		}
		if wrote {
			rel, _ := filepath.Rel(targetDir, destination)
			result.Agents = append(result.Agents, rel)
		}
	}

	// Write skills (commands).
	skillsDir := filepath.Join(aiDir, "skills")
	if !dryRun {
		if err := files.EnsureDir(skillsDir); err != nil {
			return nil, err
		}
	}
	for _, cmd := range setup.Commands {
		fileName := normalizeName(cmd.ID) + ".md"
		if fileName == ".md" {
			fileName = normalizeName(cmd.Name) + ".md"
		}
		destination := filepath.Join(skillsDir, fileName)
		wrote, err := writeCanonicalFile(destination, []byte(cmd.Content), targetDir, "migrated:"+cmd.SourcePath, fileRecords, dryRun, result)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", destination, err)
		}
		if wrote {
			rel, _ := filepath.Rel(targetDir, destination)
			result.Skills = append(result.Skills, rel)
		}
	}

	// Write prompts (templates).
	promptsDir := filepath.Join(aiDir, "prompts")
	if !dryRun {
		if err := files.EnsureDir(promptsDir); err != nil {
			return nil, err
		}
	}
	for _, tmpl := range setup.Templates {
		fileName := normalizeName(tmpl.ID) + ".md"
		if fileName == ".md" {
			fileName = normalizeName(tmpl.Name) + ".md"
		}
		destination := filepath.Join(promptsDir, fileName)
		wrote, err := writeCanonicalFile(destination, []byte(tmpl.Content), targetDir, "migrated:"+tmpl.SourcePath, fileRecords, dryRun, result)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", destination, err)
		}
		if wrote {
			rel, _ := filepath.Rel(targetDir, destination)
			result.Prompts = append(result.Prompts, rel)
		}
	}

	// Write rules.
	rulesDir := filepath.Join(aiDir, "rules")
	if !dryRun {
		if err := files.EnsureDir(rulesDir); err != nil {
			return nil, err
		}
	}
	for _, rule := range setup.Rules {
		fileName := normalizeName(rule.ID) + ".md"
		if fileName == ".md" {
			fileName = normalizeName(rule.Category) + ".md"
		}
		destination := filepath.Join(rulesDir, fileName)
		wrote, err := writeCanonicalFile(destination, []byte(rule.Content), targetDir, "migrated:"+rule.SourcePath, fileRecords, dryRun, result)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", destination, err)
		}
		if wrote {
			rel, _ := filepath.Rel(targetDir, destination)
			result.Rules = append(result.Rules, rel)
		}
	}

	// Write custom sections as root config.
	if len(setup.CustomSections) > 0 {
		adapterName := setup.Metadata["adapter"]
		if adapterName == "" {
			adapterName = setup.ProjectName
		}
		if adapterName == "" {
			adapterName = "migration"
		}
		destination := filepath.Join(aiDir, "constitution", normalizeName(adapterName)+".md")
		var sections []string
		for _, section := range setup.CustomSections {
			sections = append(sections, "## "+section.Title+"\n\n"+section.Content)
		}
		content := []byte(strings.Join(sections, "\n\n"))
		wrote, err := writeCanonicalFile(destination, content, targetDir, "migrated:"+adapterName, fileRecords, dryRun, result)
		if err != nil {
			return nil, fmt.Errorf("write %s: %w", destination, err)
		}
		if wrote {
			rel, _ := filepath.Rel(targetDir, destination)
			result.RootConfig = rel
		}
	}

	return result, nil
}

func PlannedCanonicalTargets(setup *ParsedSetup) []canonicalTarget {
	targets := []canonicalTarget{}

	for _, agent := range setup.Agents {
		targets = append(targets, canonicalTarget{
			TargetPath: filepath.Join(".ai", "agents", normalizeName(firstNonEmpty(agent.ID, agent.Name))+".md"),
			SourcePath: agent.SourcePath,
			Content:    agent.Content,
			Kind:       ActionTypeCreate,
		})
	}
	for _, command := range setup.Commands {
		targets = append(targets, canonicalTarget{
			TargetPath: filepath.Join(".ai", "skills", normalizeName(firstNonEmpty(command.ID, command.Name))+".md"),
			SourcePath: command.SourcePath,
			Content:    command.Content,
			Kind:       ActionTypeCreate,
		})
	}
	for _, template := range setup.Templates {
		targets = append(targets, canonicalTarget{
			TargetPath: filepath.Join(".ai", "prompts", normalizeName(firstNonEmpty(template.ID, template.Name))+".md"),
			SourcePath: template.SourcePath,
			Content:    template.Content,
			Kind:       ActionTypeCreate,
		})
	}
	for _, rule := range setup.Rules {
		targets = append(targets, canonicalTarget{
			TargetPath: filepath.Join(".ai", "rules", normalizeName(firstNonEmpty(rule.ID, rule.Category))+".md"),
			SourcePath: rule.SourcePath,
			Content:    rule.Content,
			Kind:       ActionTypeCreate,
		})
	}
	if len(setup.CustomSections) > 0 {
		adapterName := setup.Metadata["adapter"]
		if adapterName == "" {
			adapterName = setup.ProjectName
		}
		if adapterName == "" {
			adapterName = "migration"
		}
		targets = append(targets, canonicalTarget{
			TargetPath: filepath.Join(".ai", "constitution", normalizeName(adapterName)+".md"),
			SourcePath: setup.CustomSections[0].SourcePath,
			Content:    adapterName,
			Kind:       ActionTypeCreate,
		})
	}

	return targets
}

// writeCanonicalFile writes a file if it doesn't already exist. Returns true
// if the file was written (or would have been in dry-run mode).
func writeCanonicalFile(destination string, content []byte, targetDir, source string, fileRecords *[]types.TrackedFile, dryRun bool, result *CanonicalWriteResult) (bool, error) {
	if files.FileExists(destination) {
		rel, _ := filepath.Rel(targetDir, destination)
		result.Skipped = append(result.Skipped, rel)
		return false, nil
	}

	if dryRun {
		return true, nil
	}

	if err := files.WriteFile(destination, content, 0o644); err != nil {
		return false, err
	}

	hash, _ := files.FileHash(destination)
	rel, _ := filepath.Rel(targetDir, destination)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   rel,
		Hash:   hash,
		Source: source,
		Owner:  types.FileOwnerMigrated,
		Status: types.FileStatusInstalled,
	})

	return true, nil
}

// normalizeName creates a safe filename from a free-form string.
func normalizeName(value string) string {
	s := strings.TrimSpace(value)
	s = strings.ToLower(s)
	s = nonAlphaNumRe.ReplaceAllString(s, "-")
	s = leadingTrailingDashRe.ReplaceAllString(s, "")
	if s == "" {
		return "item"
	}
	return s
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return "item"
}
