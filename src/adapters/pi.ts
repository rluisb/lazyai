import path from 'node:path'
import * as files from '../utils/files.js'
import { stripYamlFrontmatter } from '../utils/frontmatter.js'
import { copyLibraryDirectory, installToolContextFiles } from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

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

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(piDir, 'agents', file),
      warnOnSkip: true,
      transform: stripYamlFrontmatter,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'prompts',
      selectionKey: 'prompts',
      toDestPath: (file) => path.join(piDir, 'templates', file),
      warnOnSkip: true,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => path.join(piDir, 'skills', file),
      warnOnSkip: true,
    })

    await installToolContextFiles({
      ctx,
      toolDir: piDir,
      contextFileName: 'INSTRUCTIONS.md',
      agentsDestDir: 'agents',
      skillsDestDir: 'skills',
      templatesDestDir: 'templates',
      warnOnSkip: true,
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.pi')
    console.log('🗑️  Removing Pi tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(piDir, { recursive: true, force: true })
  }
}
