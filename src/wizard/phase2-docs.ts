import * as p from '@clack/prompts'
import type { DocsDirId, WizardSelections } from '../types.js'

const ALL_DOCS_DIRS: { value: DocsDirId; label: string; hint: string }[] = [
  { value: 'features', label: 'features', hint: 'Feature specs and PRDs' },
  { value: 'bugfixes', label: 'bugfixes', hint: 'Bug investigation and fix docs' },
  { value: 'refactors', label: 'refactors', hint: 'Refactoring plans and tracking' },
  { value: 'tech-debt', label: 'tech-debt', hint: 'Technical debt documentation' },
  { value: 'adrs', label: 'adrs', hint: 'Architecture Decision Records' },
  { value: 'memory', label: 'memory', hint: 'AI memory and context persistence' },
  { value: 'prompts', label: 'prompts', hint: 'Prompt library and templates' },
  { value: 'standards', label: 'standards', hint: 'Coding standards and conventions' },
  { value: 'templates', label: 'templates', hint: 'Document templates (PRD, ADR, etc.)' },
  { value: 'rules', label: 'rules', hint: 'AI behavior rules and constraints' },
]

export interface Phase2Result {
  docsDirs: DocsDirId[]
  docsAgents: DocsDirId[]
}

export async function runPhase2(opts: {
  interactive: boolean
  prior: Partial<WizardSelections>
}): Promise<Phase2Result> {
  const allDirIds: DocsDirId[] = ALL_DOCS_DIRS.map(d => d.value)

  if (!opts.interactive) {
    return { docsDirs: allDirIds, docsAgents: allDirIds }
  }

  // Phase 2a: Which docs/ subdirectories to create
  const priorDirs = opts.prior.docsDirs ?? allDirIds
  const selectedDirs = await p.multiselect({
    message: 'Which documentation directories do you want?',
    options: ALL_DOCS_DIRS.map(d => ({
      value: d.value,
      label: d.label,
      hint: d.hint,
    })),
    initialValues: priorDirs,
  })

  if (p.isCancel(selectedDirs)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const dirs = selectedDirs as DocsDirId[]

  // Phase 2b: Which selected dirs should also get an AGENTS.md
  if (dirs.length === 0) {
    return { docsDirs: [], docsAgents: [] }
  }

  const priorAgents = (opts.prior.docsAgents ?? dirs).filter(a => dirs.includes(a))
  const selectedAgents = await p.multiselect({
    message: 'Which directories should include an AGENTS.md file?',
    options: dirs.map(d => ({
      value: d,
      label: d,
      hint: `Add AGENTS.md to docs/${d}/`,
    })),
    initialValues: priorAgents,
    required: false,
  })

  if (p.isCancel(selectedAgents)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  const agents = selectedAgents as DocsDirId[]

  return { docsDirs: dirs, docsAgents: agents }
}
