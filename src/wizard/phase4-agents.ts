import * as p from '@clack/prompts'
import type { AgentId, SkillId, PromptId, WizardSelections } from '../types.js'

const ALL_AGENTS: { value: AgentId; label: string; hint: string }[] = [
  { value: 'builder', label: 'builder', hint: 'Implements features and writes code' },
  { value: 'documenter', label: 'documenter', hint: 'Creates and maintains documentation' },
  { value: 'planner', label: 'planner', hint: 'Breaks down tasks and creates plans' },
  { value: 'red-team', label: 'red-team', hint: 'Reviews for security and edge cases' },
  { value: 'reviewer', label: 'reviewer', hint: 'Code review and quality checks' },
  { value: 'scout', label: 'scout', hint: 'Explores codebase and gathers context' },
]

const ALL_SKILLS: { value: SkillId; label: string; hint: string }[] = [
  { value: 'implement', label: 'implement', hint: 'Step-by-step implementation workflow' },
  { value: 'iterate', label: 'iterate', hint: 'Iterative refinement workflow' },
  { value: 'plan', label: 'plan', hint: 'Planning and task breakdown workflow' },
  { value: 'research', label: 'research', hint: 'Research and exploration workflow' },
]

const ALL_PROMPTS: { value: PromptId; label: string; hint: string }[] = [
  { value: 'compact', label: 'compact', hint: 'Compact context prompt' },
  { value: 'implement', label: 'implement', hint: 'Implementation prompt template' },
  { value: 'local-example', label: 'local-example', hint: 'Local example prompt' },
  { value: 'plan', label: 'plan', hint: 'Planning prompt template' },
  { value: 'research', label: 'research', hint: 'Research prompt template' },
]

export interface Phase4Result {
  agents: AgentId[]
  skills: SkillId[]
  prompts: PromptId[]
}

export async function runPhase4(opts: {
  interactive: boolean
  prior: Partial<WizardSelections>
}): Promise<Phase4Result> {
  const allAgentIds: AgentId[] = ALL_AGENTS.map(a => a.value)
  const allSkillIds: SkillId[] = ALL_SKILLS.map(s => s.value)
  const allPromptIds: PromptId[] = ALL_PROMPTS.map(p2 => p2.value)

  if (!opts.interactive) {
    return { agents: allAgentIds, skills: allSkillIds, prompts: allPromptIds }
  }

  // Agents selection
  const priorAgents = opts.prior.agents ?? allAgentIds
  const selectedAgents = await p.multiselect({
    message: 'Which AI agents do you want to install?',
    options: ALL_AGENTS.map(a => ({
      value: a.value,
      label: a.label,
      hint: a.hint,
    })),
    initialValues: priorAgents,
    required: false,
  })

  if (p.isCancel(selectedAgents)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  // Skills selection
  const priorSkills = opts.prior.skills ?? allSkillIds
  const selectedSkills = await p.multiselect({
    message: 'Which skill workflows do you want?',
    options: ALL_SKILLS.map(s => ({
      value: s.value,
      label: s.label,
      hint: s.hint,
    })),
    initialValues: priorSkills,
    required: false,
  })

  if (p.isCancel(selectedSkills)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  // Prompts selection
  const priorPrompts = opts.prior.prompts ?? allPromptIds
  const selectedPrompts = await p.multiselect({
    message: 'Which prompt templates do you want?',
    options: ALL_PROMPTS.map(pr => ({
      value: pr.value,
      label: pr.label,
      hint: pr.hint,
    })),
    initialValues: priorPrompts,
    required: false,
  })

  if (p.isCancel(selectedPrompts)) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  return {
    agents: selectedAgents as AgentId[],
    skills: selectedSkills as SkillId[],
    prompts: selectedPrompts as PromptId[],
  }
}
