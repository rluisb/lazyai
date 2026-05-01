import path from 'node:path'
import type { ToolId } from '../types.js'

export function resolveGlobalToolTargetDir(tool: ToolId, homeDir: string): string | null {
  switch (tool) {
    case 'opencode': {
      const xdgConfigHome = process.env.XDG_CONFIG_HOME
      if (xdgConfigHome) {
        return path.join(xdgConfigHome, 'opencode')
      }
      return path.join(homeDir, '.config', 'opencode')
    }
    case 'claude-code':
      return path.join(homeDir, '.claude')
    case 'copilot':
      return path.join(homeDir, '.copilot')
    default:
      return null
  }
}

export function isGlobalSupportedTool(tool: ToolId): boolean {
  return tool === 'opencode' || tool === 'claude-code' || tool === 'copilot'
}

export function logUnsupportedGlobalTool(tool: ToolId): void {
  void tool
}
