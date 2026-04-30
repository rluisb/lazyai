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
    case 'gemini':
      return path.join(homeDir, '.gemini')
    case 'codex':
      return path.join(homeDir, '.codex')
    default:
      return null
  }
}

export function isGlobalSupportedTool(tool: ToolId): boolean {
  // All tools except pi support global scope.
  // Pi only supports project/workspace because it has no file-based global config path.
  return tool !== 'pi'
}

export function logUnsupportedGlobalTool(tool: ToolId): void {
  // All tools except pi now support global scope.
  // This function is kept for backward compatibility but only logs for pi.
  if (tool === 'pi') {
    console.info("Pi doesn't support file-based global config. Use project or workspace scope instead.")
  }
}

export interface CodexRoots {
  configRoot: string
  skillsRoot: string
}

export function resolveCodexRoots(
  scope: 'global' | 'workspace' | 'project',
  targetDir: string,
  homeDir: string,
  workspaceRoot?: string,
): CodexRoots {
  switch (scope) {
    case 'project':
      return {
        configRoot: path.join(targetDir, '.codex'),
        skillsRoot: path.join(targetDir, '.agents', 'skills'),
      }
    case 'workspace':
      return {
        configRoot: path.join(workspaceRoot || path.dirname(targetDir), '.codex'),
        skillsRoot: path.join(workspaceRoot || path.dirname(targetDir), '.agents', 'skills'),
      }
    case 'global':
      return {
        configRoot: path.join(homeDir, '.codex'),
        skillsRoot: path.join(homeDir, '.agents', 'skills'),
      }
  }
}