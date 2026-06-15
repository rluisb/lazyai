package handoff

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"gopkg.in/yaml.v3"
)

// ProgressStatus is the frontmatter status for a handoff.
type ProgressStatus string

const (
	ProgressDone       ProgressStatus = "done"
	ProgressInProgress ProgressStatus = "in-progress"
	ProgressPending    ProgressStatus = "pending"
)

// ProgressItems are the per-status body items rendered in the Progress section.
type ProgressItems struct {
	Done       []string
	InProgress []string
	Pending    []string
}

// Document is the runtime handoff document contract.
type Document struct {
	Goal            string
	Constraints     []string
	Progress        ProgressStatus
	Decisions       []string
	CriticalContext string
	NextSteps       []string
	Risks           []string
	Owner           string
	SessionID       string
	OpenQuestions   []string
	Items           ProgressItems
}

type frontmatterDocument struct {
	Goal            string         `yaml:"goal"`
	Constraints     []string       `yaml:"constraints"`
	Progress        ProgressStatus `yaml:"progress"`
	Decisions       []string       `yaml:"decisions"`
	CriticalContext string         `yaml:"critical_context"`
	NextSteps       []string       `yaml:"next_steps"`
	Risks           []string       `yaml:"risks,omitempty"`
	Owner           string         `yaml:"owner,omitempty"`
	SessionID       string         `yaml:"session_id,omitempty"`
}

// Write renders doc and atomically writes it to path.
func Write(path string, doc Document) error {
	doc = normalize(doc)
	content, err := render(doc)
	if err != nil {
		return err
	}
	return atomicWrite(path, content, 0o644)
}

// Read reads a handoff markdown file and parses its frontmatter and body.
func Read(path string) (Document, error) {
	data, err := files.ReadFile(path)
	if err != nil {
		return Document{}, err
	}

	fm, body, err := frontmatter.ParseYamlFrontmatter(string(data))
	if err != nil {
		return Document{}, fmt.Errorf("parse handoff frontmatter: %w", err)
	}
	if fm == nil {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing frontmatter")
	}

	doc := Document{
		Goal:            stringField(fm, "goal"),
		Constraints:     stringSliceField(fm, "constraints"),
		Progress:        ProgressStatus(stringField(fm, "progress")),
		Decisions:       stringSliceField(fm, "decisions"),
		CriticalContext: stringField(fm, "critical_context"),
		NextSteps:       stringSliceField(fm, "next_steps"),
		Risks:           stringSliceField(fm, "risks"),
		Owner:           stringField(fm, "owner"),
		SessionID:       stringField(fm, "session_id"),
	}

	sections := parseSections(body)
	doc.OpenQuestions = parseListSection(sections["Open Assumptions/Questions"])
	doc.Items = parseProgressSection(sections["Progress"])

	if doc.Goal == "" {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing goal")
	}
	if doc.Progress == "" {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing progress")
	}
	if len(doc.Constraints) == 0 {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing constraints")
	}
	if len(doc.Decisions) == 0 {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing decisions")
	}
	if doc.CriticalContext == "" {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing critical_context")
	}
	if len(doc.NextSteps) == 0 {
		return Document{}, fmt.Errorf("parse handoff frontmatter: missing next_steps")
	}

	return doc, nil
}

func normalize(doc Document) Document {
	doc.Goal = strings.TrimSpace(doc.Goal)
	doc.Constraints = cleanItems(doc.Constraints)
	doc.Decisions = cleanItems(doc.Decisions)
	doc.NextSteps = cleanItems(doc.NextSteps)
	doc.Risks = cleanItems(doc.Risks)
	doc.OpenQuestions = cleanItems(doc.OpenQuestions)
	doc.Items.Done = cleanItems(doc.Items.Done)
	doc.Items.InProgress = cleanItems(doc.Items.InProgress)
	doc.Items.Pending = cleanItems(doc.Items.Pending)
	doc.CriticalContext = strings.TrimSpace(doc.CriticalContext)
	doc.Owner = strings.TrimSpace(doc.Owner)
	doc.SessionID = strings.TrimSpace(doc.SessionID)

	if len(doc.Constraints) == 0 {
		doc.Constraints = []string{"No runtime-recorded constraints."}
	}
	if len(doc.Decisions) == 0 {
		doc.Decisions = []string{"No explicit runtime-recorded decisions."}
	}
	if doc.CriticalContext == "" {
		doc.CriticalContext = "Resume from the recorded session state and dispatch history."
	}
	if len(doc.NextSteps) == 0 {
		doc.NextSteps = []string{"Review the session goal and latest dispatch before resuming."}
	}
	if doc.Progress == "" {
		doc.Progress = inferProgress(doc.Items)
	}

	return doc
}

func inferProgress(items ProgressItems) ProgressStatus {
	if len(items.Pending) > 0 {
		return ProgressPending
	}
	if len(items.InProgress) > 0 {
		return ProgressInProgress
	}
	return ProgressDone
}

