import type { ToolId } from '../types.js'

export const ROOT_FILE_BY_TOOL: Record<ToolId, string> = {
  opencode: 'AGENTS.md',
  'claude-code': 'AGENTS.md',
  copilot: '.github/copilot-instructions.md',
}
