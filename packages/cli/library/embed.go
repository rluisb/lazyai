package libraryembed

import "embed"

// FS embeds the active LazyAI library assets for the lazyai-cli binary.
// Paths are rooted at this directory, e.g. "canonical/agents/builder.md".
//
//go:embed all:canonical all:chatmodes all:claudecode all:constitution all:copilot/instructions all:fragments all:hooks all:infra all:mcp all:opencode all:prompts all:root all:rules all:specs-agents all:standards all:templates all:tool-agents all:tool-templates
var FS embed.FS
