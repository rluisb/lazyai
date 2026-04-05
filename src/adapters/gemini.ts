import path from 'node:path'
import * as files from '../utils/files.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
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
