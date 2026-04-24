import path from 'node:path'
import { execSync } from 'node:child_process'
import { mergeJsonFile } from '../utils/configmerge.js'
import * as files from '../utils/files.js'
import { stripFrontmatterAndInjectModel } from '../utils/frontmatter.js'
import {
  copyLibraryDirectory,
  copyWithRecord,
  getOrchestratorAgentContent,
  installToolContextFiles,
  isOrchestratorEnabled,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    // ocDir: .opencode/ for project/workspace; ~/.config/opencode for global
    const ocDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.opencode')

    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, 'skills'))
    files.ensureDir(path.join(ocDir, 'commands'))
    files.ensureDir(path.join(ocDir, 'modes'))

    console.log('Installing OpenCode tools...')

    // Config at project root for project/workspace; inside ocDir for global.
    const configRoot = isGlobal ? ocDir : ctx.targetDir
    const jsonPath = path.join(configRoot, 'opencode.json')
    const jsoncPath = path.join(configRoot, 'opencode.jsonc')

    if (!files.fileExists(jsonPath) && !files.fileExists(jsoncPath)) {
      const defaultConfig = {
        $schema: 'https://opencode.ai/config.json',
        instructions: ['AGENTS.md'],
        permission: {
          edit: 'ask',
          bash: 'ask',
        },
      }
      mergeJsonFile(jsoncPath, defaultConfig)
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, jsoncPath),
        hash: files.fileHash(jsoncPath),
        source: 'generated',
        owner: 'library',
      })
    }

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(ocDir, 'agents', file),
      warnOnSkip: true,
      transform: stripFrontmatterAndInjectModel,
      includeFile: (file) => path.parse(file).name !== 'orchestrator',
    })

    if (isOrchestratorEnabled(ctx)) {
      const orchestratorSource = path.join(ctx.libraryDir, 'agents', 'orchestrator.md')
      await copyWithRecord({
        src: orchestratorSource,
        dest: path.join(ocDir, 'agents', 'orchestrator.md'),
        ctx,
        warnOnSkip: true,
        transform: () => stripFrontmatterAndInjectModel(getOrchestratorAgentContent(ctx)),
      })
    }

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(ocDir, 'skills', name, 'SKILL.md')
      },
      warnOnSkip: true,
    })

    // Copy opencode-native slash commands (all, no selection filter)
    await this._copyAllFromLibSubdir(ctx, 'opencode/commands', path.join(ocDir, 'commands'))

    // Copy opencode chat modes (all, no selection filter)
    await this._copyAllFromLibSubdir(ctx, 'opencode/modes', path.join(ocDir, 'modes'))

    await installToolContextFiles({
      ctx,
      toolDir: ocDir,
      contextFileName: 'AGENTS.md',
      agentsDestDir: 'agents',
      skillsDestDir: 'skills',
      warnOnSkip: true,
    })

    this._installPlugins(ctx, isGlobal)
  }

  private async _copyAllFromLibSubdir(
    ctx: AdapterContext,
    subdir: string,
    destDir: string,
  ): Promise<void> {
    const sourceDir = path.join(ctx.libraryDir, subdir)
    if (!files.fileExists(sourceDir)) return

    for (const file of files.listDir(sourceDir)) {
      const srcPath = path.join(sourceDir, file)
      if (files.isDirectory(srcPath)) continue
      await copyWithRecord({
        src: srcPath,
        dest: path.join(destDir, file),
        ctx,
        warnOnSkip: true,
      })
    }
  }

  private _installPlugins(ctx: AdapterContext, isGlobal: boolean): void {
    const plugins: string[] | undefined = (ctx.selections as any)?.opencodePlugins
    if (!plugins || plugins.length === 0) return

    try {
      execSync('which opencode', { stdio: 'ignore' })
    } catch {
      return
    }

    for (const module of plugins) {
      try {
        const args = isGlobal ? `plugin ${module} -g` : `plugin ${module}`
        execSync(`opencode ${args}`, {
          cwd: isGlobal ? undefined : ctx.targetDir,
          stdio: 'pipe',
        })
      } catch (err) {
        console.warn(`Warning: opencode plugin ${module} failed: ${err}`)
      }
    }
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.opencode')
    console.log('Removing OpenCode tools...')
  }
}
