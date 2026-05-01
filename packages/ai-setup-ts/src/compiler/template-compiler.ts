import * as fs from 'node:fs'
import path from 'node:path'
import type { ToolId } from '../types.js'
import { type FragmentContext, FragmentResolver } from './fragment-resolver.js'

const SHARED_ROOT_TEMPLATE_PATH = ['tool-templates', 'shared', 'root.template.md'] as const

const TOOL_OVERRIDES: Partial<Record<ToolId, { description: string; notes: string; rootFile?: string }>> = {
  opencode: {
    description: 'This project uses OpenCode with ai-setup integration.',
    notes: '## OpenCode-Specific Notes\n\n- Project config: `opencode.json` at project root\n- Agents: `.opencode/agents/<name>.md`\n- Skills: `.opencode/skills/<name>/SKILL.md`\n- Commands: `.opencode/commands/<name>.md`\n- Multiple config sources merged (project → global → env)',
  },
  'claude-code': {
    description: 'This project uses Claude Code with ai-setup integration.',
    notes: '## Claude Code-Specific Notes\n\n- Project settings: `.claude/settings.json`\n- Modular rules: `.claude/rules/<name>.md` (supports `paths` frontmatter for scoping)\n- Skills: `.claude/skills/<name>/SKILL.md`\n- Agents: `.claude/agents/<name>.md`\n- Personal overrides: `CLAUDE.local.md` (gitignore this)',
  },
  copilot: {
    description: 'This project uses GitHub Copilot with ai-setup integration.',
    rootFile: 'copilot-instructions.md',
    notes: '## Copilot-Specific Notes\n\n- Repository-wide instructions: `.github/copilot-instructions.md`\n- Path-specific instructions: `.github/instructions/<name>.instructions.md` with `applyTo` frontmatter\n- Reusable prompts: `.github/prompts/<name>.prompt.md`\n- Agent instructions: `AGENTS.md` at project root',
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
