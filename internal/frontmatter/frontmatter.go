// Package frontmatter provides YAML frontmatter parsing for markdown files.
// Ported from the TypeScript utilities in src/utils/frontmatter.ts.
package frontmatter

import (
	"regexp"

	"gopkg.in/yaml.v3"
)

// yamlFrontmatterRe matches YAML frontmatter delimited by --- at the start of content.
var yamlFrontmatterRe = regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*(?:\n|$)`)

// ExtractFrontmatter splits content into frontmatter and body.
// Returns the parsed frontmatter as a map, the remaining body content, and any error.
func ExtractFrontmatter(content []byte) (map[string]any, []byte, error) {
	loc := yamlFrontmatterRe.FindSubmatchIndex(content)
	if loc == nil {
		// No frontmatter: return empty map and full content as body.
		return map[string]any{}, content, nil
	}

	// loc[2]:loc[3] is the first capture group (YAML body between --- delimiters).
	frontmatterBody := content[loc[2]:loc[3]]
	// Everything after the full match is the body.
	body := content[loc[1]:]

	var fm map[string]any
	if err := yaml.Unmarshal(frontmatterBody, &fm); err != nil {
		return nil, content, err
	}

	return fm, body, nil
}

// HasFrontmatter reports whether content starts with a YAML frontmatter block.
func HasFrontmatter(content []byte) bool {
	return yamlFrontmatterRe.Match(content)
}

// ExtractField returns a string field from a frontmatter map.
// Returns an empty string if the key is missing or the value is not a string.
func ExtractField(frontmatter map[string]any, key string) string {
	val, ok := frontmatter[key]
	if !ok {
		return ""
	}
	str, ok := val.(string)
	if !ok {
		return ""
	}
	return str
}

// StripFrontmatter removes the YAML frontmatter block from content and returns the body.
func StripFrontmatter(content []byte) []byte {
	return yamlFrontmatterRe.ReplaceAll(content, nil)
}

// SplitYamlFrontmatter returns the raw frontmatter body and the remaining
// body as strings, or ("", original content) if no frontmatter is found.
func SplitYamlFrontmatter(content string) (frontmatterBody string, body string) {
	match := yamlFrontmatterRe.FindStringSubmatchIndex(content)
	if match == nil {
		return "", content
	}
	// match[2]:match[3] is the first capture group (frontmatter body).
	frontmatterBody = content[match[2]:match[3]]
	// Everything after the full match is the body.
	body = content[match[1]:]
	return frontmatterBody, body
}

// ParseYamlFrontmatter parses frontmatter from content, returning the parsed
// map and the body text. If no frontmatter exists, returns nil, "", nil.
func ParseYamlFrontmatter(content string) (map[string]any, string, error) {
	match := yamlFrontmatterRe.FindStringSubmatchIndex(content)
	if match == nil {
		return nil, content, nil
	}

	frontmatterBody := content[match[2]:match[3]]
	body := content[match[1]:]

	var fm map[string]any
	if err := yaml.Unmarshal([]byte(frontmatterBody), &fm); err != nil {
		return nil, content, err
	}

	return fm, body, nil
}
