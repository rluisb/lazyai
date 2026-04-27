import { existsSync, readFileSync } from 'node:fs'
import { homedir } from 'node:os'
import { join } from 'node:path'

export interface TomlConfig {
  default_scope?: 'project' | 'workspace' | 'global'
  default_tools?: string[]
  install_mode?: 'copy' | 'symlink'
  enabled_servers?: string[]
  project_name?: string

  wizard?: {
    preset?: 'minimal' | 'recommended' | 'full'
    show_preview?: boolean
  }

  [key: string]: unknown
}

/**
 * Parse a minimal subset of TOML. Handles:
 * - key = "value", key = 'value', key = value
 * - [section] headers
 * - key = ["a", "b"] arrays
 * - # comments
 * Does NOT handle: nested tables, dotted keys, dates, multiline strings, inline tables.
 */
export function parseToml(content: string): TomlConfig {
  const result: Record<string, unknown> = {}
  let section: Record<string, unknown> = result
  const lines = content.split('\n')

  for (const line of lines) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue

    // Section header: [section]
    const sectionMatch = trimmed.match(/^\[(\w+)\]$/)
    if (sectionMatch) {
      const name = sectionMatch[1] ?? 'unknown'
      section = (result[name] as Record<string, unknown>) ?? {}
      result[name] = section
      continue
    }

    // Key-value pair
    const eqIndex = trimmed.indexOf('=')
    if (eqIndex === -1) continue

    const key = trimmed.slice(0, eqIndex).trim()
    const rawValue = trimmed.slice(eqIndex + 1).trim()
    const value = parseTomlValue(rawValue)

    section[key] = value
  }

  return result as TomlConfig
}

function parseTomlValue(raw: string): unknown {
  // Array: ["a", "b"]
  if (raw.startsWith('[') && raw.endsWith(']')) {
    const inner = raw.slice(1, -1).trim()
    if (!inner) return []
    return inner
      .split(',')
      .map((v) => parseTomlValue(v.trim()))
  }

  // Quoted string
  if ((raw.startsWith('"') && raw.endsWith('"')) || (raw.startsWith("'") && raw.endsWith("'"))) {
    return raw.slice(1, -1)
  }

  // Boolean
  if (raw === 'true') return true
  if (raw === 'false') return false

  // Number
  const num = Number(raw)
  if (!Number.isNaN(num)) return num

  // Bare string
  return raw
}

/**
 * Load and merge TOML config from global and project paths.
 * Project wins over global.
 */
export function loadConfig(targetDir: string): TomlConfig {
  const globalPath = join(homedir(), '.config', 'ai-setup', 'config.toml')
  const projectPath = join(targetDir, '.ai-setup.toml')

  let config: TomlConfig = {}

  if (existsSync(globalPath)) {
    try {
      config = { ...parseToml(readFileSync(globalPath, 'utf-8')) }
    } catch {
      // Ignore parse errors in global config
    }
  }

  if (existsSync(projectPath)) {
    try {
      const projectConfig = parseToml(readFileSync(projectPath, 'utf-8'))
      // Shallow merge: project overrides global per key
      config = { ...config, ...projectConfig }
      // Merge wizard sub-section
      if (projectConfig.wizard) {
        config.wizard = { ...config.wizard, ...projectConfig.wizard }
      }
    } catch {
      // Ignore parse errors in project config
    }
  }

  return config
}
