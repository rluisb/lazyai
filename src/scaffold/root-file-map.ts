import type { ToolId } from '../types.js'

export const ROOT_FILE_BY_TOOL: Record<ToolId, string> = {
  opencode: 'AGENTS.md',
  pi: 'INSTRUCTIONS.md',
  'claude-code': 'CLAUDE.md',
  gemini: 'GEMINI.md',
  copilot: '.github/copilot-instructions.md',
  codex: 'AGENTS.md',
}

export const ROOT_TEMPLATE_BY_FILE: Partial<Record<string, string>> = {
  'AGENTS.md': 'root/AGENTS.template.md',
  'INSTRUCTIONS.md': 'root/AGENTS.template.md',
  'CLAUDE.md': 'root/CLAUDE.template.md',
  'GEMINI.md': 'root/GEMINI.template.md',
  '.github/copilot-instructions.md': 'root/copilot-instructions.template.md',
}
