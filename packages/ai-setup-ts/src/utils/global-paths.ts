import path from 'node:path'
import type { ToolId } from '../types.js'

export function resolveGlobalToolTargetDir(tool: ToolId, homeDir: string): string | null {
  if (tool === 'opencode') {
    return path.join(homeDir, '.config', 'opencode')
  }
  return null
}

export function isGlobalSupportedTool(tool: ToolId): boolean {
  return tool === 'opencode'
}

export function logUnsupportedGlobalTool(_tool: ToolId): void {
  // All supported tools (opencode) support global config.
}
