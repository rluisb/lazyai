import fs from 'node:fs'
import { createRequire } from 'node:module'
import { homedir } from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import * as p from '@clack/prompts'
import { compileMcp } from '../adapters/mcp-compiler.js'
import { AdapterRegistry } from '../adapters/registry.js'
import { Errors } from '../errors/index.js'
import { OperationTracker } from '../errors/operation.js'
import { writeToCanonical } from '../migration/canonical-writer.js'
import { detectExistingSetup } from '../migration/detector.js'
import { getAllParsers } from '../migration/registry/discovery.js'
import { rulesForPreset, specsDirsForPreset, templatesForPreset } from '../presets.js'
import { outroSuccess } from '../prompts.js'
import { scaffoldAgentsSkillsPrompts } from '../scaffold/agents-skills-prompts.js'
import { scaffoldCompiledRoot } from '../scaffold/compiled-root.js'
import { scaffoldConstitution } from '../scaffold/constitution.js'
import { scaffoldEnvExample } from '../scaffold/env-example.js'
import { checkGitignoreGuidance } from '../scaffold/gitignore.js'
import { scaffoldInfra } from '../scaffold/infra.js'
import { scaffoldMcp } from '../scaffold/mcp.js'
import { scaffoldRepoLedgers, scaffoldRepoRoots } from '../scaffold/repo-roots.js'
import { scaffoldSpecs } from '../scaffold/specs.js'
import { scaffoldTemplatesRules } from '../scaffold/templates-rules.js'
import { appendOperation, writeStore } from '../store/index.js'
import type { FeatureFlags, GitConventions, StoreData } from '../store/schema.js'
import type {
  FileRecord,
  PresetLevel,
  SetupScope,
  SetupType,
  ToolId,
  WizardConfig,
  WizardSelections,
} from '../types.js'
import {
  ALL_AGENTS,
  ALL_INFRA,
  ALL_PROMPTS,
  ALL_SKILLS,
} from '../types.js'
import { fileExists, isDirectory, readFile, resolveLibraryDir, writeFile } from '../utils/files.js'
import {
  isGlobalSupportedTool,
  logUnsupportedGlobalTool,
  resolveGlobalToolTargetDir,
} from '../utils/global-paths.js'
import { extractSelections, readManifest } from '../utils/manifest.js'
import { runPhase1 } from './phase1-context.js'
import { runPhase2Features } from './phase2-features.js'
import { runPhase3 } from './phase3-conflicts.js'
import { runPhase4 } from './phase4-confirm.js'
import { computePlan } from './planner.js'

const _require = createRequire(import.meta.url)

function resolveCliVersion(): string {
  const candidates = ['../../package.json', '../package.json']

  for (const candidate of candidates) {
    try {
      const pkg = _require(candidate) as { version?: string }
      if (pkg.version) return pkg.version
    } catch {
      // try next candidate
    }
  }

  return '0.0.0'
}

const cliVersion = resolveCliVersion()

