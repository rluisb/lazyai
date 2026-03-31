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
  AiSetupConfig,
  FileRecord,
  SetupType,
  ToolId,
  WizardConfig,
  WizardSelections,
} from '../types.js'
import { readFile, writeFile } from '../utils/files.js'
import { extractSelections, readManifest } from '../utils/manifest.js'
import { runPhase1 } from './phase1-context.js'
import { runPhase2 } from './phase2-docs.js'
import { runPhase3 } from './phase3-templates.js'
import { runPhase4 } from './phase4-agents.js'
import { runPhase5 } from './phase5-infra.js'
import { runPhase6 } from './phase6-root.js'
import { runPhase7 } from './phase7-conflicts.js'
import { runPhase8 } from './phase8-confirm.js'
import { planFiles } from './planner.js'

export async function runWizard(opts: {
  interactive: boolean
  force?: boolean
  cliOverrides: { type?: SetupType; tools?: ToolId[]; name?: string }
  targetDir: string
}): Promise<void> {
  const __dirname = path.dirname(fileURLToPath(import.meta.url))
  const libraryDir = path.join(__dirname, '../../library')

  if (opts.interactive) {
    p.intro('ai-setup wizard')
  }

  const manifest = await readManifest(opts.targetDir)
  const prior: Partial<WizardSelections> & {
    setupType?: SetupType
    tools?: ToolId[]
    projectName?: string
  } = manifest
    ? {
        ...extractSelections(manifest),
        setupType: manifest.setupType,
        tools: manifest.tools,
        projectName: manifest.projectName,
      }
    : {}

  const { setupType, tools, projectName } = await runPhase1({
    interactive: opts.interactive,
    prior,
    cliOverrides: opts.cliOverrides,
    targetDir: opts.targetDir,
  })

  const { docsDirs, docsAgents } = await runPhase2({
    interactive: opts.interactive,
    prior,
  })

  const { templates, rules } = await runPhase3({
    interactive: opts.interactive,
    prior,
  })

  const { agents, skills, prompts } = await runPhase4({
    interactive: opts.interactive,
    prior,
  })

  const { infra } = await runPhase5({
    interactive: opts.interactive,
    prior,
    targetDir: opts.targetDir,
  })

  const { rootFiles } = await runPhase6({
    interactive: opts.interactive,
    tools,
    projectName,
  })
  void rootFiles

  const selections: WizardSelections = {
    docsDirs,
    docsAgents,
    templates,
    rules,
    agents,
    skills,
    prompts,
    infra,
  }

  const config: WizardConfig = {
    setupType,
    tools,
    projectName,
    targetDir: opts.targetDir,
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
      docsDirs,
      docsAgents,
      fileRecords,
      strategy,
      perFileOverrides,
    })

    await scaffoldTemplatesRules({
      targetDir: opts.targetDir,
      libraryDir,
      templates,
      rules,
      fileRecords,
      strategy,
      perFileOverrides,
    })

    await scaffoldInfra({
      targetDir: opts.targetDir,
      libraryDir,
      infra,
      projectName,
      fileRecords,
      strategy,
      perFileOverrides,
    })

    await scaffoldRootFiles({
      targetDir: opts.targetDir,
      libraryDir,
      tools,
      projectName,
      fileRecords,
      strategy,
      perFileOverrides,
    })

    await scaffoldAgentsSkillsPrompts({
      targetDir: opts.targetDir,
      libraryDir,
      tools,
      agents,
      skills,
      prompts,
      fileRecords,
      ...(opts.force !== undefined ? { force: opts.force } : {}),
    })
  }

  if (opts.interactive) {
    const s = p.spinner()
    s.start('Installing files...')
    await installFiles()
    s.stop('Files installed successfully!')
  } else {
    await installFiles()
  }

  const manifestData: AiSetupConfig = {
    version: '0.1.0',
    setupType,
    tools,
    projectName,
    installedAt: new Date().toISOString(),
    files: fileRecords,
    selections,
  }

  writeFile(path.join(opts.targetDir, '.ai-setup.json'), JSON.stringify(manifestData, null, 2))
  outroSuccess(config)
}
