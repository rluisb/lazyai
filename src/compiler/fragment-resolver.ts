import path from 'node:path'
import * as fs from 'node:fs'

export interface FragmentContext {
  projectName: string
  planningDir: string
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
  features?: {
    tree_of_thoughts?: boolean
    agent_harness?: boolean
    bug_resolution?: boolean
  }
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
    
    for (const part of parts) {
      if (value && typeof value === 'object' && part in value) {
        value = (value as Record<string, unknown>)[part]
      } else {
        return false
      }
    }
    
    return Boolean(value)
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
      PROJECT_INSTRUCTIONS: context.projectInstructions || '',
    }
    
    return content.replace(variableRegex, (_, varName) => {
      return variables[varName] ?? `{{${varName}}}`
    })
  }
}
