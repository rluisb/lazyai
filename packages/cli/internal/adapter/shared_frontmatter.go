package adapter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

// ExtractTools parses the tools list from YAML frontmatter in content.
func ExtractTools(content string) []string {
	fm, _, err := frontmatter.ExtractFrontmatter([]byte(content))
	if err != nil || fm == nil {
		return nil
	}
	toolsVal, ok := fm["tools"]
	if !ok {
		return nil
	}

	switch v := toolsVal.(type) {
	case string:
		parts := strings.Fields(v)
		var result []string
		for _, p := range parts {
			p = strings.TrimRight(p, ",")
			if p != "" {
				result = append(result, p)
			}
		}
		return result
	case []any:
		var result []string
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result = append(result, s)
			}
		}
		return result
	}
	return nil
}

// StripFrontmatterAndInjectModel strips YAML frontmatter and injects the model
// and tools as HTML comments at the top of the content (for OpenCode format).
func StripFrontmatterAndInjectModel(content []byte) []byte {
	str := string(content)
	_, body := frontmatter.SplitYamlFrontmatter(str)
	fm, fmBody, _ := frontmatter.ExtractFrontmatter(content)

	var comments []string
	if fm != nil {
		if model, ok := fm["model"].(string); ok && model != "" {
			comments = append(comments, fmt.Sprintf("<!-- Recommended model: %s -->", model))
		}
		tools := ExtractTools(string(fmBody))
		if len(tools) > 0 {
			comments = append(comments, fmt.Sprintf("<!-- allowed-tools: %s -->", strings.Join(tools, ", ")))
		}
	}

	bodyStr := strings.TrimLeft(body, "\n")
	if len(comments) == 0 {
		return []byte(bodyStr)
	}
	return []byte(strings.Join(comments, "\n") + "\n\n" + bodyStr)
}

// NormalizeToolsFrontmatter rewrites the tools line in frontmatter to use the
// given delimiter format.
func NormalizeToolsFrontmatter(content string, delimiter string) string {
	fmBody, body := frontmatter.SplitYamlFrontmatter(content)
	if fmBody == "" {
		return content
	}

	tools := parseToolsFromFrontmatterBody(fmBody)
	if len(tools) == 0 {
		return content
	}

	var joined string
	switch delimiter {
	case "comma":
		joined = strings.Join(tools, ", ")
	default:
		joined = strings.Join(tools, " ")
	}

	rebuilt := rewriteToolsLine(fmBody, joined)
	bodyStr := body
	if !strings.HasPrefix(bodyStr, "\n") {
		bodyStr = "\n" + bodyStr
	}
	return "---\n" + rebuilt + "\n---\n" + bodyStr
}

// EnsureModeAgentFrontmatter ensures the content has mode: agent in its frontmatter.
func EnsureModeAgentFrontmatter(content string) string {
	fmBody, body := frontmatter.SplitYamlFrontmatter(content)
	if fmBody == "" {
		return "---\nmode: agent\n---\n\n" + content
	}

	lines := strings.Split(fmBody, "\n")
	replaced := false
	for i, line := range lines {
		if strings.HasPrefix(line, "mode:") {
			lines[i] = "mode: agent"
			replaced = true
			break
		}
	}
	if !replaced {
		lines = append(lines, "mode: agent")
	}

	bodyStr := body
	if !strings.HasPrefix(bodyStr, "\n") {
		bodyStr = "\n" + bodyStr
	}
	return "---\n" + strings.Join(lines, "\n") + "\n---\n" + bodyStr
}

// parseToolsFromFrontmatterBody extracts tool names from a YAML frontmatter body.
func parseToolsFromFrontmatterBody(body string) []string {
	lines := strings.Split(body, "\n")
	toolsLineIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "tools:") {
			toolsLineIdx = i
			break
		}
	}
	if toolsLineIdx == -1 {
		return nil
	}

	afterColon := strings.TrimSpace(strings.TrimPrefix(lines[toolsLineIdx], "tools:"))
	if afterColon != "" {
		return splitToolList(afterColon)
	}

	// Parse YAML list items (- item)
	var result []string
	for i := toolsLineIdx + 1; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}
		item := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func splitToolList(value string) []string {
	var result []string
	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return r == ' ' || r == ',' || r == '\t'
	}) {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func rewriteToolsLine(frontmatterBody, joined string) string {
	lines := strings.Split(frontmatterBody, "\n")
	toolsLineIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "tools:") {
			toolsLineIdx = i
			break
		}
	}
	if toolsLineIdx == -1 {
		return frontmatterBody
	}

	var result []string
	result = append(result, lines[:toolsLineIdx]...)
	result = append(result, "tools: "+joined)

	// Skip YAML list items after tools:
	i := toolsLineIdx + 1
	for i < len(lines) {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "- ") {
			i++
			continue
		}
		break
	}
	result = append(result, lines[i:]...)
	return strings.Join(result, "\n")
}

// WriteJSONFile writes a JSON-marshaled struct to a file with indentation.
func WriteJSONFile(path string, data any) error {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')
	return files.WriteFile(path, content, 0o644)
}

// truncateOutput truncates s to maxLen characters, appending "..." when truncated.
func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ChmodScriptsExecutable walks a directory and chmods all .sh files to 0o755.
func ChmodScriptsExecutable(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".sh") {
			if err := os.Chmod(path, 0o755); err != nil {
				return err
			}
		}
		return nil
	})
}
