package libraryembed

import "embed"

// FS embeds the LazyAI library assets for the lazyai-cli binary.
// Paths are rooted at this directory, e.g. "agents/builder.md".
//
//go:embed all:agents all:chatmodes all:claudecode all:constitution all:copilot all:fragments all:infra all:mcp all:opencode all:orchestration all:prompts all:root all:rules all:skills all:specs-agents all:standards all:templates all:tool-agents all:tool-templates
var FS embed.FS
