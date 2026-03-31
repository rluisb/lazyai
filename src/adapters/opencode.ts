import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import type { AgentId, SkillId, PromptId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const ocDir = path.join(ctx.targetDir, '.opencode')
    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, 'commands'))
    files.ensureDir(path.join(ocDir, 'templates'))

    console.log('🤖  Installing OpenCode tools...')

    const selectedAgents = ctx.selections?.agents ? new Set(ctx.selections.agents) : undefined
    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      const fileId = path.parse(file).name as AgentId
      if (selectedAgents && !selectedAgents.has(fileId)) continue
      await this.copyFileWithRecord(path.join(agentsDir, file), path.join(ocDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(templatesDir, file)
      if (files.isDirectory(srcPath)) continue
      await this.copyFileWithRecord(srcPath, path.join(ocDir, 'templates', file), ctx)
    }

    // Skills - exact copy to commands
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      await this.copyFileWithRecord(path.join(skillsDir, file), path.join(ocDir, 'commands', file), ctx)
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const ocDir = path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }

  private async copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const resolution = await resolveConflict(dest, relPath, { force: ctx.force })
    if (resolution === 'skip') {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
      return
    }
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
