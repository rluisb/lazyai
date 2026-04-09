import type { ComposedAgentSpec, PromptLayer, StepOutputContract } from './types.js'

export interface ComposeAgentInput {
  base: PromptLayer
  root?: PromptLayer
  domain?: PromptLayer
  mode?: PromptLayer
  step?: PromptLayer & { outputContract?: StepOutputContract }
}

const APPROVAL_PRIORITY = ['minimal', 'normal', 'strict'] as const

export function composeAgent(input: ComposeAgentInput): ComposedAgentSpec {
  const layers = [input.root, input.base, input.domain, input.mode, input.step].filter(
    (layer): layer is PromptLayer => Boolean(layer),
  )

  const prompt = layers
    .map((layer) => layer.prompt.trim())
    .filter(Boolean)
    .join('\n\n')

  const model =
    input.step?.modelHint ?? input.mode?.modelHint ?? input.domain?.modelHint ?? input.base.modelHint ?? 'sonnet'

  const tools = intersectAllowedTools(layers)
  const constraints = dedupe(layers.flatMap((layer) => layer.constraints ?? []))
  const approvalPolicy = layers.reduce<'minimal' | 'normal' | 'strict'>((current, layer) => {
    if (!layer.approvalPolicy) return current
    return APPROVAL_PRIORITY.indexOf(layer.approvalPolicy) > APPROVAL_PRIORITY.indexOf(current)
      ? layer.approvalPolicy
      : current
  }, 'minimal')

  const composed: ComposedAgentSpec = {
    id: [input.base.name, input.domain?.name, input.mode?.name, input.step?.name].filter(Boolean).join(':'),
    base: input.base.name,
    model,
    tools,
    approvalPolicy,
    constraints,
    prompt,
    mergedFrom: layers,
  }

  if (input.domain?.name) composed.domainSkill = input.domain.name
  if (input.mode?.name) composed.modeSkill = input.mode.name
  if (input.step?.outputContract) composed.outputContract = input.step.outputContract

  return composed
}

function intersectAllowedTools(layers: PromptLayer[]): string[] {
  const sets = layers
    .map((layer) => layer.allowedTools)
    .filter((tools): tools is string[] => Array.isArray(tools) && tools.length > 0)

  if (sets.length === 0) return []
  return [...sets.reduce((acc, tools) => new Set(tools.filter((tool) => acc.has(tool))), new Set(sets[0]))].sort()
}

function dedupe(values: string[]): string[] {
  const seen = new Set<string>()
  const unique: string[] = []

  for (const value of values) {
    if (seen.has(value)) continue
    seen.add(value)
    unique.push(value)
  }

  return unique
}
