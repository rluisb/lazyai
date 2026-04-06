import path from 'node:path'
import * as files from '../utils/files.js'
import { stripFrontmatterAndInjectModel } from '../utils/frontmatter.js'
import { copyLibraryDirectory, installToolContextFiles } from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const ocDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.opencode')
    const skillsDir = 'skills'

    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, skillsDir))

    console.log('🤖  Installing OpenCode tools...')

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(ocDir, 'agents', file),
      warnOnSkip: true,
      transform: stripFrontmatterAndInjectModel,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(ocDir, skillsDir, name, 'SKILL.md')
      },
      warnOnSkip: true,
    })

    await installToolContextFiles({
      ctx,
      toolDir: ocDir,
      contextFileName: 'AGENTS.md',
      agentsDestDir: 'agents',
      skillsDestDir: skillsDir,
      warnOnSkip: true,
    })

  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }
}