func render(doc Document) ([]byte, error) {
	fm, err := yaml.Marshal(frontmatterDocument{
		Goal:            doc.Goal,
		Constraints:     doc.Constraints,
		Progress:        doc.Progress,
		Decisions:       doc.Decisions,
		CriticalContext: doc.CriticalContext,
		NextSteps:       doc.NextSteps,
		Risks:           doc.Risks,
		Owner:           doc.Owner,
		SessionID:       doc.SessionID,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal handoff frontmatter: %w", err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.Write(fm)
	b.WriteString("---\n\n")
	b.WriteString("# Session Handoff\n\n")
	writeSection(&b, "Goal", doc.Goal)
	writeBulletSection(&b, "Constraints & Preferences", doc.Constraints)
	writeProgressSection(&b, doc.Items)
	writeBulletSection(&b, "Key Decisions", doc.Decisions)
	writeSection(&b, "Critical Context", doc.CriticalContext)
	writeOrderedSection(&b, "Next Steps", doc.NextSteps)
	writeBulletSection(&b, "Open Assumptions/Questions", placeholderIfEmpty(doc.OpenQuestions))
	writeBulletSection(&b, "Risks/Watchouts", placeholderIfEmpty(doc.Risks))

	return []byte(b.String()), nil
}

func writeSection(b *strings.Builder, title string, body string) {
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(strings.TrimSpace(body))
	b.WriteString("\n\n")
}

func writeBulletSection(b *strings.Builder, title string, items []string) {
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, item := range items {
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func writeOrderedSection(b *strings.Builder, title string, items []string) {
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for i, item := range items {
		fmt.Fprintf(b, "%d. %s\n", i+1, item)
	}
	b.WriteString("\n")
}

func writeProgressSection(b *strings.Builder, items ProgressItems) {
	b.WriteString("## Progress\n\n")
	writeProgressBucket(b, "Done", placeholderIfEmpty(items.Done))
	writeProgressBucket(b, "In Progress", placeholderIfEmpty(items.InProgress))
	writeProgressBucket(b, "Pending", placeholderIfEmpty(items.Pending))
}

func writeProgressBucket(b *strings.Builder, title string, items []string) {
	b.WriteString("### ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, item := range items {
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func placeholderIfEmpty(items []string) []string {
	if len(items) == 0 {
		return []string{"None recorded."}
	}
	return items
}

func atomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := files.EnsureDir(dir); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".handoff-*")
	if err != nil {
		return fmt.Errorf("create temp handoff: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("write temp handoff: %w", err)
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("chmod temp handoff: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp handoff: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("replace handoff: %w", err)
	}
	return nil
}

func parseSections(body string) map[string]string {
	sections := make(map[string]string)
	var current string
	var lines []string
	flush := func() {
		if current == "" {
			return
		}
		sections[current] = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	for _, rawLine := range strings.Split(body, "\n") {
		line := strings.TrimRight(rawLine, "\r")
		if strings.HasPrefix(line, "## ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			lines = lines[:0]
			continue
		}
		if current != "" {
			lines = append(lines, line)
		}
	}
	flush()
	return sections
}

func parseProgressSection(section string) ProgressItems {
	subsections := make(map[string]string)
	var current string
	var lines []string
	flush := func() {
		if current == "" {
			return
		}
		subsections[current] = strings.TrimSpace(strings.Join(lines, "\n"))
	}

	for _, rawLine := range strings.Split(section, "\n") {
		line := strings.TrimRight(rawLine, "\r")
		if strings.HasPrefix(line, "### ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "### "))
			lines = lines[:0]
			continue
		}
		if current != "" {
			lines = append(lines, line)
		}
	}
	flush()

	return ProgressItems{
		Done:       parseListSection(subsections["Done"]),
		InProgress: parseListSection(subsections["In Progress"]),
		Pending:    parseListSection(subsections["Pending"]),
	}
}

func parseListSection(section string) []string {
	var items []string
	for _, rawLine := range strings.Split(section, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
		} else if orderedItem, ok := trimOrderedPrefix(line); ok {
			line = orderedItem
		} else {
			continue
		}
		if line == "None recorded." {
			continue
		}
		items = append(items, line)
	}
	return items
}

func trimOrderedPrefix(line string) (string, bool) {
	idx := 0
	for idx < len(line) && line[idx] >= '0' && line[idx] <= '9' {
		idx++
	}
	if idx == 0 || idx+1 >= len(line) || line[idx] != '.' || line[idx+1] != ' ' {
		return "", false
	}
	return strings.TrimSpace(line[idx+2:]), true
}

func stringField(fields map[string]any, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func stringSliceField(fields map[string]any, key string) []string {
	value, ok := fields[key]
	if !ok {
		return nil
	}
	items, ok := value.([]any)
	if ok {
		result := make([]string, 0, len(items))
		for _, item := range items {
			text, ok := item.(string)
			if ok && strings.TrimSpace(text) != "" {
				result = append(result, strings.TrimSpace(text))
			}
		}
		return result
	}
	stringsSlice, ok := value.([]string)
	if ok {
		return cleanItems(stringsSlice)
	}
	return nil
}

func cleanItems(items []string) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
