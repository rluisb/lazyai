import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { resolveConflict } from '../utils/conflicts.js'

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
    const templatesDir = path.join(githubDir, 'templates')
    files.ensureDir(templatesDir)

    console.log('🤖  Installing GitHub Copilot tools...')

    // Copy agent files to .github/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      await this.copyFileWithRecord(
        path.join(agentsDir, file),
        path.join(githubDir, file),
        ctx,
      )
    }

    // Prompts - exact copy
    const promptTemplatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(promptTemplatesDir)) {
      await this.copyFileWithRecord(
        path.join(promptTemplatesDir, file),
        path.join(templatesDir, file),
        ctx,
      )
    }

    // Skills - transformed into Copilot prompt files
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      const src = path.join(skillsDir, file)
      const parsed = path.parse(file)
      const destFile = `${parsed.name}.prompt.md`
      const dest = path.join(promptsDir, destFile)
      await this.copySkillAsPromptWithRecord(src, dest, ctx)
    }

    // Create copilot-instructions.md from template (only if not already created by scaffold)
    const alreadyCreated = ctx.fileRecords.some(r => r.path === '.github/copilot-instructions.md')
    if (!alreadyCreated) {
      const templatePath = path.join(ctx.libraryDir, 'root', 'copilot-instructions.template.md')
      const copilotMdPath = path.join(githubDir, 'copilot-instructions.md')

      if (files.fileExists(templatePath)) {
        const resolution = await resolveConflict(copilotMdPath, '.github/copilot-instructions.md', { force: ctx.force })
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

  private async copySkillAsPromptWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const resolution = await resolveConflict(dest, relPath, { force: ctx.force })
    if (resolution === 'skip') return
    if (resolution === 'backup-and-overwrite') {
      files.backupFile(dest, ctx.targetDir)
    }

    const transformed = `---\nmode: agent\n---\n\n${files.readFile(src)}`
    files.writeFile(dest, transformed)
    ctx.fileRecords.push({
      path: relPath,
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
