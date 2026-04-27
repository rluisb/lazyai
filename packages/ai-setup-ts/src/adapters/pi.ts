import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

/**
 * Adapter for Pi Coding Agent (@mariozechner/pi-coding-agent)
 *
 * Structure:
 * - Root: AGENTS.md (or CLAUDE.md)
 * - Config: .pi/settings.json
 * - Skills: .pi/skills/{name}/SKILL.md (AgentSkills standard)
 * - Prompts: .pi/prompts/{name}.md (prompt templates)
 * - No agents concept (agents are inline in AGENTS.md)
 */
export class PiAdapter implements ToolAdapter {
  getToolId(): string {
    return 'pi'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const piDir = path.join(ctx.targetDir, '.pi')
    files.ensureDir(piDir)
    files.ensureDir(path.join(piDir, 'skills'))
    files.ensureDir(path.join(piDir, 'prompts'))

    console.log('🤖  Installing Pi tools...')

    const settingsPath = path.join(piDir, 'settings.json')
    if (!files.fileExists(settingsPath)) {
      const defaultSettings = {
        compaction: {
          enabled: true,
        },
      }
      files.writeFile(settingsPath, JSON.stringify(defaultSettings, null, 2))
      ctx.fileRecords.push({
        path: '.pi/settings.json',
        hash: files.fileHash(settingsPath),
        source: 'generated',
        owner: 'library',
      })
    }

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(piDir, 'skills', name, 'SKILL.md')
      },
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'prompts',
      selectionKey: 'prompts',
      toDestPath: (file) => path.join(piDir, 'prompts', file),
    })

    await installRootTemplateIfMissing({
      ctx,
      recordPath: 'AGENTS.md',
      destPath: path.join(ctx.targetDir, 'AGENTS.md'),
      templateSource: 'root/AGENTS.template.md',
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.pi')
    console.log('🗑️  Removing Pi tools...')
  }
}
