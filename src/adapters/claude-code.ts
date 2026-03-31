import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import type { AgentId, SkillId, PromptId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'

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

    const selectedAgents = ctx.selections?.agents ? new Set(ctx.selections.agents) : undefined
    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Copy agent files to .claude/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      const fileId = path.parse(file).name as AgentId
      if (selectedAgents && !selectedAgents.has(fileId)) continue
      await this.copyFileWithRecord(
        path.join(agentsDir, file),
        path.join(claudeDir, file),
        ctx,
      )
    }

    // Skills - exact copy to .claude/commands
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      await this.copyFileWithRecord(
        path.join(skillsDir, file),
        path.join(claudeDir, 'commands', file),
        ctx,
      )
    }

    // Templates - exact copy to .claude/templates
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(templatesDir, file)
      if (files.isDirectory(srcPath)) continue
      await this.copyFileWithRecord(
        srcPath,
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
        const resolution = await resolveConflict(claudeMdPath, 'CLAUDE.md', { force: ctx.force })
        if (resolution !== 'skip') {
          if (resolution === 'backup-and-overwrite') {
            files.backupFile(claudeMdPath, ctx.targetDir)
          }

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
    const relPath = path.relative(ctx.targetDir, dest)
    const resolution = await resolveConflict(dest, relPath, { force: ctx.force })
    if (resolution === 'skip') return
    if (resolution === 'backup-and-overwrite') {
      files.backupFile(dest, ctx.targetDir)
    }

    files.copyFile(src, dest)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
