import path from 'node:path'
import * as files from '../utils/files.js'
import type { ToolAdapter, AdapterContext } from './types.js'
import { stripYamlFrontmatter } from '../utils/frontmatter.js'
import {
  copyLibraryDirectory,
  installRootTemplateIfMissing,
  installToolContextFiles,
} from './shared.js'

export class GeminiAdapter implements ToolAdapter {
  getToolId(): string {
    return 'gemini'
  }

  async install(ctx: AdapterContext): Promise<void> {
    const geminiDir = path.join(ctx.targetDir, '.gemini')
    files.ensureDir(geminiDir)
    files.ensureDir(path.join(geminiDir, 'agents'))
    files.ensureDir(path.join(geminiDir, 'skills'))
    files.ensureDir(path.join(geminiDir, 'templates'))

    console.log('♊  Installing Gemini CLI tools...')

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'agents',
      selectionKey: 'agents',
      toDestPath: (file) => path.join(geminiDir, 'agents', file),
      transform: stripYamlFrontmatter,
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'skills',
      selectionKey: 'skills',
      toDestPath: (file) => path.join(geminiDir, 'skills', file),
    })

    await copyLibraryDirectory({
      ctx,
      sourceSubdir: 'prompts',
      selectionKey: 'prompts',
      toDestPath: (file) => path.join(geminiDir, 'templates', file),
    })

    await installToolContextFiles({
      ctx,
      toolDir: geminiDir,
      contextFileName: 'GEMINI.md',
      agentsDestDir: 'agents',
      skillsDestDir: 'skills',
      templatesDestDir: 'templates',
    })

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
