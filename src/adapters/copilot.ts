import path from 'node:path'
import type { PromptId, SkillId } from '../types.js'
import { resolveConflict } from '../utils/conflicts.js'
import * as files from '../utils/files.js'
import { ensureModeAgentFrontmatter, stripYamlFrontmatter } from '../utils/frontmatter.js'
import { getOrchestratorPromptContent, installRootTemplateIfMissing, isOrchestratorEnabled, writeContentWithRecord } from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class CopilotAdapter implements ToolAdapter {
  getToolId(): string {
    return 'copilot'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const githubDir = path.join(ctx.targetDir, '.github')
    files.ensureDir(githubDir)

    const instructionsDir = path.join(githubDir, 'instructions')
    files.ensureDir(instructionsDir)
    const promptsDir = path.join(githubDir, 'prompts')
    files.ensureDir(promptsDir)

    console.log('🤖  Installing GitHub Copilot tools...')

    const selectedSkills = ctx.selections?.skills ? new Set(ctx.selections.skills) : undefined
    const selectedPrompts = ctx.selections?.prompts ? new Set(ctx.selections.prompts) : undefined

    // Prompts - Copilot prompt files use the .prompt.md suffix
    const promptTemplatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(promptTemplatesDir)) {
      const fileId = path.parse(file).name as PromptId
      if (selectedPrompts && !selectedPrompts.has(fileId)) continue
      const srcPath = path.join(promptTemplatesDir, file)
      if (files.isDirectory(srcPath)) continue
      const destFile = `${path.parse(file).name}.prompt.md`
      await this.copyFileWithRecord(
        srcPath,
        path.join(promptsDir, destFile),
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

    if (isOrchestratorEnabled(ctx)) {
      await writeContentWithRecord({
        dest: path.join(promptsDir, 'orchestrator.prompt.md'),
        content: getOrchestratorPromptContent(),
        ctx,
        source: 'generated:orchestrator-prompt',
      })
    }

    await installRootTemplateIfMissing({
      ctx,
      recordPath: 'AGENTS.md',
      destPath: path.join(ctx.targetDir, 'AGENTS.md'),
      templateSource: 'root/AGENTS.template.md',
    })

    await installRootTemplateIfMissing({
      ctx,
      recordPath: '.github/copilot-instructions.md',
      destPath: path.join(githubDir, 'copilot-instructions.md'),
      templateSource: 'root/copilot-instructions.template.md',
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const _githubDir = path.join(ctx.targetDir, '.github')
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
      owner: 'library',
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
      owner: 'library',
    })
  }
}
