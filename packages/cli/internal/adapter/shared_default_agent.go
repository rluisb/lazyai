package adapter

import (
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
)

func copyCanonicalDefaultAgent(ctx *AdapterContext, dest string, transform func([]byte) []byte) error {
	return CopyWithRecord("canonical/agents/"+defaultAgentID+".md", dest, ctx, true, transform, 0o644)
}

func openCodeDefaultAgentContent(source []byte) []byte {
	fm, body, _ := frontmatter.ExtractFrontmatter(source)
	opts := OpenCodeAgentOpts{
		Description:   inheritedDescription(fm),
		ManagedMarker: managedAgentMarker("opencode", defaultAgentID),
	}
	// Preserve exact body trailing content while adding the marker.
	return BuildOpenCodeAgentFrontmatter(append([]byte{'\n'}, body...), opts)
}

func claudeDefaultAgentContent(source []byte) []byte {
	fm, body, _ := frontmatter.ExtractFrontmatter(source)
	description := inheritedDescription(fm)
	if description == "Agent" {
		description = defaultAgentDescription
	}
	body = trimLeadingNewlines(body)
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: ")
	b.WriteString(defaultAgentID)
	b.WriteByte('\n')
	b.WriteString("description: ")
	b.WriteString(yamlDoubleQuote(description))
	b.WriteString("\n---\n\n")
	b.WriteString(managedAgentMarker("claude", defaultAgentID))
	b.WriteString("\n\n")
	b.Write(body)
	return []byte(b.String())
}
