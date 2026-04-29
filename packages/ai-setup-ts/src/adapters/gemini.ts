import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyWithRecord,
  copyLibraryDirectory,
  getOrchestratorSkillContent,
  installRootTemplateIfMissing,
  isOrchestratorEnabled,
  writeContentWithRecord,
} from './shared.js'
import type { AdapterContext, ToolAdapter } from './types.js'

/**
 * Resolve the preferred Gemini commands directory.
 * Checks library/gemini/commands/ first (spec 017 layout), falls back to
 * library/commands/ (legacy). Returns the library-relative path.
 */
function resolveGeminiCommandsSubdir(libraryDir: string): string | undefined {
  const preferred = path.join(libraryDir, 'gemini', 'commands')
  if (files.isDirectory(preferred) && files.listDir(preferred).length > 0) {
    return path.join('gemini', 'commands')
  }
  const legacy = path.join(libraryDir, 'commands')
  if (files.isDirectory(legacy) && files.listDir(legacy).length > 0) {
    return 'commands'
  }
  return undefined
}

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

    // Copy custom slash commands (TOML files). Prefer library/gemini/commands/,
    // fall back to library/commands/ (legacy).
    const commandsSubdir = resolveGeminiCommandsSubdir(ctx.libraryDir)
    if (commandsSubdir) {
      const commandsDir = path.join(ctx.libraryDir, commandsSubdir)
      files.ensureDir(path.join(geminiDir, 'commands'))
      for (const file of files.listDir(commandsDir)) {
        const srcPath = path.join(commandsDir, file)
        if (files.isDirectory(srcPath)) continue
        await copyWithRecord({
          src: srcPath,
          dest: path.join(geminiDir, 'commands', path.basename(file)),
          ctx,
          warnOnSkip: true,
        })
      }
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
