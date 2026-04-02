import * as p from '@clack/prompts'
import { homedir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { outroSuccess } from '../prompts.js'
import { scaffoldAgentsSkillsPrompts } from '../scaffold/agents-skills-prompts.js'
import { scaffoldDocs } from '../scaffold/docs.js'
import { scaffoldConstitution } from '../scaffold/constitution.js'
import { scaffoldInfra } from '../scaffold/infra.js'
import { scaffoldRootFiles } from '../scaffold/root-files.js'
import { scaffoldTemplatesRules } from '../scaffold/templates-rules.js'
import type {
  FileRecord,
  SetupScope,
  SetupType,
  ToolId,
  WizardConfig,
  WizardSelections,
} from '../types.js'
import {
  ALL_AGENTS,
  ALL_DOCS_DIRS,
  ALL_INFRA,
  ALL_PROMPTS,
  ALL_RULES,
  ALL_SKILLS,
  ALL_TEMPLATES,
} from '../types.js'
import type { StoreData } from '../store/schema.js'
import { fileExists, readFile, resolveLibraryDir } from '../utils/files.js'
import { appendOperation, writeStore } from '../store/index.js'
import { Errors } from '../errors/index.js'
import { OperationTracker } from '../errors/operation.js'
import { extractSelections, readManifest } from '../utils/manifest.js'
import { runPhase1 } from './phase1-context.js'
import { runPhase7 } from './phase7-conflicts.js'
import { runPhase8 } from './phase8-confirm.js'
import { planFiles } from './planner.js'
import { AdapterRegistry } from '../adapters/registry.js'
import {
  isGlobalSupportedTool,
  logUnsupportedGlobalTool,
  resolveGlobalToolTargetDir,
} from '../utils/global-paths.js'

function buildDefaultSelections(targetDir: string): WizardSelections {
  const hasGitDir = fileExists(path.join(targetDir, '.git'))

  return {
    templates: ALL_TEMPLATES,
    rules: ALL_RULES,
    agents: ALL_AGENTS,
    skills: ALL_SKILLS,
    prompts: ALL_PROMPTS,
    infra: hasGitDir ? ALL_INFRA : ALL_INFRA.filter((item) => item !== 'pre-commit'),
  }
}

function resolveTargetDirForScope(scope: SetupScope, targetDir: string, homeDir: string): string {
  if (scope === 'global') {
    return path.join(homeDir, '.ai')
  }

  return targetDir
}

export async function runWizard(opts: {
  interactive: boolean
  force?: boolean
  homeDir?: string
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    name?: string
    planningRepo?: string
    repos?: string[]
  }
  targetDir: string
}): Promise<void> {
  const tracker = new OperationTracker('init')

  try {
    const libraryDir = resolveLibraryDir(path.dirname(fileURLToPath(import.meta.url)))

    if (opts.interactive) {
      p.intro('ai-setup wizard')
    }

    const manifest = await readManifest(opts.targetDir)
    const prior: Partial<WizardSelections> & {
      setupScope?: SetupScope
      setupType?: SetupType
      tools?: ToolId[]
      projectName?: string
      workspaceName?: string
      planningRepoPath?: string
    } = manifest
      ? {
          ...extractSelections(manifest),
          setupScope: manifest.setupScope,
          ...(manifest.setupType ? { setupType: manifest.setupType } : {}),
          tools: manifest.tools,
          projectName: manifest.projectName,
        }
      : {}

    const { setupScope, tools, projectName, workspaceName, planningRepoPath, repos } = await runPhase1({
      interactive: opts.interactive,
      prior,
      cliOverrides: opts.cliOverrides,
      targetDir: opts.targetDir,
    })

    const userHomeDir = opts.homeDir ?? homedir()
    const effectiveTargetDir =
      setupScope === 'workspace'
        ? (() => {
            if (!planningRepoPath) {
              throw Errors.invalidInput('workspace setup requires planningRepoPath')
            }
            return path.resolve(planningRepoPath)
          })()
        : resolveTargetDirForScope(setupScope, opts.targetDir, userHomeDir)
    const effectiveProjectName = projectName || (setupScope === 'global' ? 'global' : path.basename(effectiveTargetDir))
    const globalRef = setupScope === 'workspace' && fileExists(path.join(userHomeDir, '.ai')) ? '~/.ai/' : undefined
    const installableTools = setupScope === 'global' ? tools.filter(isGlobalSupportedTool) : tools

    if (setupScope === 'global') {
      for (const tool of tools) {
        if (!isGlobalSupportedTool(tool)) {
          logUnsupportedGlobalTool(tool)
        }
      }
    }

    const selections = buildDefaultSelections(effectiveTargetDir)

    const config: WizardConfig = {
      setupScope,
      setupType: setupScope,
      tools: installableTools,
      projectName: effectiveProjectName,
      ...(workspaceName ? { workspaceName } : {}),
      targetDir: effectiveTargetDir,
      ...(planningRepoPath ? { planningRepoPath } : {}),
      ...(repos && repos.length > 0 ? { repos } : {}),
      ...(globalRef ? { globalRef } : {}),
      selections,
      interactive: opts.interactive,
      force: opts.force,
    }

    const plan = planFiles({ targetDir: effectiveTargetDir, libraryDir, config })

    const plannedFiles = plan.map(file => {
        const destPath = path.join(effectiveTargetDir, file.destPath)

      let srcContent = ''
      if (!file.isNew && file.srcPath) {
        srcContent = readFile(path.join(libraryDir, file.srcPath))
      }

      return {
        destPath,
        srcContent,
      }
    })

    const phase7Opts = {
      interactive: opts.interactive,
      targetDir: effectiveTargetDir,
      plannedFiles,
      ...(opts.force !== undefined ? { force: opts.force } : {}),
    }

    const { strategy, perFileOverrides } = await runPhase7(phase7Opts)

    const phase8Opts = { interactive: opts.interactive, plan, config }
    const confirmed = await runPhase8(phase8Opts)
    if (!confirmed) return

    const fileRecords: FileRecord[] = []

    const installFiles = async (): Promise<void> => {
      await scaffoldDocs({
        targetDir: effectiveTargetDir,
        setupScope,
        libraryDir,
        docsDirs: ALL_DOCS_DIRS,
        docsAgents: ALL_DOCS_DIRS,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:docs')

      await scaffoldConstitution({
        targetDir: effectiveTargetDir,
        libraryDir,
        projectName: effectiveProjectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:constitution')

      await scaffoldTemplatesRules({
          targetDir: effectiveTargetDir,
          libraryDir,
        templates: selections.templates,
        rules: selections.rules,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:templates-rules')

      await scaffoldInfra({
          targetDir: effectiveTargetDir,
          libraryDir,
        infra: selections.infra,
          projectName: effectiveProjectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:infra')

      await scaffoldRootFiles({
          targetDir: effectiveTargetDir,
          libraryDir,
          tools: installableTools,
          projectName: effectiveProjectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:root-files')

      await scaffoldAgentsSkillsPrompts({
          targetDir: effectiveTargetDir,
          libraryDir,
          tools: installableTools,
        agents: selections.agents,
        skills: selections.skills,
        prompts: selections.prompts,
        fileRecords,
        strategy,
        perFileOverrides,
        ...(opts.force !== undefined ? { force: opts.force } : {}),
      })
      tracker.trackSuccess('scaffold:agents-skills-prompts')

      for (const [sourcePath, strategyOverride] of perFileOverrides.entries()) {
        if (strategyOverride === 'backup-and-replace' || strategyOverride === 'align') {
          tracker.registerBackup(sourcePath, `${sourcePath}.backup`)
        }
      }
    }

    const installGlobalAdapters = async (): Promise<void> => {
      if (setupScope !== 'global') return

      const registry = new AdapterRegistry()
      for (const tool of installableTools) {
        const globalToolTargetDir = resolveGlobalToolTargetDir(tool, userHomeDir)
        if (!globalToolTargetDir) continue

        const adapter = registry.get(tool)
        if (!adapter) continue

        await adapter.install({
          targetDir: globalToolTargetDir,
          setupScope,
          libraryDir,
          fileRecords: [],
          force: opts.force,
          strategy,
        })
      }
    }

    if (opts.interactive) {
      const s = p.spinner()
      s.start('Installing files...')
      await installFiles()
      await installGlobalAdapters()
      s.stop('Files installed successfully!')
    } else {
      await installFiles()
      await installGlobalAdapters()
    }

    const now = new Date().toISOString()
    const storeData: StoreData = {
      meta: {
        schemaVersion: 1,
        cliVersion: '0.1.0',
        installedAt: now,
        lastUpdatedAt: now,
      },
      config: {
        setupScope,
        setupType: setupScope === 'global' ? 'project' : setupScope,
        tools: installableTools,
        projectName: effectiveProjectName,
        ...(workspaceName ? { workspaceName } : {}),
        targetDir: effectiveTargetDir,
        ...(planningRepoPath ? { planningRepoPath } : {}),
        ...(repos && repos.length > 0 ? { repos } : {}),
        ...(globalRef ? { globalRef } : {}),
      },
      selections,
      files: fileRecords.map((file) => ({
        ...file,
        status: 'installed',
        installedAt: now,
        lastCheckedAt: now,
      })),
      sync: {
        lastSyncAt: now,
        dirty: false,
      },
      operations: [],
    }

    await writeStore(effectiveTargetDir, storeData)
    await appendOperation(effectiveTargetDir, tracker.toOperation())
    outroSuccess(config)
  } catch (error) {
    if (p.isCancel(error)) {
      throw Errors.userCancelled()
    }

    tracker.trackFailure('wizard:init', error instanceof Error ? error.message : String(error))

    if (error instanceof Error) throw error
    throw Errors.unknown(String(error))
  }
}
