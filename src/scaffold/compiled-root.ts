import path from 'node:path'
import { type FragmentContext, TemplateCompiler } from '../compiler/index.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import type { ConflictStrategy, FileRecord, ToolId } from '../types.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import { ensureDir, fileHash, writeFile } from '../utils/files.js'
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
  // Optional context overrides
  primaryLanguage?: string
  framework?: string
  workspaceType?: string
  projectInstructions?: string
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
    primaryLanguage,
    framework,
    workspaceType,
    projectInstructions,
  } = opts

  const effectiveFeatures: FeatureFlags = {
    ...DEFAULT_ENABLED_FEATURES,
    ...(features ?? {}),
  }

  // Build fragment context from options
  const context: FragmentContext = {
    projectName,
    planningDir,
    ...(primaryLanguage != null ? { primaryLanguage } : {}),
    ...(framework != null ? { framework } : {}),
    ...(workspaceType != null ? { workspaceType } : {}),
    ...(projectInstructions != null ? { projectInstructions } : {}),
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
      writeFile(destPath, file.content)

      // Record the file
      fileRecords.push({
        path: outputPath,
        hash: fileHash(destPath),
        source: `compiled:${tool}`,
      })
    }
  }
}
