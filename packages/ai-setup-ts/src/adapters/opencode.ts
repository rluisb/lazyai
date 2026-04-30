import fs from 'node:fs'
import { execFileSync } from 'node:child_process'
import path from 'node:path'
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

function installOpenCodePlugins(ctx: AdapterContext): void {
  const plugins = ctx.selections?.opencodePlugins ?? []
  if (plugins.length === 0) return

  try {
    execFileSync('opencode', ['--version'], { stdio: 'ignore' })
  } catch {
    console.warn('⚠️  OpenCode CLI not found; skipping plugin install')
    return
  }

  for (const pluginModule of plugins) {
    const args =
      ctx.setupScope === 'global' ? ['plugin', '-g', pluginModule] : ['plugin', pluginModule]
    try {
      execFileSync('opencode', args, { stdio: 'ignore' })
    } catch {
      console.warn(`⚠️  Failed to install OpenCode plugin ${pluginModule}; continuing`)
    }
  }
}

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const ocDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.opencode')
    const legacyConfigPath = path.join(ocDir, 'opencode.json')
    const configPath = path.join(ocDir, 'opencode.jsonc')
    const skillsDir = 'skills'
    const commandsDir = 'commands'

    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, skillsDir))
    files.ensureDir(path.join(ocDir, commandsDir))

    console.log('🤖  Installing OpenCode tools...')

    if (files.fileExists(legacyConfigPath)) {
      const backupPath = `${legacyConfigPath}.bak`
      if (!files.fileExists(backupPath)) {
        files.copyFile(legacyConfigPath, backupPath)
      }
      if (!files.fileExists(configPath)) {
        files.copyFile(legacyConfigPath, configPath)
      }
      fs.rmSync(legacyConfigPath, { force: true })
    }

    if (!files.fileExists(configPath)) {
      const defaultConfig = {
        $schema: 'https://opencode.ai/config.json',
        instructions: ['AGENTS.md'],
        permission: {
          edit: 'ask',
          bash: 'ask',
        },
      }
      files.writeFile(configPath, JSON.stringify(defaultConfig, null, 2))
      ctx.fileRecords.push({
        path: path.relative(ctx.targetDir, configPath),
        hash: files.fileHash(configPath),
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

    installOpenCodePlugins(ctx)
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }
}
