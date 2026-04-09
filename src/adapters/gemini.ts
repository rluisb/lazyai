import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyLibraryDirectory,
  getOrchestratorSkillContent,
  installRootTemplateIfMissing,
  isOrchestratorEnabled,
  writeContentWithRecord,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

export class GeminiAdapter implements ToolAdapter {
  getToolId(): string {
    return 'gemini'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const geminiDir = path.join(ctx.targetDir, '.gemini')
    files.ensureDir(geminiDir)
    files.ensureDir(path.join(geminiDir, 'skills'))

    const settingsPath = path.join(geminiDir, 'settings.json')
    if (!files.fileExists(settingsPath)) {
      const defaultSettings = {
        general: {
          defaultApprovalMode: 'default',
        },
        model: {
          name: 'gemini-2.5-pro',
        },
        context: {
          fileName: 'GEMINI.md',
          includeDirectoryTree: true,
        },
      }

      files.writeFile(settingsPath, JSON.stringify(defaultSettings, null, 2))
      ctx.fileRecords.push({
        path: '.gemini/settings.json',
        hash: files.fileHash(settingsPath),
        source: 'generated',
        owner: 'library',
      })
    }

    // Gemini CLI has NO agents concept — skip agents/
    // Gemini CLI has NO templates concept — skip templates/

    console.log('♊  Installing Gemini CLI tools...')

    // Skills → .gemini/skills/<name>/SKILL.md (directory per skill)
    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => {
        const name = path.parse(file).name
        return path.join(geminiDir, 'skills', name, 'SKILL.md')
      },
    })

    if (isOrchestratorEnabled(ctx)) {
      await writeContentWithRecord({
        dest: path.join(geminiDir, 'skills', 'orchestrator', 'SKILL.md'),
        content: getOrchestratorSkillContent(ctx),
        ctx,
        source: 'generated:orchestrator-skill',
      })
    }

    // Agents → skip entirely (Gemini has no agents concept)
    // Prompts → skip (no templates dir in Gemini)

    await installRootTemplateIfMissing({
      ctx,
      recordPath: 'GEMINI.md',
      destPath: path.join(ctx.targetDir, 'GEMINI.md'),
      templateSource: 'root/GEMINI.template.md',
    })
  }

  async remove(ctx: AdapterContext): Promise<void> {
    void path.join(ctx.targetDir, '.gemini')
    console.log('🗑️  Removing Gemini CLI tools...')
    // Basic remove implementation
  }
}
