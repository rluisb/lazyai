import * as fs from 'node:fs'
import path from 'node:path'
import type { ToolId } from '../types.js'
import { type FragmentContext, FragmentResolver } from './fragment-resolver.js'

const SHARED_ROOT_TEMPLATE_PATH = ['tool-templates', 'shared', 'root.template.md'] as const

const TOOL_OVERRIDES: Partial<Record<ToolId, { description: string; notes: string }>> = {
  opencode: {
    description: 'This project uses OpenCode with ai-setup integration.',
    notes: '',
  },
  'claude-code': {
    description: 'This project uses Claude Code with ai-setup integration.',
    notes: '',
  },
  copilot: {
    description: 'This project uses GitHub Copilot with ai-setup integration.',
    notes: '## Copilot-Specific Notes\n\n- Agents are in `.github/agents/*.md`\n- Prompts are in `.github/prompts/*.prompt.md`',
  },
  gemini: {
    description: 'This project uses Gemini CLI with ai-setup integration.',
    notes: '## Gemini-Specific Notes\n\n- Gemini does not have a separate agents concept\n- Skills are in `.gemini/skills/*/SKILL.md` and function as pseudo-agents',
  },
  pi: {
    description: 'This project uses Pi Coding Agent with ai-setup integration.',
    notes: '## Pi-Specific Notes\n\n- Agents are in `.pi/agents/*.md`\n- Templates are in `.pi/templates/*.md`\n- Skills are in `.pi/skills/*.md`',
  },
  codex: {
    description: 'This project uses OpenAI Codex CLI with ai-setup integration.',
    notes: '## Codex-Specific Notes\n\n- Agents are defined inline in this file (no separate agents directory)\n- Skills are in `.codex/skills/*/SKILL.md`',
  },
}

export interface CompilerConfig {
  libraryDir: string
  outputDir: string
  tool: ToolId
  context: FragmentContext
}

export interface CompiledOutput {
  tool: ToolId
  files: Array<{
    relativePath: string
    content: string
  }>
}

/**
 * Compiles templates for specific tools by resolving fragments
 * and generating tool-native output files.
 */
export class TemplateCompiler {
  private resolver: FragmentResolver
  
  constructor(private config: CompilerConfig) {
    this.resolver = new FragmentResolver(config.libraryDir)
  }

  /**
   * Compile all templates for the configured tool
   */
  compile(): CompiledOutput {
    const toolTemplateDir = path.join(this.config.libraryDir, 'tool-templates', this.config.tool)
    const sharedRootTemplate = path.join(this.config.libraryDir, ...SHARED_ROOT_TEMPLATE_PATH)
    const hasSharedRootTemplate = fs.existsSync(sharedRootTemplate)
    const hasToolTemplateDir = fs.existsSync(toolTemplateDir)

    if (!hasSharedRootTemplate && !hasToolTemplateDir) {
      throw new Error(`Tool template directory not found: ${toolTemplateDir}`)
    }

    const files: CompiledOutput['files'] = []
    const context = this.getContextWithToolOverrides()

    if (hasSharedRootTemplate) {
      const content = fs.readFileSync(sharedRootTemplate, 'utf-8')
      const resolved = this.resolver.resolve(content, context)
      files.push({ relativePath: 'root.md', content: resolved })
    }

    const templateFiles = hasToolTemplateDir
      ? this.findTemplateFiles(toolTemplateDir).filter((templateFile) => {
          if (!hasSharedRootTemplate) return true
          return !path.basename(templateFile).startsWith('root.template.')
        })
      : []

    for (const templateFile of templateFiles) {
      const content = fs.readFileSync(templateFile, 'utf-8')
      const resolved = this.resolver.resolve(content, context)

      // Convert template path to output path
      const relativePath = path.relative(toolTemplateDir, templateFile)
        .replace('.template.md', '.md')
        .replace('.template.', '.')
      
      files.push({ relativePath, content: resolved })
    }
    
    return { tool: this.config.tool, files }
  }

  /**
   * Compile and write output files to disk
   */
  compileAndWrite(): void {
    const output = this.compile()
    
    for (const file of output.files) {
      const fullPath = path.join(this.config.outputDir, file.relativePath)
      const dir = path.dirname(fullPath)
      
      if (!fs.existsSync(dir)) {
        fs.mkdirSync(dir, { recursive: true })
      }
      
      fs.writeFileSync(fullPath, file.content, 'utf-8')
      console.log(`  ✓ ${file.relativePath}`)
    }
  }

  private findTemplateFiles(dir: string): string[] {
    const files: string[] = []
    
    const entries = fs.readdirSync(dir, { withFileTypes: true })
    
    for (const entry of entries) {
      const fullPath = path.join(dir, entry.name)
      
      if (entry.isDirectory()) {
        files.push(...this.findTemplateFiles(fullPath))
      } else if (entry.name.includes('.template.')) {
        files.push(fullPath)
      }
    }
    
    return files
  }

  private getContextWithToolOverrides(): FragmentContext {
    const overrides = TOOL_OVERRIDES[this.config.tool] ?? { description: '', notes: '' }

    return {
      ...this.config.context,
      toolDescription: overrides.description,
      toolNotes: overrides.notes,
    }
  }
}

/**
 * Compile templates for multiple tools
 */
export function compileForTools(
  tools: ToolId[],
  libraryDir: string,
  outputDir: string,
  context: FragmentContext
): Map<ToolId, CompiledOutput> {
  const results = new Map<ToolId, CompiledOutput>()
  
  for (const tool of tools) {
    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir,
      tool,
      context,
    })
    
    results.set(tool, compiler.compile())
  }
  
  return results
}
