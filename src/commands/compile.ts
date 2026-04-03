import type { Command } from 'commander'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
import { homedir } from 'node:os'
import type { SetupScope, ToolId } from '../types.js'
import { AdapterRegistry } from '../adapters/registry.js'
import { compileMcp } from '../adapters/mcp-compiler.js'
import { scaffoldCompiledRoot } from '../scaffold/compiled-root.js'
import { scaffoldRootFiles } from '../scaffold/root-files.js'
import { Errors } from '../errors/index.js'
import { fileExists, resolveLibraryDir } from '../utils/files.js'
import { readStoreReadonly } from '../store/index.js'
import {
  isGlobalSupportedTool,
  logUnsupportedGlobalTool,
  resolveGlobalToolTargetDir,
} from '../utils/global-paths.js'

interface CompileOptions {
  scope?: SetupScope
  tools?: string
  force?: boolean
  dryRun?: boolean
  planningRepo?: string
}

const libraryDir = resolveLibraryDir(dirname(fileURLToPath(import.meta.url)))

function parseTools(tools: string | undefined, registry: AdapterRegistry): ToolId[] | undefined {
  if (!tools) return undefined

  const parsed = tools
    .split(',')
    .map((tool) => tool.trim())
    .filter(Boolean)

  const registered = new Set(registry.getRegisteredIds())
  const invalid = parsed.filter((tool) => !registered.has(tool))

  if (invalid.length > 0) {
    throw Errors.invalidInput(`unknown tool(s): ${invalid.join(', ')}`, {
      available: [...registered],
    })
  }

  return parsed as ToolId[]
}

function resolveStoreDir(
  scope: SetupScope | undefined,
  cwd: string,
  userHomeDir: string,
  planningRepo: string | undefined,
  storePlanningRepoPath: string | undefined,
): string {
  if (scope === 'global') {
    return join(userHomeDir, '.ai')
  }

  if (scope === 'workspace') {
    if (planningRepo) return planningRepo
    if (storePlanningRepoPath) return storePlanningRepoPath
  }

  return cwd
}

export function registerCompile(program: Command): void {
  program
    .command('compile')
    .description('Compile .ai artifacts to tool-native directories')
    .option('--scope <scope>', 'Scope: global | workspace | project')
    .option('--tools <tools>', 'Comma-separated tool list to compile')
    .option('--force', 'Overwrite existing files')
     .option('--dry-run', 'Preview changes without writing files')
    .option('--planning-repo <path>', 'Planning repo path (for workspace scope)')
    .action(async (opts: CompileOptions) => {
      const cwd = process.cwd()
      const userHomeDir = homedir()
      const registry = new AdapterRegistry()

      let workspaceStorePlanningRepoPath: string | undefined
      let workspaceStoreFromCwd: Awaited<ReturnType<typeof readStoreReadonly>> | undefined

      if (opts.scope === 'workspace' && !opts.planningRepo) {
        const cwdManifestPath = join(cwd, '.ai-setup.json')
        if (fileExists(cwdManifestPath)) {
          workspaceStoreFromCwd = await readStoreReadonly(cwd)
          workspaceStorePlanningRepoPath = workspaceStoreFromCwd.config.planningRepoPath
        }
      }

      const storeDir = resolveStoreDir(
        opts.scope,
        cwd,
        userHomeDir,
        opts.planningRepo,
        workspaceStorePlanningRepoPath,
      )
      const manifestPath = join(storeDir, '.ai-setup.json')

      if (!fileExists(manifestPath)) {
        if (opts.scope === 'workspace') {
          throw Errors.invalidInput(
            `Setup manifest not found in ${storeDir}. For workspace scope, run compile from the planning repo directory.`,
          )
        }
        throw Errors.manifestNotFound(storeDir)
      }

      const store = workspaceStoreFromCwd && storeDir === cwd ? workspaceStoreFromCwd : await readStoreReadonly(storeDir)
      const effectiveScope = opts.scope ?? store.config.setupScope

      const overrideTools = parseTools(opts.tools, registry)
      const requestedTools = overrideTools ?? store.config.tools

      if (requestedTools.length === 0) {
        throw Errors.invalidInput('no tools configured to compile')
      }

      const installableTools =
        effectiveScope === 'global'
          ? requestedTools.filter((tool: ToolId) => {
              const supported = isGlobalSupportedTool(tool)
              if (!supported) {
                logUnsupportedGlobalTool(tool)
              }
              return supported
            })
          : requestedTools

      if (installableTools.length === 0) {
        throw Errors.invalidInput('no globally supported tools selected for compile')
      }

      const selections = {
        agents: store.selections.agents,
        skills: store.selections.skills,
        prompts: store.selections.prompts,
      }
      const strategy = opts.force ? 'backup-and-replace' : 'skip'
      const planningDir = store.config.planningDir ?? '.planning'
      const useCompiledRoot = store.config.useCompiledRoot ?? true

      if (opts.dryRun) {
        console.log('[dry-run] Compile preview:')
        console.log(`[dry-run] Root strategy: ${useCompiledRoot ? 'compiled' : 'simple'} (planningDir=${planningDir})`)
        for (const tool of installableTools) {
          const adapterTargetDir =
            effectiveScope === 'global'
              ? resolveGlobalToolTargetDir(tool, userHomeDir)
              : store.config.targetDir

          if (!adapterTargetDir) continue

          console.log(`[dry-run] Would compile tool: ${tool} -> ${adapterTargetDir}`)
        }

        console.log(`Dry run complete. Would compile ${installableTools.length} tool(s): ${installableTools.join(', ')}`)
        return
      }

      if (useCompiledRoot) {
        await scaffoldCompiledRoot({
          targetDir: store.config.targetDir,
          libraryDir,
          tools: installableTools,
          projectName: store.config.projectName,
          planningDir,
          ...(store.selections.features != null ? { features: store.selections.features } : {}),
          ...(store.selections.gitConventions != null ? { gitConventions: store.selections.gitConventions } : {}),
          fileRecords: [],
          strategy,
          perFileOverrides: new Map(),
        })
      } else {
        await scaffoldRootFiles({
          targetDir: store.config.targetDir,
          libraryDir,
          tools: installableTools,
          projectName: store.config.projectName,
          fileRecords: [],
          strategy,
          perFileOverrides: new Map(),
        })
      }

      for (const tool of installableTools) {
        const adapter = registry.get(tool)
        if (!adapter) continue

        const adapterTargetDir =
          effectiveScope === 'global'
            ? resolveGlobalToolTargetDir(tool, userHomeDir)
            : store.config.targetDir

        if (!adapterTargetDir) continue

        await adapter.install({
          targetDir: adapterTargetDir,
          setupScope: effectiveScope,
          libraryDir,
          fileRecords: [],
          force: opts.force,
          strategy,
          selections,
        })

        await compileMcp({
          canonicalDir: storeDir,
          toolTargetDir: adapterTargetDir ?? storeDir,
          toolId: tool,
          fileRecords: [],
        })
      }

      console.log(`✅ Compiled ${installableTools.length} tool(s): ${installableTools.join(', ')}`)
    })
}
