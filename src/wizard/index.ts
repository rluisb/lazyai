import * as p from '@clack/prompts'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { outroSuccess } from '../prompts.js'
import { scaffoldAgentsSkillsPrompts } from '../scaffold/agents-skills-prompts.js'
import { scaffoldDocs } from '../scaffold/docs.js'
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

export async function runWizard(opts: {
  interactive: boolean
  force?: boolean
  cliOverrides: {
    scope?: SetupScope
    type?: SetupType
    tools?: ToolId[]
    name?: string
    planningRepo?: string
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

    const { setupScope, tools, projectName, planningRepoPath } = await runPhase1({
      interactive: opts.interactive,
      prior,
      cliOverrides: opts.cliOverrides,
      targetDir: opts.targetDir,
    })

    const selections = buildDefaultSelections(opts.targetDir)

    const config: WizardConfig = {
      setupScope,
      setupType: setupScope,
      tools,
      projectName,
      targetDir: opts.targetDir,
      ...(planningRepoPath ? { planningRepoPath } : {}),
      selections,
      interactive: opts.interactive,
      force: opts.force,
    }

    const plan = planFiles({ targetDir: opts.targetDir, libraryDir, config })

    const plannedFiles = plan.map(file => {
      const destPath = path.join(opts.targetDir, file.destPath)

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
      targetDir: opts.targetDir,
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
        targetDir: opts.targetDir,
        libraryDir,
        docsDirs: ALL_DOCS_DIRS,
        docsAgents: ALL_DOCS_DIRS,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:docs')

      await scaffoldTemplatesRules({
        targetDir: opts.targetDir,
        libraryDir,
        templates: selections.templates,
        rules: selections.rules,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:templates-rules')

      await scaffoldInfra({
        targetDir: opts.targetDir,
        libraryDir,
        infra: selections.infra,
        projectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:infra')

      await scaffoldRootFiles({
        targetDir: opts.targetDir,
        libraryDir,
        tools,
        projectName,
        fileRecords,
        strategy,
        perFileOverrides,
      })
      tracker.trackSuccess('scaffold:root-files')

      await scaffoldAgentsSkillsPrompts({
        targetDir: opts.targetDir,
        libraryDir,
        tools,
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

    if (opts.interactive) {
      const s = p.spinner()
      s.start('Installing files...')
      await installFiles()
      s.stop('Files installed successfully!')
    } else {
      await installFiles()
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
        tools,
        projectName,
        targetDir: opts.targetDir,
        ...(planningRepoPath ? { planningRepoPath } : {}),
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

    await writeStore(opts.targetDir, storeData)
    await appendOperation(opts.targetDir, tracker.toOperation())
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
