import type { ToolId } from '../types.js'

export const ROOT_FILE_BY_TOOL: Record<ToolId, string> = {
  opencode: 'AGENTS.md',
  'claude-code': 'CLAUDE.md',
  gemini: 'GEMINI.md',
  copilot: '.github/copilot-instructions.md',
  codex: 'AGENTS.md',
  pi: 'AGENTS.md',
}
