import * as fs from 'node:fs'
import path from 'node:path'
import type { ToolId } from '../types.js'
import { type FragmentContext, FragmentResolver } from './fragment-resolver.js'

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
    
    if (!fs.existsSync(toolTemplateDir)) {
      throw new Error(`Tool template directory not found: ${toolTemplateDir}`)
    }

    const files: CompiledOutput['files'] = []
    
    // Find all template files
    const templateFiles = this.findTemplateFiles(toolTemplateDir)
    
    for (const templateFile of templateFiles) {
      const content = fs.readFileSync(templateFile, 'utf-8')
      const resolved = this.resolver.resolve(content, this.config.context)
      
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
