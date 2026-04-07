import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
  installToolContextFiles,
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
    if (!isGlobal) {
      files.ensureDir(path.join(claudeDir, 'agents'))
    }

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
      toDestPath: (file) => isGlobal ? path.join(claudeDir, file) : path.join(claudeDir, 'agents', file),
    })

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
      agentsDestDir: isGlobal ? '.' : 'agents',
      skillsDestDir: 'skills',
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
