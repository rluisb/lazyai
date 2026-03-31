import { AdapterRegistry } from '../adapters/registry.js'
import type { ToolId, AgentId, SkillId, PromptId, FileRecord } from '../types.js'

export interface ScaffoldAgentsSkillsPromptsOptions {
  targetDir: string
  libraryDir: string
  tools: ToolId[]
  agents: AgentId[]
  skills: SkillId[]
  prompts: PromptId[]
  fileRecords: FileRecord[]
  force?: boolean
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
  const { targetDir, libraryDir, tools, fileRecords, force } = opts

  const registry = new AdapterRegistry()
  const adapters = registry.getAll(tools)

  for (const adapter of adapters) {
    await adapter.install({
      targetDir,
      libraryDir,
      fileRecords,
      force,
    })
  }
}
