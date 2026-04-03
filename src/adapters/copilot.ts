import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import type { AgentId, SkillId, PromptId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'
import { ensureModeAgentFrontmatter, stripYamlFrontmatter } from '../utils/frontmatter.js'

export class CopilotAdapter implements ToolAdapter {
  getToolId(): string {
    return 'copilot'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const githubDir = path.join(ctx.targetDir, '.github')
    files.ensureDir(githubDir)

    const instructionsDir = path.join(githubDir, 'instructions')
    files.ensureDir(instructionsDir)
    const agentsContextDir = path.join(githubDir, 'agents')
    files.ensureDir(agentsContextDir)
    const promptsDir = path.join(githubDir, 'prompts')
    files.ensureDir(promptsDir)

    console.log('🤖  Installing GitHub Copilot tools...')

    const selectedAgents = ctx.selections?.agents ? new Set(ctx.selections.agents) : undefined
    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Copy agent files to .github/agents/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      const src = path.join(agentsDir, file)
      if (files.isDirectory(src)) continue
      const fileId = path.parse(file).name as AgentId
      if (selectedAgents && !selectedAgents.has(fileId)) continue
      await this.copyAgentFileWithRecord(
        src,
        path.join(agentsContextDir, file),
        ctx,
      )
    }

    // Prompts - exact copy
    const promptTemplatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(promptTemplatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(promptTemplatesDir, file)
      if (files.isDirectory(srcPath)) continue
      await this.copyFileWithRecord(
        srcPath,
        path.join(promptsDir, file),
        ctx,
      )
    }

    // Skills - transformed into Copilot prompt files
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const src = path.join(skillsDir, file)
      if (files.isDirectory(src)) continue
      const fileId = path.parse(file).name as SkillId
      if (selectedSkills && !selectedSkills.has(fileId)) continue
      const parsed = path.parse(file)
      const destFile = `${parsed.name}.prompt.md`
      const dest = path.join(promptsDir, destFile)
      await this.copySkillAsPromptWithRecord(src, dest, ctx)
    }

    // Install tool-agents context files
    const toolAgentsDir = path.join(ctx.libraryDir, 'tool-agents')
    const contextFileName = 'AGENTS.md'

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'agents-dir.md'),
      path.join(agentsContextDir, contextFileName),
      ctx,
    )

    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'skills-dir.md'),
      path.join(promptsDir, contextFileName),
      ctx,
    )


    await this.copyFileWithRecord(
      path.join(toolAgentsDir, 'root-dir.md'),
      path.join(githubDir, contextFileName),
      ctx,
    )

    // Create copilot-instructions.md from template (only if not already created by scaffold)
    const alreadyCreated = ctx.fileRecords.some(r => r.path === '.github/copilot-instructions.md')
    if (!alreadyCreated) {
      const templatePath = path.join(ctx.libraryDir, 'root', 'copilot-instructions.template.md')
      const copilotMdPath = path.join(githubDir, 'copilot-instructions.md')

      if (files.fileExists(templatePath)) {
        const effectiveStrategy = ctx.perFileOverrides?.get(copilotMdPath) ?? ctx.strategy
        const resolution = await resolveConflict(copilotMdPath, '.github/copilot-instructions.md', {
          force: ctx.force,
          ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
        })
        if (resolution !== 'skip') {
          if (resolution === 'backup-and-overwrite') {
            files.backupFile(copilotMdPath, ctx.targetDir)
          }

          const content = files.readFile(templatePath)
          files.writeFile(copilotMdPath, content)
          ctx.fileRecords.push({
            path: '.github/copilot-instructions.md',
            hash: files.fileHash(copilotMdPath),
            source: 'root/copilot-instructions.template.md',
          })
        }
      }
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const githubDir = path.join(ctx.targetDir, '.github')
    console.log('🗑️  Removing GitHub Copilot tools...')
    // Basic remove implementation
  }

  private async copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const effectiveStrategy = ctx.perFileOverrides?.get(dest) ?? ctx.strategy
    const resolution = await resolveConflict(dest, relPath, {
      force: ctx.force,
      ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
    })
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

  private async copyAgentFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const effectiveStrategy = ctx.perFileOverrides?.get(dest) ?? ctx.strategy
    const resolution = await resolveConflict(dest, relPath, {
      force: ctx.force,
      ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
    })
    if (resolution === 'skip') return
    if (resolution === 'backup-and-overwrite') {
      files.backupFile(dest, ctx.targetDir)
    }

    const transformed = stripYamlFrontmatter(files.readFile(src))
    files.writeFile(dest, transformed)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }

  private async copySkillAsPromptWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const effectiveStrategy = ctx.perFileOverrides?.get(dest) ?? ctx.strategy
    const resolution = await resolveConflict(dest, relPath, {
      force: ctx.force,
      ...(effectiveStrategy ? { strategy: effectiveStrategy } : {}),
    })
    if (resolution === 'skip') return
    if (resolution === 'backup-and-overwrite') {
      files.backupFile(dest, ctx.targetDir)
    }

    const transformed = ensureModeAgentFrontmatter(stripYamlFrontmatter(files.readFile(src)))
    files.writeFile(dest, transformed)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
