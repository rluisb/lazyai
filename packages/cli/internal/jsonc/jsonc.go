// Package jsonc provides JSONC (JSON with comments) support.
// Ported from the TypeScript utilities in src/utils/jsonc.ts.
package jsonc

import (
	"encoding/json"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

// StripComments removes single-line (//) and multi-line (/* */) comments
// from a JSONC byte slice without modifying comment-like sequences inside
// string literals (e.g. URLs like "https://example.com").
func StripComments(input []byte) []byte {
	out := make([]byte, 0, len(input))
	i := 0
	inString := false
	n := len(input)

	for i < n {
		c := input[i]
		var next byte
		if i+1 < n {
			next = input[i+1]
		}

		if inString {
			out = append(out, c)
			if c == '\\' && next != 0 {
				out = append(out, next)
				i += 2
				continue
			}
			if c == '"' {
				inString = false
			}
			i++
			continue
		}

		if c == '"' {
			inString = true
			out = append(out, c)
			i++
			continue
		}

		if c == '/' && next == '/' {
			// Skip single-line comment until end of line.
			for i < n && input[i] != '\n' {
				i++
			}
			continue
		}

		if c == '/' && next == '*' {
			// Skip multi-line comment.
			i += 2
			for i < n-1 {
				if input[i] == '*' && input[i+1] == '/' {
					i += 2
					break
				}
				i++
			}
			// If we reached the end without finding closing */, just stop.
			if i >= n-1 {
				i = n
			}
			continue
		}

		out = append(out, c)
		i++
	}

	return out
}

// ParseJSONC strips comments from input and parses the result as JSON into a map.
func ParseJSONC(input []byte) (map[string]any, error) {
	stripped := StripComments(input)
	var result map[string]any
	if err := json.Unmarshal(stripped, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ReadJSONCFile reads a file and parses it as JSONC.
func ReadJSONCFile(path string) (map[string]any, error) {
	data, err := files.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseJSONC(data)
}
