import * as fs from 'node:fs'
import path from 'node:path'

export interface FragmentContext {
  projectName: string
  planningDir: string
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
  toolDescription?: string
  toolNotes?: string
  features?: FeatureFlags
}

export interface FeatureFlags {
  contextEngineering?: boolean
  rpiWorkflow?: boolean
  chainOfThought?: boolean
  treeOfThoughts?: boolean
  adrEnforcement?: boolean
  qualityGates?: boolean
  agentHarness?: boolean
  bugResolution?: boolean
  pivotHandling?: boolean
  gitConventions?: boolean
  // Legacy aliases kept for backwards compatibility with older templates
  context_engineering?: boolean
  rpi_workflow?: boolean
  chain_of_thought?: boolean
  tree_of_thoughts?: boolean
  adr_enforcement?: boolean
  quality_gates?: boolean
  agent_harness?: boolean
  bug_resolution?: boolean
  pivot_handling?: boolean
  git_conventions?: boolean
}

/**
 * Resolves XML fragments and variable interpolation in templates.
 * 
 * Supports:
 * - {{#include fragments/name.xml}} - Include fragment file
 * - {{VARIABLE_NAME}} - Simple variable substitution
 * - {{#if features.name}}...{{/if}} - Conditional inclusion
 */
export class FragmentResolver {
  private fragmentCache: Map<string, string> = new Map()
  
  constructor(private libraryDir: string) {}

  /**
   * Resolve all fragments and variables in template content
   */
  resolve(content: string, context: FragmentContext): string {
    // First pass: resolve conditionals
    let result = this.resolveConditionals(content, context)
    
    // Second pass: resolve includes
    result = this.resolveIncludes(result)
    
    // Third pass: resolve variables
    result = this.resolveVariables(result, context)
    
    return result
  }

  private resolveConditionals(content: string, context: FragmentContext): string {
    // Match {{#if features.name}}...{{/if}}
    const conditionalRegex = /\{\{#if\s+([\w.]+)\}\}([\s\S]*?)\{\{\/if\}\}/g
    
    return content.replace(conditionalRegex, (_, condition, body) => {
      const value = this.evaluateCondition(condition, context)
      return value ? body : ''
    })
  }

  private evaluateCondition(condition: string, context: FragmentContext): boolean {
    const parts = condition.split('.')
    let value: unknown = context
    
    for (let index = 0; index < parts.length; index += 1) {
      const part = parts[index]!

      if (value && typeof value === 'object' && part in value) {
        value = (value as Record<string, unknown>)[part]
        continue
      }

      // Feature flags support both camelCase and snake_case condition names
      if (parts[0] === 'features' && index === 1 && value && typeof value === 'object') {
        const alias = this.getFeatureAlias(part)
        if (alias && alias in (value as Record<string, unknown>)) {
          value = (value as Record<string, unknown>)[alias]
          continue
        }
      }

      return false
    }
    
    return Boolean(value)
  }

  private getFeatureAlias(name: string): string | undefined {
    const aliases: Record<string, string> = {
      contextEngineering: 'context_engineering',
      context_engineering: 'contextEngineering',
      rpiWorkflow: 'rpi_workflow',
      rpi_workflow: 'rpiWorkflow',
      chainOfThought: 'chain_of_thought',
      chain_of_thought: 'chainOfThought',
      treeOfThoughts: 'tree_of_thoughts',
      tree_of_thoughts: 'treeOfThoughts',
      adrEnforcement: 'adr_enforcement',
      adr_enforcement: 'adrEnforcement',
      qualityGates: 'quality_gates',
      quality_gates: 'qualityGates',
      agentHarness: 'agent_harness',
      agent_harness: 'agentHarness',
      bugResolution: 'bug_resolution',
      bug_resolution: 'bugResolution',
      pivotHandling: 'pivot_handling',
      pivot_handling: 'pivotHandling',
      gitConventions: 'git_conventions',
      git_conventions: 'gitConventions',
    }

    return aliases[name]
  }

  private resolveIncludes(content: string): string {
    // Match {{#include fragments/name.xml}}
    const includeRegex = /\{\{#include\s+([\w/.-]+)\}\}/g
    
    return content.replace(includeRegex, (_, fragmentPath) => {
      return this.loadFragment(fragmentPath)
    })
  }

  private loadFragment(fragmentPath: string): string {
    if (this.fragmentCache.has(fragmentPath)) {
      return this.fragmentCache.get(fragmentPath)!
    }

    const fullPath = path.join(this.libraryDir, fragmentPath)
    
    if (!fs.existsSync(fullPath)) {
      console.warn(`Fragment not found: ${fragmentPath}`)
      return `<!-- Fragment not found: ${fragmentPath} -->`
    }

    const content = fs.readFileSync(fullPath, 'utf-8')
    this.fragmentCache.set(fragmentPath, content)
    return content
  }

  private resolveVariables(content: string, context: FragmentContext): string {
    // Match {{VARIABLE_NAME}}
    const variableRegex = /\{\{(\w+)\}\}/g
    
    const variables: Record<string, string> = {
      PROJECT_NAME: context.projectName,
      PLANNING_DIR: context.planningDir,
      PRIMARY_LANGUAGE: context.primaryLanguage || 'TypeScript',
      FRAMEWORK: context.framework || '',
      WORKSPACE_TYPE: context.workspaceType || 'project',
      TOOL_DESCRIPTION: context.toolDescription || '',
      TOOL_NOTES: context.toolNotes || '',
      PROJECT_INSTRUCTIONS: context.projectInstructions || '',
    }
    
    return content.replace(variableRegex, (_, varName) => {
      return variables[varName] ?? `{{${varName}}}`
    })
  }
}
