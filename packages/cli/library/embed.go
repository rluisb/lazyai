package libraryembed

import "embed"

// FS embeds the active LazyAI library assets for the lazyai-cli binary.
// Paths are rooted at this directory, e.g. "canonical/agents/implementer.md".
//
//go:embed all:antigravity all:canonical all:chatmodes all:claudecode all:constitution all:copilot all:fragments all:hooks all:infra all:mcp all:opencode all:omp all:prompts all:root all:rules all:skills all:specs-agents all:standards all:templates all:tool-agents all:tool-templates all:workflows
var FS embed.FS
