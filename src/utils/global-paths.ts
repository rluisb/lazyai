import path from 'node:path'
import type { ToolId } from '../types.js'

export function resolveGlobalToolTargetDir(tool: ToolId, homeDir: string): string | null {
  if (tool === 'opencode') {
    return path.join(homeDir, '.config', 'opencode')
  }

  if (tool === 'claude-code') {
    return path.join(homeDir, '.claude')
  }

  return null
}

export function isGlobalSupportedTool(tool: ToolId): boolean {
  return tool === 'opencode' || tool === 'claude-code'
}

export function logUnsupportedGlobalTool(tool: ToolId): void {
  const messages: Partial<Record<ToolId, string>> = {
    copilot: "Copilot doesn't support file-based global config. Use project scope instead.",
    gemini: "Gemini doesn't support file-based global config. Use project scope instead.",
  }

  const message = messages[tool]
  if (message) {
    console.info(message)
  }
}
