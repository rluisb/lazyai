import { z } from 'zod'

// Shared base — all definitions have at least a name
const baseSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
})

export const agentFrontmatterSchema = baseSchema.extend({
  model: z.string().optional(),
  mode: z.enum(['primary', 'subagent', 'review']).optional(),
  temperature: z.number().min(0).max(2).optional(),
  steps: z.number().int().positive().optional(),
  tools: z.record(z.boolean()).optional(),
  permissions: z.record(z.unknown()).optional(),
})

export const skillFrontmatterSchema = baseSchema.extend({
  kind: z.enum(['domain', 'mode']).optional(),
  allowed_tools: z.array(z.string()).optional(),
  model_hint: z.string().optional(),
  approval_policy: z.enum(['minimal', 'normal', 'strict']).optional(),
  applies_to: z.array(z.string()).optional(),
})

export const chainBodySchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  kind: z.literal('chain'),
  entry: z.string().min(1),
  steps: z.array(z.object({
    id: z.string().min(1),
    agent: z.string().min(1),
    skills: z.array(z.string()),
    description: z.string(),
    transitions: z.record(z.union([z.string(), z.object({ retry: z.number(), then: z.string() })])),
    gate: z.enum(['user_approval', 'severity_confirmation', 'cost_confirmation']).optional(),
    prompt: z.string().optional(),
    allowedTools: z.array(z.string()).optional(),
    model: z.string().optional(),
  })),
  domain_skill_injection: z.enum(['all_steps', 'builder_steps_only', 'none']).optional(),
  mode_skill_injection: z.enum(['all_steps', 'builder_steps_only', 'none']).optional(),
})

export const teamBodySchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  kind: z.literal('team'),
  budget_multiplier: z.number().optional(),
  user_confirmation_required: z.boolean().optional(),
  parallel: z.array(z.object({
    role: z.string(),
    agent: z.string(),
    skills: z.array(z.string()),
    focus: z.string(),
  })),
  synthesize: z.object({
    agent: z.string(),
    description: z.string(),
  }),
})

export const workflowBodySchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
  kind: z.literal('workflow'),
  entry: z.string().min(1),
  phases: z.array(z.object({
    id: z.string(),
    kind: z.enum(['chain', 'team', 'gate', 'terminal']),
    ref: z.string().optional(),
    gate: z.enum(['user_approval', 'severity_confirmation', 'cost_confirmation']).optional(),
    prompt: z.string().optional(),
    when: z.string().optional(),
    on: z.record(z.string()).optional(),
  })),
})

export const commandFrontmatterSchema = baseSchema.extend({
  agent: z.string().optional(),
  subtask: z.boolean().optional(),
})

export type AgentFrontmatter = z.infer<typeof agentFrontmatterSchema>
export type SkillFrontmatter = z.infer<typeof skillFrontmatterSchema>

export type CatalogKindExtended = 'agent' | 'skill' | 'chain' | 'team' | 'workflow' | 'mode' | 'command'

export function validateFrontmatterForKind(
  kind: CatalogKindExtended,
  frontmatter: Record<string, unknown>,
): { valid: true } | { valid: false; issues: string[] } {
  let schema: z.ZodTypeAny
  switch (kind) {
    case 'agent':
      schema = agentFrontmatterSchema
      break
    case 'skill':
    case 'mode':
      schema = skillFrontmatterSchema
      break
    case 'command':
      schema = commandFrontmatterSchema
      break
    case 'chain':
    case 'team':
    case 'workflow':
      return { valid: true }
  }

  const result = schema.safeParse(frontmatter)
  if (result.success) return { valid: true }
  return {
    valid: false,
    issues: result.error.issues.map((i) => `${i.path.join('.')}: ${i.message}`),
  }
}
