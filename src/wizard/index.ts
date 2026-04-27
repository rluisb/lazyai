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
import { scaffoldOrchestration } from '../scaffold/orchestration.js'
import { scaffoldRepoLedgers } from '../scaffold/repo-roots.js'
import { scaffoldSpecs } from '../scaffold/specs.js'
import { scaffoldTemplatesRules } from '../scaffold/templates-rules.js'
import { appendOperation, writeStore } from '../store/index.js'
import type { FeatureFlags, GitConventions, StoreData } from '../store/schema.js'
import type {
  AgentId,
  FileRecord,
  PresetLevel,
  SetupScope,
  SetupType,
  SkillId,
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
import { formatOrchestratorHintBody } from '../utils/orchestrator-hints.js'
import { GO_BACK, showPhaseComplete, showPhaseProgress } from '../utils/ui.js'
import { defaultMcpServersForPreset, type McpWizardPreset, runPhase1 } from './phase1-context.js'
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

/**
 * Collected answers from the wizard phases.
 * Used as mutable state during the phase loop — values accumulate
 * as each phase completes and are available for re-derivation
 * when the user navigates back.
 */
interface WizardState {
  setupScope: SetupScope
  tools: ToolId[]
  skills: SkillId[]
  agents: AgentId[]
  mcpPreset: McpWizardPreset
  projectName: string
  workspaceName?: string
  planningRepoPath?: string
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
  cliTools?: string[]
  enableServers?: string[]
  organization?: string
  team?: string
  planningDir: string
  features: FeatureFlags
  gitConventions: GitConventions
  preset: PresetLevel
}

/**
 * Run Phases 1-2 of the wizard, supporting back navigation between them.
 * Returns the accumulated state that Phase 1 and 2 produce.
 */
async function runPhase12Loop(opts: {
  interactive: boolean
  prior: Partial<WizardSelections> & {
    setupScope?: SetupScope
    setupType?: SetupType
    tools?: ToolId[]
    skills?: SkillId[]
    agents?: AgentId[]
    mcpPreset?: McpWizardPreset
    projectName?: string
    workspaceName?: string
    planningRepoPath?: string
    planningDir?: string
    enableServers?: string[]
    organization?: string
    team?: string
    features?: Partial<FeatureFlags>
    gitConventions?: Partial<GitConventions>
  }
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
    installMode?: 'copy' | 'symlink'
  }
  targetDir: string
}): Promise<WizardState> {
  // Accumulated answers — start from prior manifest or defaults
  let accSetupScope: SetupScope = opts.prior.setupScope ?? opts.prior.setupType ?? 'project'
  let accTools: ToolId[] = opts.prior.tools ?? []
  let accSkills: SkillId[] = opts.prior.skills ?? ALL_SKILLS
  let accAgents: AgentId[] = opts.prior.agents ?? ALL_AGENTS
  let accMcpPreset: McpWizardPreset = opts.prior.mcpPreset ?? 'recommended'
  let accProjectName: string = opts.prior.projectName ?? ''
  let accWorkspaceName: string | undefined = opts.prior.workspaceName
  let accPlanningRepoPath: string | undefined = opts.prior.planningRepoPath
  let accRepos: Array<{ name: string; path: string; type?: string; description?: string }> | undefined
  let accCliTools: string[] | undefined
  let accEnableServers: string[] | undefined = opts.prior.enableServers ?? defaultMcpServersForPreset(accMcpPreset, opts.targetDir)
  let accOrganization: string | undefined = opts.prior.organization
  let accTeam: string | undefined = opts.prior.team
  let accPlanningDir: string = opts.prior.planningDir ?? '.planning'
  let accFeatures: FeatureFlags = opts.prior.features as FeatureFlags | undefined ?? {
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
  let accGitConventions: GitConventions = opts.prior.gitConventions as GitConventions | undefined ?? {
      branchPattern: '{type}/{ticket}-{description}',
      commitPattern: '{type}({scope}): {description}',
      types: ['feat', 'fix', 'docs', 'style', 'refactor', 'perf', 'test', 'build', 'ci', 'chore', 'revert'],
      requireTicket: false,
      ticketPattern: '[A-Z]+-[0-9]+',
    }
  let accPreset: PresetLevel = 'standard'

  let currentPhase = 1

  // Loop until both phases complete without going back
  while (currentPhase >= 1 && currentPhase <= 2) {
    if (currentPhase === 1) {
      if (opts.interactive) {
        showPhaseProgress({ current: 1, total: 4, name: 'Setup Context' })
      }

      // Build prior for Phase 1 from accumulated state + original manifest prior
      const phase1Prior: typeof opts.prior = {
        ...opts.prior,
        setupScope: accSetupScope,
        tools: accTools,
        skills: accSkills,
        agents: accAgents,
        mcpPreset: accMcpPreset,
        projectName: accProjectName,
        ...(accWorkspaceName != null ? { workspaceName: accWorkspaceName } : {}),
        ...(accPlanningRepoPath != null ? { planningRepoPath: accPlanningRepoPath } : {}),
        ...(accEnableServers != null ? { enableServers: accEnableServers } : {}),
        ...(accOrganization != null ? { organization: accOrganization } : {}),
        ...(accTeam != null ? { team: accTeam } : {}),
      }

      const phase1Result = await runPhase1({
        interactive: opts.interactive,
        prior: phase1Prior,
        cliOverrides: opts.cliOverrides,
        targetDir: opts.targetDir,
        canGoBack: false, // Phase 1 is the first phase
      })

      if (phase1Result === GO_BACK) {
        // Can't go back from Phase 1 — just continue
        continue
      }

      // Accumulate Phase 1 answers
      accSetupScope = phase1Result.setupScope
      accTools = phase1Result.tools
      accSkills = phase1Result.skills
      accAgents = phase1Result.agents
      accMcpPreset = phase1Result.mcpPreset
      accProjectName = phase1Result.projectName
      accWorkspaceName = phase1Result.workspaceName
      accPlanningRepoPath = phase1Result.planningRepoPath
      accRepos = phase1Result.repos
      accCliTools = phase1Result.cliTools
      accEnableServers = phase1Result.enableServers
      accOrganization = phase1Result.organization
      accTeam = phase1Result.team

      currentPhase = 2
    }

    if (currentPhase === 2) {
      if (opts.interactive) {
        showPhaseProgress({ current: 2, total: 4, name: 'Features & Conventions' })
      }

      const phase2Prior: {
        planningDir?: string
        features?: Partial<FeatureFlags>
        gitConventions?: Partial<GitConventions>
      } = {
        planningDir: accPlanningDir,
        ...(accFeatures ? { features: accFeatures } : {}),
        ...(accGitConventions ? { gitConventions: accGitConventions } : {}),
      }

      const phase2Result = await runPhase2Features({
        interactive: opts.interactive,
        setupScope: accSetupScope,
        prior: phase2Prior,
        cliOverrides: {
          ...(opts.cliOverrides.planningDir != null ? { planningDir: opts.cliOverrides.planningDir } : {}),
          ...(opts.cliOverrides.preset != null ? { preset: opts.cliOverrides.preset } : {}),
          ...(opts.cliOverrides.features != null ? { features: opts.cliOverrides.features } : {}),
          ...(opts.cliOverrides.disableFeatures != null ? { disableFeatures: opts.cliOverrides.disableFeatures } : {}),
          ...(opts.cliOverrides.branchPattern != null ? { branchPattern: opts.cliOverrides.branchPattern } : {}),
          ...(opts.cliOverrides.commitPattern != null ? { commitPattern: opts.cliOverrides.commitPattern } : {}),
        },
      })

      if (phase2Result === GO_BACK) {
        currentPhase = 1
        continue
      }

      // Accumulate Phase 2 answers
      accPlanningDir = phase2Result.planningDir
      accFeatures = phase2Result.features
      accGitConventions = phase2Result.gitConventions
      accPreset = phase2Result.preset

      currentPhase = 3 // Done with Phase 1-2
    }
  }

  return {
    setupScope: accSetupScope,
    tools: accTools,
    skills: accSkills,
    agents: accAgents,
    mcpPreset: accMcpPreset,
    projectName: accProjectName,
    ...(accWorkspaceName != null ? { workspaceName: accWorkspaceName } : {}),
    ...(accPlanningRepoPath != null ? { planningRepoPath: accPlanningRepoPath } : {}),
    ...(accRepos != null ? { repos: accRepos } : {}),
    ...(accCliTools != null ? { cliTools: accCliTools } : {}),
    ...(accEnableServers != null ? { enableServers: accEnableServers } : {}),
    ...(accOrganization != null ? { organization: accOrganization } : {}),
    ...(accTeam != null ? { team: accTeam } : {}),
    planningDir: accPlanningDir,
    features: accFeatures,
    gitConventions: accGitConventions,
    preset: accPreset,
  }
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
      installMode?: 'copy' | 'symlink'
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
      skills?: SkillId[]
      agents?: AgentId[]
      mcpPreset?: McpWizardPreset
      projectName?: string
      workspaceName?: string
      planningRepoPath?: string
      planningDir?: string
      cliTools?: string[]
      enableServers?: string[]
      organization?: string
      team?: string
      features?: Partial<FeatureFlags>
      gitConventions?: Partial<GitConventions>
    } = manifest
      ? {
          ...restSelections,
          setupScope: manifest.setupScope,
          ...(manifest.setupType ? { setupType: manifest.setupType } : {}),
          tools: manifest.tools,
          projectName: manifest.projectName,
          ...(manifest.workspaceName != null ? { workspaceName: manifest.workspaceName } : {}),
          ...(manifest.planningRepoPath != null ? { planningRepoPath: manifest.planningRepoPath } : {}),
          ...(manifest.cliTools != null ? { cliTools: manifest.cliTools } : {}),
          ...(manifest.enableServers != null ? { enableServers: manifest.enableServers } : {}),
          ...(manifest.planningDir != null ? { planningDir: manifest.planningDir } : {}),
          ...(manifest.features != null ? { features: manifest.features } : {}),
          ...(manifest.gitConventions != null ? { gitConventions: manifest.gitConventions } : {}),
        }
      : {}

    // --- Phase 1-2 Loop with Back Navigation ---
    // This loop collects all Phase 1+2 answers. If user goes back from Phase 2,
    // Phase 1 reruns with accumulated state pre-filled.
    let state = await runPhase12Loop({
      interactive: opts.interactive,
      prior,
      cliOverrides: opts.cliOverrides,
      targetDir: opts.targetDir,
    })

    // --- Phases 3-4 with outer loop for back navigation ---
    // Phase 4's "Back" returns to Phase 2, which requires re-running
    // the Phase 1-2 loop and recomputing all derived values.
    while (true) {
      // Derive computed values from the current state
      const {
        setupScope,
        tools,
        skills,
        agents,
        projectName,
        workspaceName,
        planningRepoPath,
        repos,
        cliTools,
        enableServers,
        planningDir,
        features,
        gitConventions,
        preset,
      } = state

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
        agents,
        skills,
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

      if (opts.interactive) {
        showPhaseProgress({ current: 3, total: 4, name: 'Conflict Resolution' })
      }
      const { strategy, perFileOverrides } = await runPhase3(phase3Opts)

      if (opts.interactive) {
        showPhaseProgress({ current: 4, total: 4, name: 'Review & Confirm' })
      }
      const phase4Opts = { interactive: opts.interactive, plan, config }
      const phase4Result = await runPhase4(phase4Opts)

      if (phase4Result === GO_BACK) {
        // Go back to Phase 2 — re-run the Phase 1-2 loop with current state
        state = await runPhase12Loop({
          interactive: opts.interactive,
          prior,
          cliOverrides: opts.cliOverrides,
          targetDir: opts.targetDir,
        })
        // Loop back to recompute derived values and re-run Phase 3-4
        continue
      }

      if (!phase4Result) return

      // === Installation Phase (confirmed) ===

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

        if (enableServers?.includes('orchestrator')) {
          await scaffoldOrchestration({
            targetDir: effectiveTargetDir,
            libraryDir,
            fileRecords,
            strategy,
            perFileOverrides,
          })
          tracker.trackSuccess('scaffold:orchestration')
        }

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
          ...(repos && repos.length > 0 ? { repos } : {}),
        })
        tracker.trackSuccess('scaffold:compiled-root')

        await scaffoldAgentsSkillsPrompts({
          targetDir: effectiveTargetDir,
          libraryDir,
          tools: installableTools,
          agents: selections.agents,
          skills: selections.skills,
          prompts: selections.prompts,
          ...(enableServers ? { enableServers } : {}),
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
            setupScope,
            homeDir: userHomeDir,
          })
        }
        tracker.trackSuccess('compile:mcp')

        if (setupScope === 'workspace' && repos && repos.length > 0) {
          // Ledgers are written in the workspace root (specs/memory/repos/), not in referenced repos
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
            homeDir: userHomeDir,
            libraryDir,
            fileRecords: [],
            force: opts.force,
            strategy,
            ...(opts.cliOverrides.installMode ? { installMode: opts.cliOverrides.installMode } : {}),
          })

          await compileMcp({
            canonicalDir: effectiveTargetDir,
            toolTargetDir: globalToolTargetDir,
            toolId: tool,
            fileRecords: [],
            setupScope,
            homeDir: userHomeDir,
          })
        }
      }

      if (opts.interactive) {
        const s = p.spinner()
        s.start('Installing files...')
        await installFiles()
        await installGlobalAdapters()
        s.stop('Files installed successfully!')

        // Show what was created
        const createdCount = fileRecords.length
        if (createdCount > 0) {
          p.log.success(`Created ${createdCount} file${createdCount === 1 ? '' : 's'}`)
        }

        showPhaseComplete(4)
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
          ...(enableServers && enableServers.length > 0 ? { enableServers } : {}),
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

      if (enableServers?.includes('orchestrator') && installableTools.length > 0) {
        const representativeTool = installableTools[0] as ToolId
        p.note(formatOrchestratorHintBody(representativeTool), '💡 Orchestrator MCP is configured')
      }

      outroSuccess(config)

      // Break out of the outer while(true) loop — installation complete
      break
    }
  } catch (error) {
    if (p.isCancel(error)) {
      throw Errors.userCancelled()
    }

    tracker.trackFailure('wizard:init', error instanceof Error ? error.message : String(error))

    if (error instanceof Error) throw error
    throw Errors.unknown(String(error))
  }
}
