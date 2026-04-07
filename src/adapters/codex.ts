import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
  installToolContextFiles,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

/**
 * Adapter for OpenAI Codex CLI
 *
 * Structure:
 * - Root: AGENTS.md (shared project notes)
 * - Skills: .agents/skills/{name}/SKILL.md (AgentSkills standard)
 * - No .codex/ project directory (config only in ~/.codex/)
 * - Agents: Inline in AGENTS.md (no separate directory)
 */
export class CodexAdapter implements ToolAdapter {
  getToolId(): string {
    return 'codex'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const agentsDir = isGlobal
      ? path.join(path.dirname(ctx.targetDir), '.agents')
      : path.join(ctx.targetDir, '.agents')

    files.ensureDir(agentsDir)
    files.ensureDir(path.join(agentsDir, 'skills'))

    console.log('🤖  Installing Codex tools...')

    // Codex uses skills in directory format (like Claude Code)
    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(agentsDir, 'skills', name, 'SKILL.md')
      },
    })

    // Install context files (AGENTS.md references agents inline)
    await installToolContextFiles({
      ctx,
      toolDir: agentsDir,
      contextFileName: 'AGENTS.md',
      agentsDestDir: '.', // Inline - agents referenced in root file
      skillsDestDir: 'skills',
    })

    // Install root AGENTS.md template if missing
    await installRootTemplateIfMissing({
      ctx,
      recordPath: 'AGENTS.md',
      destPath: path.join(ctx.targetDir, 'AGENTS.md'),
      templateSource: 'root/AGENTS.template.md',
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void ctx
    console.log('🗑️  Removing Codex tools...')
    // Basic remove implementation
  }
}
