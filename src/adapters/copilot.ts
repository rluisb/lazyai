import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { confirmReplace } from '../utils/conflicts.js'

export class CopilotAdapter implements ToolAdapter {
  getToolId(): string {
    return 'copilot'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const githubDir = path.join(ctx.targetDir, '.github')
    files.ensureDir(githubDir)

    const instructionsDir = path.join(githubDir, 'instructions')
    files.ensureDir(instructionsDir)

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

    // Create copilot-instructions.md from template (only if not already created by scaffold)
    const alreadyCreated = ctx.fileRecords.some(r => r.path === '.github/copilot-instructions.md')
    if (!alreadyCreated) {
      const templatePath = path.join(ctx.libraryDir, 'root', 'copilot-instructions.template.md')
      const copilotMdPath = path.join(githubDir, 'copilot-instructions.md')

      if (files.fileExists(templatePath)) {
        const shouldWrite = await confirmReplace(copilotMdPath, '.github/copilot-instructions.md')
        if (shouldWrite) {
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
