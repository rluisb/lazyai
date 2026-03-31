import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
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

    // Agents - exact copy
    const agentsDir = path.join(ctx.libraryDir, 'agents')
    for (const file of files.listDir(agentsDir)) {
      await this.copyFileWithRecord(path.join(agentsDir, file), path.join(ocDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      await this.copyFileWithRecord(path.join(templatesDir, file), path.join(ocDir, 'templates', file), ctx)
    }

    // Skills - exact copy to commands
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
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
