import path from 'node:path'
import * as files from '../utils/files.js'
import {
  driveClaudeMcpViaCli,
} from './mcp-compiler.js'
import {
  copyLibraryDirectory,
  copyWithRecord,
  getOrchestratorAgentContent,
  installRootTemplateIfMissing,
  installToolContextFiles,
  isOrchestratorEnabled,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class ClaudeCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'claude-code'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const claudeDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.claude')
    const settingsPath = path.join(claudeDir, 'settings.json')
    const rulesDir = path.join(claudeDir, 'rules')
    const sampleRulePath = path.join(rulesDir, 'typescript.md')
    files.ensureDir(claudeDir)
    files.ensureDir(rulesDir)
    files.ensureDir(path.join(claudeDir, 'skills'))
    files.ensureDir(path.join(claudeDir, 'agents'))

    if (!files.fileExists(settingsPath)) {
      const defaultSettings = {
        permissions: {
          allow: [],
          deny: [],
        },
      }
      files.writeFile(settingsPath, JSON.stringify(defaultSettings, null, 2))
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, settingsPath),
        hash: files.fileHash(settingsPath),
        source: 'generated',
        owner: 'library',
      })
    }

    if (!files.fileExists(sampleRulePath)) {
      files.writeFile(
        sampleRulePath,
        [
          '---',
          'paths:',
          '  - "src/**/*.ts"',
          '---',
          '',
          '# TypeScript Rules',
          '',
          '- Use strict TypeScript',
          '- Prefer interfaces over types for objects',
        ].join('\n'),
      )
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, sampleRulePath),
        hash: files.fileHash(sampleRulePath),
        source: 'generated',
        owner: 'library',
      })
    }

    console.log('🤖  Installing Claude Code tools...')

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(claudeDir, 'agents', file),
      includeFile: (file) => path.parse(file).name !== 'orchestrator',
    })

    if (isOrchestratorEnabled(ctx)) {
      const orchestratorSource = path.join(ctx.libraryDir, 'agents', 'orchestrator.md')
      await copyWithRecord({
        src: orchestratorSource,
        dest: path.join(claudeDir, 'agents', 'orchestrator.md'),
        ctx,
        transform: () => getOrchestratorAgentContent(ctx),
      })
    }

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(claudeDir, 'skills', name, 'SKILL.md')
      },
    })

    await installToolContextFiles({
      ctx,
      toolDir: claudeDir,
      contextFileName: 'CLAUDE.md',
      agentsDestDir: 'agents',
      skillsDestDir: 'skills',
      skipRootIfExists: isGlobal,
    })

    if (!isGlobal) {
      await installRootTemplateIfMissing({
        ctx,
        recordPath: 'CLAUDE.md',
        destPath: path.join(ctx.targetDir, 'CLAUDE.md'),
        templateSource: 'root/CLAUDE.template.md',
      })
    }

    if ((ctx.driveCLI === true || ctx.driveCli === true) && !isGlobal) {
      driveClaudeMcpViaCli(ctx.targetDir, ctx.targetDir, ctx.setupScope ?? 'project')
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void ctx
    console.log('🗑️  Removing Claude Code tools...')
    // Basic remove implementation
  }
}
