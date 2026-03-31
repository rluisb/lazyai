import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import type { AgentId, SkillId, PromptId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'

export class GeminiAdapter implements ToolAdapter {
  getToolId(): string {
    return 'gemini'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const geminiDir = path.join(ctx.targetDir, '.gemini')
    files.ensureDir(geminiDir)
    files.ensureDir(path.join(geminiDir, 'skills'))
    files.ensureDir(path.join(geminiDir, 'templates'))

    console.log('♊  Installing Gemini CLI tools...')

    const selectedAgents = ctx.selections?.agents ? new Set(ctx.selections.agents) : undefined
    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Copy agent files to .gemini/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      const fileId = path.parse(file).name as AgentId
      if (selectedAgents && !selectedAgents.has(fileId)) continue
      await this.copyFileWithRecord(
        path.join(agentsDir, file),
        path.join(geminiDir, file),
        ctx,
      )
    }

    // Skills - exact copy
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      await this.copyFileWithRecord(
        path.join(skillsDir, file),
        path.join(geminiDir, 'skills', file),
        ctx,
      )
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(templatesDir, file)
      if (files.isDirectory(srcPath)) continue
      await this.copyFileWithRecord(
        srcPath,
        path.join(geminiDir, 'templates', file),
        ctx,
      )
    }

    // Create GEMINI.md from template (only if not already created by scaffold)
    const alreadyCreated = ctx.fileRecords.some(r => r.path === 'GEMINI.md')
    if (!alreadyCreated) {
      const templatePath = path.join(ctx.libraryDir, 'root', 'GEMINI.template.md')
      const geminiMdPath = path.join(ctx.targetDir, 'GEMINI.md')

      if (files.fileExists(templatePath)) {
        const resolution = await resolveConflict(geminiMdPath, 'GEMINI.md', { force: ctx.force })
        if (resolution !== 'skip') {
          if (resolution === 'backup-and-overwrite') {
            files.backupFile(geminiMdPath, ctx.targetDir)
          }

          const content = files.readFile(templatePath)
          files.writeFile(geminiMdPath, content)
          ctx.fileRecords.push({
            path: 'GEMINI.md',
            hash: files.fileHash(geminiMdPath),
            source: 'root/GEMINI.template.md',
          })
        }
      }
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const geminiDir = path.join(ctx.targetDir, '.gemini')
    console.log('🗑️  Removing Gemini CLI tools...')
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
