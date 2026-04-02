import { AdapterRegistry } from '../adapters/registry.js'
import type { ToolId, AgentId, SkillId, PromptId, FileRecord, ConflictStrategy, SetupScope } from '../types.js'

export interface ScaffoldAgentsSkillsPromptsOptions {
  targetDir: string
  setupScope?: SetupScope
  libraryDir: string
  tools: ToolId[]
  agents: AgentId[]
  skills: SkillId[]
  prompts: PromptId[]
  fileRecords: FileRecord[]
  force?: boolean
  strategy?: ConflictStrategy
  perFileOverrides?: Map<string, ConflictStrategy>
}

/**
 * Scaffold agents, skills, and prompts for selected tools.
 *
 * For now, this passes all agents/skills/prompts through to the adapter (the full adapter call).
 * In T17, the adapter interface will be extended to accept `selections` filtering, allowing
 * this function to pass filtered selections to each adapter.
 *
 * Current behavior: Calls adapter.install() which copies all agents/skills/prompts for each tool.
 */
export async function scaffoldAgentsSkillsPrompts(opts: ScaffoldAgentsSkillsPromptsOptions): Promise<void> {
  const { targetDir, setupScope, libraryDir, tools, fileRecords, force, strategy, perFileOverrides } = opts

  const registry = new AdapterRegistry()
  const adapters = registry.getAll(tools)

  for (const adapter of adapters) {
    const strategyOpts = strategy ? { strategy } : {}
    const perFileOverrideOpts = perFileOverrides ? { perFileOverrides } : {}
    const scopeOpts = setupScope ? { setupScope } : {}

    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force,
      ...scopeOpts,
      ...strategyOpts,
      ...perFileOverrideOpts,
    })
  }
}
