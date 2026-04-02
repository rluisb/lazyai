import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { stripYamlFrontmatter } from '../utils/frontmatter.js'
import { copyLibraryDirectory, installToolContextFiles } from './shared.js'

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const ocDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.opencode')
    const commandsDir = isGlobal ? 'command' : 'commands'
    const skillsDir = isGlobal ? 'skill' : 'skills'

    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, commandsDir))
    files.ensureDir(path.join(ocDir, skillsDir))
    files.ensureDir(path.join(ocDir, 'templates'))

    console.log('🤖  Installing OpenCode tools...')

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(ocDir, 'agents', file),
      warnOnSkip: true,
      transform: stripYamlFrontmatter,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'prompts',
      selectionKey: 'prompts',
      toDestPath: (file) => path.join(ocDir, 'templates', file),
      warnOnSkip: true,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => path.join(ocDir, commandsDir, file),
      warnOnSkip: true,
    })

    await installToolContextFiles({
      ctx,
      toolDir: ocDir,
      contextFileName: 'AGENTS.md',
      agentsDestDir: 'agents',
      skillsDestDir: commandsDir,
      templatesDestDir: 'templates',
      warnOnSkip: true,
    })

  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }
}
