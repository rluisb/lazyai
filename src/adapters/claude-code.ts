import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
  installToolContextFiles,
} from './shared.js'

export class ClaudeCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'claude-code'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const claudeDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.claude')
    files.ensureDir(claudeDir)
    files.ensureDir(path.join(claudeDir, 'rules'))
    if (!isGlobal) {
      files.ensureDir(path.join(claudeDir, 'agents'))
    }
    files.ensureDir(path.join(claudeDir, 'commands'))
    files.ensureDir(path.join(claudeDir, 'templates'))

    console.log('🤖  Installing Claude Code tools...')

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(claudeDir, isGlobal ? file : 'agents', file),
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => path.join(claudeDir, 'commands', file),
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'prompts',
      selectionKey: 'prompts',
      toDestPath: (file) => path.join(claudeDir, 'templates', file),
    })

    await installToolContextFiles({
      ctx,
      toolDir: claudeDir,
      contextFileName: 'CLAUDE.md',
      agentsDestDir: isGlobal ? '.' : 'agents',
      skillsDestDir: 'commands',
      templatesDestDir: 'templates',
    })

    await installRootTemplateIfMissing({
      ctx,
      recordPath: 'CLAUDE.md',
      destPath: path.join(isGlobal ? claudeDir : ctx.targetDir, 'CLAUDE.md'),
      templateSource: 'root/CLAUDE.template.md',
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void ctx
    console.log('🗑️  Removing Claude Code tools...')
    // Basic remove implementation
  }
}
