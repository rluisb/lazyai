/**
 * Post-install sanity checks for OpenCode. Runs `opencode debug config` and
 * `opencode debug agent <name>` for each installed agent; collects non-fatal
 * warnings. No-ops when the `opencode` binary is not on PATH.
 *
 * Ported from `internal/adapter/opencode_validate.go`.
 */

import { execFileSync, execSync } from 'node:child_process'
import path from 'node:path'
import { fileExists, listDir } from '../utils/files.js'

export interface ValidationWarning {
  scope: 'config' | 'agent'
  item: string
  reason: string
}

export interface OpenCodeValidateContext {
  /** Absolute path to the resolved `.opencode/` directory (or `~/.config/opencode/` for global scope). */
  ocDir: string
}

/** Executes an opencode subcommand and returns stdout. Throws on non-zero exit. */
export type CmdRunner = (command: string, args: string[]) => string

const defaultCmdRunner: CmdRunner = (command, args) => {
  return execFileSync(command, args, { encoding: 'utf-8', stdio: ['ignore', 'pipe', 'pipe'] })
}

function isOpenCodeOnPath(): boolean {
  try {
    execSync('command -v opencode', { stdio: 'ignore' })
    return true
  } catch {
    return false
  }
}

export function validateOpenCodeInstall(
  ctx: OpenCodeValidateContext,
  runner: CmdRunner = defaultCmdRunner,
): ValidationWarning[] {
  if (process.env.AI_SETUP_SKIP_VALIDATION === '1') return []
  if (!isOpenCodeOnPath()) return []

  const warnings: ValidationWarning[] = []

  // Validate config.
  try {
    const out = runner('opencode', ['debug', 'config'])
    const configPath = path.join(ctx.ocDir, 'opencode.jsonc')
    if (!out.includes('mcp') && fileExists(configPath)) {
      warnings.push({
        scope: 'config',
        item: 'opencode.jsonc',
        reason: 'opencode debug config output does not mention mcp — MCP entries may not have been picked up',
      })
    }
  } catch (err) {
    warnings.push({
      scope: 'config',
      item: 'opencode.jsonc',
      reason: `opencode debug config failed: ${err instanceof Error ? err.message : String(err)}`,
    })
  }

  // Validate each installed agent.
  const agentsDir = path.join(ctx.ocDir, 'agents')
  if (!fileExists(agentsDir)) return warnings

  for (const file of listDir(agentsDir)) {
    if (!file.endsWith('.md')) continue
    const name = path.parse(file).name
    try {
      runner('opencode', ['debug', 'agent', name])
    } catch (err) {
      warnings.push({
        scope: 'agent',
        item: name,
        reason: `opencode debug agent failed: ${err instanceof Error ? err.message : String(err)}`,
      })
    }
  }

  return warnings
}

export function formatWarning(w: ValidationWarning): string {
  return `[opencode validate] ${w.scope} / ${w.item}: ${w.reason}`
}
