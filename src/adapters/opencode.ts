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

    console.log('🤖  Installing OpenCode tools...')

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      this.copyFileWithRecord(path.join(agentsDir, file), path.join(ocDir, 'agents', file), ctx)
    }

    // Note: skills would be transformed here when we add skills to library
    // For now, leaving the commands dir empty as per current MVP state
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const ocDir = path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
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
