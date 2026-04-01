import path from 'node:path'
import * as files from '../utils/files.js'
import { backupFile } from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import type { AgentId, SkillId, PromptId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'
import { stripYamlFrontmatter } from '../utils/frontmatter.js'

export class PiAdapter implements ToolAdapter {
  getToolId(): string {
    return 'pi'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const piDir = path.join(ctx.targetDir, '.pi')
    files.ensureDir(piDir)
    files.ensureDir(path.join(piDir, 'agents'))
    files.ensureDir(path.join(piDir, 'templates'))
    files.ensureDir(path.join(piDir, 'skills'))

    console.log('🤖  Installing Pi (Claude Code) tools...')

    const selectedAgents = ctx.selections?.agents ? new Set(ctx.selections.agents) : undefined
    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      const fileId = path.parse(file).name as AgentId
      if (selectedAgents && !selectedAgents.has(fileId)) continue
      await this.copyAgentFileWithRecord(path.join(agentsDir, file), path.join(piDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(templatesDir, file)
      if (files.isDirectory(srcPath)) continue
      await this.copyFileWithRecord(srcPath, path.join(piDir, 'templates', file), ctx)
    }

    // Skills - exact copy
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      await this.copyFileWithRecord(path.join(skillsDir, file), path.join(piDir, 'skills', file), ctx)
    }

    // Install tool-agents context files
    const toolAgentsDir = path.join(ctx.libraryDir, 'tool-agents')
    const contextFileName = 'INSTRUCTIONS.md'

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'agents-dir.md'),
      path.join(piDir, 'agents', contextFileName),
      ctx,
    )

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'skills-dir.md'),
      path.join(piDir, 'skills', contextFileName),
      ctx,
    )

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'templates-dir.md'),
      path.join(piDir, 'templates', contextFileName),
      ctx,
    )

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'root-dir.md'),
      path.join(piDir, contextFileName),
      ctx,
    )
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const piDir = path.join(ctx.targetDir, '.pi')
    console.log('🗑️  Removing Pi tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(piDir, { recursive: true, force: true })
  }

  private async copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const effectiveStrategy = ctx.perFileOverrides?.get(dest) ?? ctx.strategy
    const resolution = await resolveConflict(dest, relPath, {
      force: ctx.force,
      ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
    })

    if (resolution === 'skip') {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
      return
    }

    if (resolution === 'backup-and-overwrite') {
      backupFile(dest, ctx.targetDir)
    }

    files.copyFile(src, dest)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }

  private async copyAgentFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const effectiveStrategy = ctx.perFileOverrides?.get(dest) ?? ctx.strategy
    const resolution = await resolveConflict(dest, relPath, {
      force: ctx.force,
      ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
    })

    if (resolution === 'skip') {
      console.warn(`⚠️  Skipping existing file: ${relPath}`)
      return
    }

    if (resolution === 'backup-and-overwrite') {
      backupFile(dest, ctx.targetDir)
    }

    const transformed = stripYamlFrontmatter(files.readFile(src))
    files.writeFile(dest, transformed)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
