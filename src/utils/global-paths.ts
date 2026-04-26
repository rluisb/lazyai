import path from 'node:path'
import type { ToolId } from '../types.js'

export function resolveGlobalToolTargetDir(tool: ToolId, homeDir: string): string | null {
  if (tool === 'opencode') {
    return path.join(homeDir, '.config', 'opencode')
  }

  if (tool === 'claude-code') {
    return path.join(homeDir, '.claude')
  }

  if (tool === 'copilot') {
    return path.join(homeDir, '.copilot')
  }

  return null
}

export function isGlobalSupportedTool(tool: ToolId): boolean {
  return tool === 'opencode' || tool === 'claude-code' || tool === 'copilot'
}

export function logUnsupportedGlobalTool(tool: ToolId): void {
  const messages: Partial<Record<ToolId, string>> = {
    gemini: "Gemini doesn't support file-based global config. Use project scope instead.",
  }

  const message = messages[tool]
  if (message) {
    console.info(message)
  }
}
