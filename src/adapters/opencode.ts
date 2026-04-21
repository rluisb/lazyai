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

export class OpenCodeAdapter implements ToolAdapter {
  getToolId(): string {
    return 'opencode'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const isGlobal = ctx.setupScope === 'global'
    const ocDir = isGlobal ? ctx.targetDir : path.join(ctx.targetDir, '.opencode')
    const skillsDir = 'skills'
    const commandsDir = 'commands'

    files.ensureDir(ocDir)
    files.ensureDir(path.join(ocDir, 'agents'))
    files.ensureDir(path.join(ocDir, skillsDir))
    files.ensureDir(path.join(ocDir, commandsDir))

    console.log('🤖  Installing OpenCode tools...')

    if (!isGlobal) {
      const configPath = path.join(ctx.targetDir, 'opencode.json')
      const jsoncConfigPath = path.join(ctx.targetDir, 'opencode.jsonc')
      if (!files.fileExists(configPath) && !files.fileExists(jsoncConfigPath)) {
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
          path: 'opencode.json',
          hash: files.fileHash(configPath),
          source: 'generated',
          owner: 'library',
        })
      }
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

  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.opencode')
    console.log('🗑️  Removing OpenCode tools...')
    // Basic remove implementation - in a real scenario we'd use fs.rmSync(ocDir, { recursive: true, force: true })
  }
}
