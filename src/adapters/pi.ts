import path from 'node:path'
import * as files from '../utils/files.js'
import { backupFile } from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { resolveConflict } from '../utils/conflicts.js'

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
      await this.copyFileWithRecord(path.join(agentsDir, file), path.join(piDir, 'agents', file), ctx)
    }

    // Templates - exact copy
    const templatesDir = path.join(ctx.libraryDir, 'prompts')
    for (const file of files.listDir(templatesDir)) {
      await this.copyFileWithRecord(path.join(templatesDir, file), path.join(piDir, 'templates', file), ctx)
    }

    // Skills - exact copy
    const skillsDir = path.join(ctx.libraryDir, 'skills')
    for (const file of files.listDir(skillsDir)) {
      await this.copyFileWithRecord(path.join(skillsDir, file), path.join(piDir, 'skills', file), ctx)
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    const piDir = path.join(ctx.targetDir, '.pi')
    console.log('🗑️  Removing Pi tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(piDir, { recursive: true, force: true })
  }

  private async copyFileWithRecord(src: string, dest: string, ctx: AdapterContext): Promise<void> {
    const relPath = path.relative(ctx.targetDir, dest)
    const resolution = await resolveConflict(dest, relPath, { force: ctx.force })

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
}
