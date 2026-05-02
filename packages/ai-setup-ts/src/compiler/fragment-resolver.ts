import * as fs from 'node:fs'
import path from 'node:path'

export interface FragmentContext {
  projectName: string
  planningDir: string
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
  features?: FeatureFlags
  toolDescription?: string
  toolNotes?: string
  testFramework?: string
  packageManager?: string
  testCommand?: string
  lintCommand?: string
  buildCommand?: string
  devCommand?: string
  installCommand?: string
  projectDescription?: string
  constitution?: ConstitutionContext
  fallbacks?: Record<string, string>
}

export interface ConstitutionContext {
  projectOverview?: string | null
  stack?: {
    language?: string | null
    framework?: string | null
    database?: string | null
    orm?: string | null
    testing?: string | null
    packageManager?: string | null
  }
  conventions?: {
    naming?: string | null
    errorHandling?: string | null
    apiResponses?: string | null
    importOrder?: string | null
  }
  commands?: {
    test?: string | null
    lint?: string | null
    build?: string | null
  }
  protectedBranch?: string | null
  coverageThreshold?: number | null
  codebaseMap?: Array<{ path: string; responsibility?: string | null }>
  namingConventions?: string | null
  errorHandling?: string | null
  apiConventions?: string | null
  importOrder?: string | null
  testCommand?: string | null
  lintCommand?: string | null
  buildCommand?: string | null
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
  adversarialDesign?: boolean
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
  adversarial_design?: boolean
}

const DEFAULT_TEMPLATE_FALLBACKS: Record<string, string> = {
  PROJECT_NAME: '[YOUR_PROJECT_NAME]',
  PROJECT_OVERVIEW: '[YOUR_PROJECT_OVERVIEW]',
  LANGUAGE: '[YOUR_LANGUAGE]',
  FRAMEWORK: '[YOUR_FRAMEWORK]',
  DATABASE: '[YOUR_DATABASE]',
  ORM: '[YOUR_ORM]',
  TEST_FRAMEWORK: '[YOUR_TEST_FRAMEWORK]',
  PACKAGE_MANAGER: '[YOUR_PACKAGE_MANAGER]',
  NAMING_CONVENTIONS: '[YOUR_NAMING_CONVENTION]',
  ERROR_HANDLING: '[YOUR_ERROR_PATTERN]',
  API_CONVENTIONS: '[YOUR_API_CONVENTION]',
  IMPORT_ORDER: '[YOUR_IMPORT_ORDER]',
  PROTECTED_BRANCH: '[YOUR_PROTECTED_BRANCH]',
  TEST_COMMAND: '<!-- fill-in: test command -->',
  LINT_COMMAND: '[YOUR_LINT_COMMAND]',
  BUILD_COMMAND: '<!-- fill-in: build command -->',
  DEV_COMMAND: '<!-- fill-in: dev command -->',
  INSTALL_COMMAND: '[YOUR_INSTALL_COMMAND]',
  COVERAGE_THRESHOLD: '80',
  CODEBASE_MAP:
    '| [YOUR_PATH_1] | [WHAT_IT_DOES] |\n| [YOUR_PATH_2] | [WHAT_IT_DOES] |\n| [YOUR_PATH_3] | [WHAT_IT_DOES] |\n| [YOUR_SHARED_PATH] | Shared utilities — check all importers before editing |\n| [YOUR_INFRA_PATH] | Infrastructure — read-only for AI agents |',
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
    
    for (const [index, part] of parts.entries()) {

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
      adversarialDesign: 'adversarial_design',
      adversarial_design: 'adversarialDesign',
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
      const cachedFragment = this.fragmentCache.get(fragmentPath)
      if (cachedFragment !== undefined) {
        return cachedFragment
      }
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
    const constitution = context.constitution
    
    const variables: Record<string, string | undefined> = {
      PROJECT_NAME: context.projectName,
      PLANNING_DIR: context.planningDir,
      PRIMARY_LANGUAGE: context.primaryLanguage || 'TypeScript',
      FRAMEWORK: firstNonEmpty(constitution?.stack?.framework, context.framework),
      WORKSPACE_TYPE: context.workspaceType || 'project',
      TOOL_DESCRIPTION: context.toolDescription || '',
      TOOL_NOTES: context.toolNotes || '',
      PROJECT_INSTRUCTIONS: context.projectInstructions || '',
      TEST_FRAMEWORK: firstNonEmpty(constitution?.stack?.testing, context.testFramework),
      PACKAGE_MANAGER: firstNonEmpty(constitution?.stack?.packageManager, context.packageManager),
      TEST_COMMAND: firstNonEmpty(constitution?.commands?.test, constitution?.testCommand, context.testCommand),
      LINT_COMMAND: firstNonEmpty(constitution?.commands?.lint, constitution?.lintCommand, context.lintCommand),
      BUILD_COMMAND: firstNonEmpty(constitution?.commands?.build, constitution?.buildCommand, context.buildCommand),
      DEV_COMMAND: context.devCommand || '',
      INSTALL_COMMAND: context.installCommand || '',
      PROJECT_DESCRIPTION: context.projectDescription || '',
      PROJECT_OVERVIEW: cleanString(constitution?.projectOverview),
      LANGUAGE: firstNonEmpty(constitution?.stack?.language, context.primaryLanguage),
      DATABASE: cleanString(constitution?.stack?.database),
      ORM: cleanString(constitution?.stack?.orm),
      NAMING_CONVENTIONS: firstNonEmpty(constitution?.conventions?.naming, constitution?.namingConventions),
      ERROR_HANDLING: firstNonEmpty(constitution?.conventions?.errorHandling, constitution?.errorHandling),
      API_CONVENTIONS: firstNonEmpty(constitution?.conventions?.apiResponses, constitution?.apiConventions),
      IMPORT_ORDER: firstNonEmpty(constitution?.conventions?.importOrder, constitution?.importOrder),
      PROTECTED_BRANCH: cleanString(constitution?.protectedBranch),
      COVERAGE_THRESHOLD: constitution?.coverageThreshold != null ? String(constitution.coverageThreshold) : '',
      CODEBASE_MAP: renderCodebaseMap(constitution),
    }
    
    return content.replace(variableRegex, (_, varName) => {
      const value = variables[varName]
      if (value != null && value !== '') return value
      return context.fallbacks?.[varName] ?? DEFAULT_TEMPLATE_FALLBACKS[varName] ?? `{{${varName}}}`
    })
  }
}

function cleanString(value: string | null | undefined): string {
  return value?.trim() ?? ''
}

function firstNonEmpty(...values: Array<string | null | undefined>): string {
  for (const value of values) {
    const cleaned = cleanString(value)
    if (cleaned !== '') return cleaned
  }
  return ''
}

function renderCodebaseMap(constitution: ConstitutionContext | undefined): string {
  if (!constitution?.codebaseMap || constitution.codebaseMap.length === 0) return ''

  return constitution.codebaseMap
    .map((entry) => ({
      path: entry.path.trim().replaceAll('\\', '/'),
      responsibility: firstNonEmpty(entry.responsibility) || '[WHAT_IT_DOES]',
    }))
    .filter((entry) => entry.path !== '' && !isIgnoredCodebasePath(entry.path))
    .map((entry) => `| ${entry.path} | ${entry.responsibility} |`)
    .join('\n')
}

function isIgnoredCodebasePath(pathName: string): boolean {
  const ignored = new Set(['node_modules', 'dist', '.git', 'vendor'])
  return pathName.split('/').some((part) => ignored.has(part))
}