function buildDefaultSelections(targetDir: string): WizardSelections {
  const hasGitDir = fileExists(path.join(targetDir, '.git'))

  return {
    templates: [],
    rules: [],
    agents: ALL_AGENTS,
    skills: ALL_SKILLS,
    prompts: ALL_PROMPTS,
    infra: hasGitDir ? ALL_INFRA : ALL_INFRA.filter((item) => item !== 'pre-commit'),
    constitution: [] as string[],
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
  absorb?: boolean
  dryRun?: boolean
  homeDir?: string
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    cliTools?: string[]
    name?: string
    planningRepo?: string
    repos?: string[]
    planningDir?: string
    preset?: PresetLevel
    features?: string[]
    disableFeatures?: string[]
    branchPattern?: string
    commitPattern?: string
    enableServers?: string[]
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
    // Destructure to exclude features/gitConventions from spread (exactOptionalPropertyTypes)
    const { features: _ef, gitConventions: _eg, ...restSelections } = manifest
      ? extractSelections(manifest)
      : {}
    const prior: Partial<WizardSelections> & {
      setupScope?: SetupScope
      setupType?: SetupType
      tools?: ToolId[]
      projectName?: string
      workspaceName?: string
      planningRepoPath?: string
      planningDir?: string
      features?: Partial<FeatureFlags>
      gitConventions?: Partial<GitConventions>
    } = manifest
      ? {
          ...restSelections,
          setupScope: manifest.setupScope,
          ...(manifest.setupType ? { setupType: manifest.setupType } : {}),
          tools: manifest.tools,
          projectName: manifest.projectName,
          ...(manifest.planningDir != null ? { planningDir: manifest.planningDir } : {}),
          ...(manifest.features != null ? { features: manifest.features } : {}),
          ...(manifest.gitConventions != null ? { gitConventions: manifest.gitConventions } : {}),
        }
      : {}

    const { setupScope, tools, projectName, workspaceName, planningRepoPath, repos, cliTools, enableServers } = await runPhase1({
      interactive: opts.interactive,
      prior,
      cliOverrides: opts.cliOverrides,
      targetDir: opts.targetDir,
    })

    // Phase 2: Planning directory, feature flags, and git conventions
    const { planningDir, features, gitConventions, preset } = await runPhase2Features({
      interactive: opts.interactive,
      setupScope,
      prior: {
        ...(prior.planningDir != null ? { planningDir: prior.planningDir } : {}),
        ...(prior.features != null ? { features: prior.features } : {}),
        ...(prior.gitConventions != null ? { gitConventions: prior.gitConventions } : {}),
      },
      cliOverrides: {
        ...(opts.cliOverrides.planningDir != null ? { planningDir: opts.cliOverrides.planningDir } : {}),
        ...(opts.cliOverrides.preset != null ? { preset: opts.cliOverrides.preset } : {}),
        ...(opts.cliOverrides.features != null ? { features: opts.cliOverrides.features } : {}),
        ...(opts.cliOverrides.disableFeatures != null ? { disableFeatures: opts.cliOverrides.disableFeatures } : {}),
        ...(opts.cliOverrides.branchPattern != null ? { branchPattern: opts.cliOverrides.branchPattern } : {}),
        ...(opts.cliOverrides.commitPattern != null ? { commitPattern: opts.cliOverrides.commitPattern } : {}),
      },
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
    const specsDirs = specsDirsForPreset(preset)
    const templateIds = templatesForPreset(preset)
    const ruleIds = rulesForPreset(preset)

    if (setupScope === 'workspace' && repos && repos.length > 0) {
      for (const repo of repos) {
        const absPath = path.resolve(effectiveTargetDir, repo.path)

        if (!fileExists(absPath)) {
          p.log.warn(`⚠️  Repo path not found: ${absPath} (${repo.name})`)
          continue
        }

        if (!isDirectory(absPath)) {
          p.log.warn(`⚠️  Path is not a directory: ${absPath} (${repo.name})`)
          continue
        }

        try {
          const testPath = path.join(absPath, '.ai-setup-write-test')
          writeFile(testPath, '')
          fs.unlinkSync(testPath)
        } catch {
          p.log.warn(`⚠️  No write permission: ${absPath} (${repo.name}) — AI tools may not be able to modify files`)
        }
      }
    }

    if (setupScope === 'global') {
      for (const tool of tools) {
        if (!isGlobalSupportedTool(tool)) {
          logUnsupportedGlobalTool(tool)
        }
      }
    }

    const selections = {
      ...buildDefaultSelections(effectiveTargetDir),
      templates: templateIds,
      rules: ruleIds,
    }
    const fileRecords: FileRecord[] = []

    const migrationContext = {
      sourcePath: effectiveTargetDir,
      targetPath: effectiveTargetDir,
      options: {
        preview: false,
        mergeStrategy: 'preserve' as const,
        verbose: false,
        skipBackup: true,
        interactive: false,
      },
    }

    const detections = await detectExistingSetup(migrationContext)
    if (detections.length > 0) {
      let shouldAbsorb = opts.absorb === true

      if (!shouldAbsorb && opts.interactive) {
        const detectionSummary = detections.map((item) => item.adapterName).join(', ')
        const absorbAnswer = await p.confirm({
          message: `Found existing tool setup (${detectionSummary}). Absorb into .ai/?`,
          initialValue: true,
        })

        if (p.isCancel(absorbAnswer)) {
          throw Errors.userCancelled()
        }

        shouldAbsorb = absorbAnswer === true
      }

      if (shouldAbsorb) {
        const parsers = await getAllParsers(effectiveTargetDir)

        for (const detection of detections) {
          const parser = parsers.find((candidate) => candidate.id === detection.adapterId)
          if (!parser) continue

          const parsedSetup = await parser.parse(migrationContext)
          await writeToCanonical({
            targetDir: effectiveTargetDir,
            parsedSetup,
            fileRecords,
          })
        }
      }
    }

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

    const plan = await computePlan(config, effectiveTargetDir, selections)

    const plannedFiles = plan.map((file) => {
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

    const phase3Opts = {
      interactive: opts.interactive,
      targetDir: effectiveTargetDir,
      plannedFiles,
      ...(opts.force !== undefined ? { force: opts.force } : {}),
    }

    const { strategy, perFileOverrides } = await runPhase3(phase3Opts)

    const phase4Opts = { interactive: opts.interactive, plan, config }
    const confirmed = await runPhase4(phase4Opts)
    if (!confirmed) return

    if (opts.dryRun) {
      let plannedCount = 0
      for (const file of plan) {
        console.log(`[dry-run] Would create: ${file.destPath}`)
        plannedCount += 1
      }
      console.log(`Dry run complete. Would create ${plannedCount} files.`)
      return
    }

    const installFiles = async (): Promise<void> => {
      await scaffoldSpecs({
        targetDir: effectiveTargetDir,
        setupScope,
        libraryDir,
        specsDirs,
        specsAgents: specsDirs,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:specs')

      await scaffoldConstitution({
        targetDir: effectiveTargetDir,
        libraryDir,
        projectName: effectiveProjectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:constitution')

      await scaffoldMcp({
        targetDir: effectiveTargetDir,
        libraryDir,
        fileRecords,
        strategy,
        perFileOverrides,
        ...(cliTools ? { cliTools } : {}),
        ...(enableServers ? { enableServers } : {}),
      })
      tracker.trackSuccess('scaffold:mcp')

      await scaffoldEnvExample({
        targetDir: effectiveTargetDir,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:env-example')

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

      await scaffoldCompiledRoot({
        targetDir: effectiveTargetDir,
        libraryDir,
        tools: installableTools,
        projectName: effectiveProjectName,
        planningDir,
        setupScope,
        features,
        gitConventions,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:compiled-root')

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

      for (const tool of installableTools) {
        await compileMcp({
          canonicalDir: effectiveTargetDir,
          toolTargetDir: effectiveTargetDir,
          toolId: tool,
          fileRecords,
        })
      }
      tracker.trackSuccess('compile:mcp')

      if (setupScope === 'workspace' && repos && repos.length > 0) {
        const repoResults = await scaffoldRepoRoots({
          repos,
          planningRepoPath: effectiveTargetDir,
          tools: installableTools,
          strategy,
          perFileOverrides,
        })

        for (const records of repoResults.values()) {
          fileRecords.push(...records)
        }
        tracker.trackSuccess('scaffold:repo-roots')

        await scaffoldRepoLedgers({
          planningRepoPath: effectiveTargetDir,
          repos,
          fileRecords,
          strategy,
          perFileOverrides,
        })
        tracker.trackSuccess('scaffold:repo-ledgers')
      }

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

        await compileMcp({
          canonicalDir: effectiveTargetDir,
          toolTargetDir: globalToolTargetDir,
          toolId: tool,
          fileRecords: [],
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
        cliVersion,
        installedAt: now,
        lastUpdatedAt: now,
      },
      config: {
        setupScope,
        setupType: setupScope === 'global' ? 'project' : setupScope,
        tools: installableTools,
        ...(cliTools && cliTools.length > 0 ? { cliTools } : {}),
        projectName: effectiveProjectName,
        ...(workspaceName ? { workspaceName } : {}),
        targetDir: effectiveTargetDir,
        ...(planningRepoPath ? { planningRepoPath } : {}),
        ...(repos && repos.length > 0 ? { repos } : {}),
        ...(globalRef ? { globalRef } : {}),
        planningDir,
      },
      selections: {
        ...selections,
        features,
        gitConventions,
      },
      files: fileRecords.map((file) => ({
        ...file,
        owner: file.owner ?? 'library',
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
    checkGitignoreGuidance(effectiveTargetDir)
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
