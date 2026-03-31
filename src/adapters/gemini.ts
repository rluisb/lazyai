import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { confirmReplace } from '../utils/conflicts.js'

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

    // Copy agent files to .gemini/
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      await this.copyFileWithRecord(
        path.join(agentsDir, file),
        path.join(geminiDir, file),
        ctx,
      )
    }

    // Skills - exact copy
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      await this.copyFileWithRecord(
        path.join(skillsDir, file),
        path.join(geminiDir, 'skills', file),
        ctx,
      )
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      await this.copyFileWithRecord(
        path.join(templatesDir, file),
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
        const shouldWrite = await confirmReplace(geminiMdPath, 'GEMINI.md')
        if (shouldWrite) {
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
