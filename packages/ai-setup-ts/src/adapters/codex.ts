import path from 'node:path'
import * as files from '../utils/files.js'
import { resolveCodexRoots } from '../utils/global-paths.js'
import { driveCodexMcpViaCli } from './mcp-compiler.js'
import {
  copyLibraryDirectory,
  getOrchestratorSkillContent,
  installRootTemplateIfMissing,
  installToolContextFiles,
  isOrchestratorEnabled,
  writeContentWithRecord,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

/**
 * Adapter for OpenAI Codex CLI
 *
 * Codex uses a two-root architecture:
 * - configRoot (.codex/ or ~/.codex/): holds config.toml + AGENTS.override.md
 * - skillsRoot (.agents/skills/ or ~/.agents/skills/): holds per-skill <name>/SKILL.md
 *
 * At project/workspace scope:
 *   configRoot = <target>/.codex
 *   skillsRoot = <target>/.agents/skills
 *   Root AGENTS.md is placed at <target>/AGENTS.md
 *
 * At global scope:
 *   configRoot = ~/.codex
 *   skillsRoot = ~/.agents/skills
 *   No root AGENTS.md written
 */
export class CodexAdapter implements ToolAdapter {
  getToolId(): string {
    return 'codex'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const scope = ctx.setupScope ?? 'project'
    const homeDir = ctx.homeDir ?? ''
    const roots = resolveCodexRoots(scope, ctx.targetDir, homeDir, ctx.workspaceRoot)

    const configRoot = roots.configRoot
    const skillsRoot = roots.skillsRoot

    files.ensureDir(configRoot)
    files.ensureDir(skillsRoot)

    console.log('🤖  Installing Codex tools...')

    // Codex uses skills in directory format (like Claude Code)
    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(skillsRoot, name, 'SKILL.md')
      },
    })

    if (isOrchestratorEnabled(ctx)) {
      await writeContentWithRecord({
        dest: path.join(skillsRoot, 'orchestrator', 'SKILL.md'),
        content: getOrchestratorSkillContent(ctx),
        ctx,
        source: 'generated:orchestrator-skill',
      })
    }

    // Install context files (AGENTS.md references agents inline)
    await installToolContextFiles({
      ctx,
      toolDir: path.dirname(skillsRoot),
      contextFileName: 'AGENTS.md',
      agentsDestDir: '.', // Inline - agents referenced in root file
      skillsDestDir: 'skills',
    })

    // Write a minimal config.toml with [mcp_servers] placeholder so Codex
    // recognises the project/global config as trusted. User-authored tables
    // survive via merge on subsequent installs.
    const configPath = path.join(configRoot, 'config.toml')
    if (!files.fileExists(configPath)) {
      files.writeFile(configPath, '[mcp_servers]\n')
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, configPath),
        hash: files.fileHash(configPath),
        source: 'generated',
        owner: 'library',
      })
    }

    // Write AGENTS.override.md from library template on first install (never overwrite).
    // This provides a ready-to-edit override scaffold.
    const overridePath = path.join(configRoot, 'AGENTS.override.md')
    if (!files.fileExists(overridePath)) {
      const templatePath = path.join(ctx.libraryDir, 'codex', 'AGENTS.override.template.md')
      if (files.fileExists(templatePath)) {
        files.writeFile(overridePath, files.readFile(templatePath))
      } else {
        // Fallback stub when library template is missing (e.g. minimal test FS)
        files.writeFile(overridePath, [
          '# AGENTS Override',
          '',
          'Add custom instructions here. Codex reads this file at startup',
          'and merges it with the project-level AGENTS.md.',
        ].join('\n'))
      }
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, overridePath),
        hash: files.fileHash(overridePath),
        source: 'generated:codex-override',
        owner: 'library',
      })
    }

    // Install root AGENTS.md template — only at project/workspace scope, not global
    const isGlobal = scope === 'global'
    if (!isGlobal) {
      await installRootTemplateIfMissing({
        ctx,
        recordPath: 'AGENTS.md',
        destPath: path.join(ctx.targetDir, 'AGENTS.md'),
        templateSource: 'root/AGENTS.template.md',
      })

      if (ctx.driveCLI === true || ctx.driveCli === true) {
        driveCodexMcpViaCli(ctx.targetDir, ctx.targetDir)
      }
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void ctx
    console.log('🗑️  Removing Codex tools...')
    // Basic remove implementation
  }
}
