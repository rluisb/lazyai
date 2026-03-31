import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { confirmReplace } from '../utils/conflicts.js'

export class ClaudeCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'claude-code'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const claudeDir = path.join(ctx.targetDir, '.claude')
    files.ensureDir(claudeDir)
    files.ensureDir(path.join(claudeDir, 'rules'))
    files.ensureDir(path.join(claudeDir, 'commands'))
    files.ensureDir(path.join(claudeDir, 'templates'))

    console.log('🤖  Installing Claude Code tools...')

    // Copy agent files to .claude/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      await this.copyFileWithRecord(
        path.join(agentsDir, file),
        path.join(claudeDir, file),
        ctx,
      )
    }

    // Skills - exact copy to .claude/commands
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      await this.copyFileWithRecord(
        path.join(skillsDir, file),
        path.join(claudeDir, 'commands', file),
        ctx,
      )
    }

    // Templates - exact copy to .claude/templates
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      await this.copyFileWithRecord(
        path.join(templatesDir, file),
        path.join(claudeDir, 'templates', file),
        ctx,
      )
    }

    // Create CLAUDE.md from template (only if not already created by scaffold)
    const alreadyCreated = ctx.fileRecords.some(r => r.path === 'CLAUDE.md')
    if (!alreadyCreated) {
      const templatePath = path.join(ctx.libraryDir, 'root', 'CLAUDE.template.md')
      const claudeMdPath = path.join(ctx.targetDir, 'CLAUDE.md')

      if (files.fileExists(templatePath)) {
        const shouldWrite = await confirmReplace(claudeMdPath, 'CLAUDE.md')
        if (shouldWrite) {
          const content = files.readFile(templatePath)
          files.writeFile(claudeMdPath, content)
          ctx.fileRecords.push({
            path: 'CLAUDE.md',
            hash: files.fileHash(claudeMdPath),
            source: 'root/CLAUDE.template.md',
          })
        }
      }
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const claudeDir = path.join(ctx.targetDir, '.claude')
    console.log('🗑️  Removing Claude Code tools...')
    // Basic remove implementation
  }

  private async copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const shouldWrite = await confirmReplace(dest, path.relative(ctx.targetDir, dest))
    if (!shouldWrite) return

    files.copyFile(src, dest)
    ctx.fileRecords.push({
      path: path.relative(ctx.targetDir, dest),
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
