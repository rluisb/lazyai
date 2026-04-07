import path from 'node:path'
import { type FragmentContext, TemplateCompiler } from '../compiler/index.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import type { ConflictStrategy, FileRecord, SetupScope, ToolId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileHash, writeFile } from '../utils/files.js'
import { detectProjectStack } from '../utils/repo-detection.js'
import { ROOT_FILE_BY_TOOL } from './root-file-map.js'

const DEFAULT_ENABLED_FEATURES: FeatureFlags = {
  contextEngineering: true,
  rpiWorkflow: true,
  chainOfThought: true,
  treeOfThoughts: true,
  adrEnforcement: true,
  qualityGates: true,
  agentHarness: true,
  bugResolution: true,
  pivotHandling: true,
}

export interface ScaffoldCompiledRootOptions {
  targetDir: string
  libraryDir: string
  tools: ToolId[]
  projectName: string
  planningDir: string
  features?: FeatureFlags
  gitConventions?: GitConventions
  fileRecords: FileRecord[]
  strategy: ConflictStrategy
  perFileOverrides: Map<string, ConflictStrategy>
  setupScope?: SetupScope
  // Optional context overrides
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
  /** Referenced repos for workspace scope — appended as context to compiled root */
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
}

/**
 * Compiles and writes root AI tool configuration files using the shared
 * fragment/template compilation system.
 *
 * Default behavior:
 * - When features are omitted, schema/wizard defaults are used (all feature flags enabled)
 * - Callers can pass explicit features to disable specific blocks
 * - Git-conventions blocks are included when gitConventions context is provided
 */
export async function scaffoldCompiledRoot(opts: ScaffoldCompiledRootOptions): Promise<void> {
  const {
    targetDir,
    libraryDir,
    tools,
    projectName,
    planningDir,
    features,
    gitConventions,
    fileRecords,
    strategy,
    perFileOverrides,
    setupScope,
    primaryLanguage,
    framework,
    workspaceType,
    projectInstructions,
    repos,
  } = opts

  const effectiveFeatures: FeatureFlags = {
    ...DEFAULT_ENABLED_FEATURES,
    ...(features ?? {}),
  }
  const stack = setupScope === 'project' ? detectProjectStack(targetDir) : undefined
  const effectivePrimaryLanguage = primaryLanguage ?? stack?.language
  const effectiveFramework = framework ?? stack?.framework

  // Build fragment context from options
  const context: FragmentContext = {
    projectName,
    planningDir,
    ...(effectivePrimaryLanguage !== undefined ? { primaryLanguage: effectivePrimaryLanguage } : {}),
    ...(effectiveFramework !== undefined ? { framework: effectiveFramework } : {}),
    ...(workspaceType != null ? { workspaceType } : {}),
    ...(projectInstructions != null ? { projectInstructions } : {}),
    ...(stack?.testFramework ? { testFramework: stack.testFramework } : {}),
    ...(stack?.packageManager ? { packageManager: stack.packageManager } : {}),
    ...(stack?.commands.test ? { testCommand: stack.commands.test } : {}),
    ...(stack?.commands.lint ? { lintCommand: stack.commands.lint } : {}),
    ...(stack?.commands.build ? { buildCommand: stack.commands.build } : {}),
    ...(stack?.commands.dev ? { devCommand: stack.commands.dev } : {}),
    ...(stack?.commands.install ? { installCommand: stack.commands.install } : {}),
    ...(stack?.description ? { projectDescription: stack.description } : {}),
    features: {
      contextEngineering: effectiveFeatures.contextEngineering,
      rpiWorkflow: effectiveFeatures.rpiWorkflow,
      chainOfThought: effectiveFeatures.chainOfThought,
      treeOfThoughts: effectiveFeatures.treeOfThoughts,
      adrEnforcement: effectiveFeatures.adrEnforcement,
      qualityGates: effectiveFeatures.qualityGates,
      agentHarness: effectiveFeatures.agentHarness,
      bugResolution: effectiveFeatures.bugResolution,
      pivotHandling: effectiveFeatures.pivotHandling,
      gitConventions: Boolean(gitConventions),
      // Legacy aliases for existing snake_case template conditionals
      context_engineering: effectiveFeatures.contextEngineering,
      rpi_workflow: effectiveFeatures.rpiWorkflow,
      chain_of_thought: effectiveFeatures.chainOfThought,
      tree_of_thoughts: effectiveFeatures.treeOfThoughts,
      adr_enforcement: effectiveFeatures.adrEnforcement,
      quality_gates: effectiveFeatures.qualityGates,
      agent_harness: effectiveFeatures.agentHarness,
      bug_resolution: effectiveFeatures.bugResolution,
      pivot_handling: effectiveFeatures.pivotHandling,
      git_conventions: Boolean(gitConventions),
    },
  }

  // Compile for each tool
  for (const tool of tools) {
    const compiler = new TemplateCompiler({
      libraryDir,
      outputDir: targetDir,
      tool,
      context,
    })

    const result = compiler.compile()

    // For workspace scope, append repo context to compiled output
    let workspaceReposSection = ''
    if (repos && repos.length > 0) {
      const lines = [
        '',
        '## Workspace Repos',
        '',
        'This workspace contains the following repositories:',
        '',
      ]
      for (const repo of repos) {
        lines.push(`### ${repo.name}`)
        lines.push('')
        lines.push(`- **Path**: \`${repo.path}\``)
        if (repo.type && repo.type !== 'unknown') lines.push(`- **Type**: ${repo.type}`)
        if (repo.description) lines.push(`- **Description**: ${repo.description}`)
        lines.push('')
      }
      lines.push('When working in a repo, refer to its README or package.json for repo-specific details.')
      lines.push('')
      workspaceReposSection = lines.join('\n')
    }

    // Write each compiled file
    for (const file of result.files) {
      // Map 'root.md' to tool-specific filename (e.g., CLAUDE.md, AGENTS.md)
      let outputPath = file.relativePath
      if (outputPath === 'root.md') {
        outputPath = ROOT_FILE_BY_TOOL[tool]
      }

      const destPath = path.join(targetDir, outputPath)
      const destDir = path.dirname(destPath)
      
      // Ensure parent directory exists
      ensureDir(destDir)

      // Check conflict strategy
      const action = applyStrategy(destPath, strategy, perFileOverrides, targetDir)
      if (action === 'skip') continue

      // Write the compiled content
      const content = workspaceReposSection ? file.content + workspaceReposSection : file.content
      writeFile(destPath, content)

      // Record the file
      fileRecords.push({
        path: outputPath,
        hash: fileHash(destPath),
        source: `compiled:${tool}`,
        owner: 'library',
      })
    }
  }
}
