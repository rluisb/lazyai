import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { extractTools } from '../utils/frontmatter.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const libraryDir = findMonorepoLibraryDir()
const repoRoot = path.dirname(libraryDir)

const CANONICAL_TOOLS = [
  'list_catalog',
  'compose_agent',
  'start_chain',
  'advance_chain',
  'build_team',
  'assign_team_task',
  'complete_team_task',
  'start_workflow',
  'advance_workflow',
  'get_status',
  'get_budget',
  'retry_step',
  'escalate_step',
  'handoff',
  'catalog_list',
  'catalog_list_versions',
  'catalog_get_version',
  'catalog_create_version',
  'catalog_set_active',
  'catalog_deactivate',
  'catalog_remove',
  'catalog_diff',
  'catalog_export_version',
  'catalog_import',
  'invoke_agent',
  'subscribe_run',
  'unsubscribe_run',
  'enqueue_job',
  'get_job',
  'list_jobs',
] as const

function readRegisteredTools(): string[] {
  const serverPath = path.join(repoRoot, 'packages', 'orchestrator', 'src', 'server.ts')
  const source = fs.readFileSync(serverPath, 'utf8')
  const registerRegex = /server\.registerTool\(\s*['"]([a-z_]+)['"]/g
  const names: string[] = []
  let match = registerRegex.exec(source)
  while (match !== null) {
    if (match[1]) names.push(match[1])
    match = registerRegex.exec(source)
  }
  return names
}

function readOrchestratorAgentTools(): string[] {
  const agentPath = path.join(repoRoot, 'library', 'agents', 'orchestrator.md')
  const source = fs.readFileSync(agentPath, 'utf8')
  return extractTools(source)
}

describe('orchestrator tools sync', () => {
  it('orchestrator/src/server.ts registers the canonical tool set', () => {
    const registered = readRegisteredTools()
    expect(new Set(registered)).toEqual(new Set(CANONICAL_TOOLS))
  })

  it('library/agents/orchestrator.md declares the canonical tool set in frontmatter', () => {
    const declared = readOrchestratorAgentTools()
    expect(new Set(declared)).toEqual(new Set(CANONICAL_TOOLS))
  })

  it('agent-declared tools match server-registered tools exactly', () => {
    const declared = new Set(readOrchestratorAgentTools())
    const registered = new Set(readRegisteredTools())
    expect(declared).toEqual(registered)
  })
})
