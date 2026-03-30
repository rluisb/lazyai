import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'

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

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      this.copyFileWithRecord(path.join(agentsDir, file), path.join(piDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      this.copyFileWithRecord(path.join(templatesDir, file), path.join(piDir, 'templates', file), ctx)
    }

    // Note: skills would be transformed here when we add skills to library
    // For now, leaving the skills dir empty as per current MVP state
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const piDir = path.join(ctx.targetDir, '.pi')
    console.log('🗑️  Removing Pi tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(piDir, { recursive: true, force: true })
  }

  private copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): void {
    if (files.fileExists(dest)) {
      console.warn(`⚠️  Skipping existing file: ${path.relative(process.cwd(), dest)}`)
      return
    }
    files.copyFile(src, dest)
    ctx.fileRecords.push({
      path: path.relative(process.cwd(), dest),
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
