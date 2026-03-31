import { existsSync } from 'node:fs'
import { join } from 'node:path'
import * as p from '@clack/prompts'
import type { InfraId, WizardSelections } from '../types.js'

const ALL_INFRA: { value: InfraId; label: string; hint: string }[] = [
  { value: 'pre-commit', label: 'pre-commit', hint: 'Git pre-commit hook for AI checks' },
  { value: 'compliance', label: 'compliance', hint: 'AI compliance documentation' },
  { value: 'KNOWLEDGE_MAP', label: 'KNOWLEDGE_MAP', hint: 'Project knowledge map for AI context' },
]

export interface Phase5Result {
  infra: InfraId[]
}

export async function runPhase5(opts: {
  interactive: boolean
  prior: Partial<WizardSelections>
  targetDir: string
}): Promise<Phase5Result> {
  const hasGit = existsSync(join(opts.targetDir, '.git'))
  const allInfraIds: InfraId[] = ALL_INFRA.map(i => i.value)

  if (!opts.interactive) {
    // Non-interactive: include all, but exclude pre-commit if no .git
    const infra = hasGit ? allInfraIds : allInfraIds.filter(id => id !== 'pre-commit')
    return { infra }
  }

  const priorInfra = opts.prior.infra ?? allInfraIds
  const options = ALL_INFRA.map(i => ({
    value: i.value,
    label: i.label,
    hint: i.value === 'pre-commit' && !hasGit ? '⚠ No .git directory detected — will be skipped' : i.hint,
  }))

  const selectedInfra = await p.multiselect({
    message: 'Which infrastructure files do you want?',
    options,
    initialValues: priorInfra.filter(id => hasGit || id !== 'pre-commit'),
    required: false,
  })

  if (p.isCancel(selectedInfra)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  // Always exclude pre-commit if no .git regardless of selection
  const infra = (selectedInfra as InfraId[]).filter(id => hasGit || id !== 'pre-commit')

  return { infra }
}
