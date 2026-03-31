import * as p from '@clack/prompts'
import type { TemplateId, RuleId, WizardSelections } from '../types.js'

const ALL_TEMPLATES: { value: TemplateId; label: string; hint: string }[] = [
  { value: 'adr', label: 'adr', hint: 'Architecture Decision Record template' },
  { value: 'prd', label: 'prd', hint: 'Product Requirements Document template' },
  { value: 'progress', label: 'progress', hint: 'Progress tracking template' },
  { value: 'standard', label: 'standard', hint: 'Coding standard template' },
  { value: 'task', label: 'task', hint: 'Single task template' },
  { value: 'tasks', label: 'tasks', hint: 'Task list template' },
  { value: 'tech-debt', label: 'tech-debt', hint: 'Tech debt tracking template' },
  { value: 'techspec', label: 'techspec', hint: 'Technical specification template' },
]

const ALL_RULES: { value: RuleId; label: string; hint: string }[] = [
  { value: 'cost', label: 'cost', hint: 'AI cost management rules' },
  { value: 'review', label: 'review', hint: 'Code review guidelines' },
  { value: 'security', label: 'security', hint: 'Security best practices' },
  { value: 'workflow', label: 'workflow', hint: 'Development workflow rules' },
]

export interface Phase3Result {
  templates: TemplateId[]
  rules: RuleId[]
}

export async function runPhase3(opts: {
  interactive: boolean
  prior: Partial<WizardSelections>
}): Promise<Phase3Result> {
  const allTemplateIds: TemplateId[] = ALL_TEMPLATES.map(t => t.value)
  const allRuleIds: RuleId[] = ALL_RULES.map(r => r.value)

  if (!opts.interactive) {
    return { templates: allTemplateIds, rules: allRuleIds }
  }

  // Templates selection
  const priorTemplates = opts.prior.templates ?? allTemplateIds
  const selectedTemplates = await p.multiselect({
    message: 'Which document templates do you want?',
    options: ALL_TEMPLATES.map(t => ({
      value: t.value,
      label: t.label,
      hint: t.hint,
    })),
    initialValues: priorTemplates,
    required: false,
  })

  if (p.isCancel(selectedTemplates)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  // Rules selection
  const priorRules = opts.prior.rules ?? allRuleIds
  const selectedRules = await p.multiselect({
    message: 'Which AI rule sets do you want?',
    options: ALL_RULES.map(r => ({
      value: r.value,
      label: r.label,
      hint: r.hint,
    })),
    initialValues: priorRules,
    required: false,
  })

  if (p.isCancel(selectedRules)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  return {
    templates: selectedTemplates as TemplateId[],
    rules: selectedRules as RuleId[],
  }
}
