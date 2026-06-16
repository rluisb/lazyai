# Orchestration runtime removal

The legacy orchestration runtime was removed from the active LazyAI product surface during the runtime refactor.

## What changed

- `lazyai-cli task`, `lazyai-cli workflow`, `lazyai-cli orchestration`, `lazyai-cli mcp-setup`, and obsolete eval command surfaces were removed.
- The dedicated `lazyai-orchestrator` module is no longer part of the active workspace build.
- Fortnite/orchestration library content was archived out of the active embed path.
- OpenCode, Claude Code, and Copilot now use the neutral baseline-facing `implementer` contract instead of the old orchestrator entry path.

## Where to look now

- [Migration: Fortnite / orchestrator removal](../migration/fortnite-orchestrator-removal.md)
- [CLI reference](../cli/reference.md)
- [Product boundaries](../concepts/product-boundaries.md)
- [MCP integration](mcp.md)

## Historical note

Older design and research documents in the repository may still discuss the retired orchestration runtime for context. They are historical references, not current product documentation.
