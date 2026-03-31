import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'

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

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      this.copyFileWithRecord(path.join(agentsDir, file), path.join(ocDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      this.copyFileWithRecord(path.join(templatesDir, file), path.join(ocDir, 'templates', file), ctx)
    }

    // Skills - exact copy to commands
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      this.copyFileWithRecord(path.join(skillsDir, file), path.join(ocDir, 'commands', file), ctx)
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const ocDir = path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }

  private copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): void {
    if (files.fileExists(dest)) {
      console.warn(`⚠️  Skipping existing file: ${path.relative(ctx.targetDir, dest)}`)
      return
    }
    files.copyFile(src, dest)
    ctx.fileRecords.push({
      path: path.relative(ctx.targetDir, dest),
      hash: files.fileHash(dest),
      source: path.relative(ctx.libraryDir, src),
    })
  }
}
