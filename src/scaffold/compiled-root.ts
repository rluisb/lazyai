import path from 'node:path'
import { ensureDir, writeFile, fileHash } from '../utils/files.js'
import { applyStrategy } from '../utils/conflict-strategy.js'
import type { ToolId, FileRecord, ConflictStrategy } from '../types.js'
import type { FeatureFlags, GitConventions } from '../store/schema.js'
import { TemplateCompiler, type FragmentContext } from '../compiler/index.js'

// Root file mapping per tool - maps template output to tool-native filename
const ROOT_FILE_BY_TOOL: Record<ToolId, string> = {
  'claude-code': 'CLAUDE.md',
  'opencode': 'AGENTS.md',
  'codex': 'AGENTS.md',
  'copilot': '.github/copilot-instructions.md',
  'pi': 'INSTRUCTIONS.md',
  'gemini': 'GEMINI.md',
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
 * Compiles and writes root AI tool configuration files using the TemplateCompiler.
 * 
 * This is an alternative to scaffoldRootFiles that uses the fragment/template
 * compilation system to embed XML-tagged prompt engineering blocks.
 * 
 * Features enabled by default:
 * - context-engineering
 * - rpi-workflow  
 * - chain-of-thought
 * - adr-enforcement
 * - quality-gates
 * - pivot-handling
 * 
 * Optional features (off by default):
 * - tree-of-thoughts
 * - agent-harness
 * - bug-resolution
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

  // Build fragment context from options
  const context: FragmentContext = {
    projectName,
    planningDir,
    ...(primaryLanguage != null ? { primaryLanguage } : {}),
    ...(framework != null ? { framework } : {}),
    ...(workspaceType != null ? { workspaceType } : {}),
    ...(projectInstructions != null ? { projectInstructions } : {}),
    ...(features || gitConventions ? {
      features: {
        ...(features ? {
          contextEngineering: features.contextEngineering,
          rpiWorkflow: features.rpiWorkflow,
          chainOfThought: features.chainOfThought,
          treeOfThoughts: features.treeOfThoughts,
          adrEnforcement: features.adrEnforcement,
          qualityGates: features.qualityGates,
          agentHarness: features.agentHarness,
          bugResolution: features.bugResolution,
          pivotHandling: features.pivotHandling,
          // Legacy aliases for existing snake_case template conditionals
          tree_of_thoughts: features.treeOfThoughts,
          agent_harness: features.agentHarness,
          bug_resolution: features.bugResolution,
        } : {}),
        ...(gitConventions ? { gitConventions: true } : {}),
      },
    } : {}),
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
