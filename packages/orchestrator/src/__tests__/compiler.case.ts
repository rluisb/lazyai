import { describe, expect, it } from 'vitest'
import { compileChainDefinition, validateStepOutput } from '../compiler.js'
import type { OrchestrationCatalog } from '../types.js'

const catalog: OrchestrationCatalog = {
  agents: {
    builder: {
      kind: 'agent',
      name: 'builder',
      displayName: 'Builder',
      description: 'Builds changes',
      source: 'library',
      path: '/agents/builder.md',
      prompt: 'Base builder prompt',
      allowedTools: ['Read', 'Bash'],
      constraints: ['stay scoped'],
      modelHint: 'sonnet',
    },
  },
  domains: {
    typescript: {
      kind: 'domain',
      name: 'typescript',
      description: 'TypeScript domain',
      source: 'library',
      path: '/domains/typescript.md',
      prompt: 'Use TypeScript',
      constraints: ['use types'],
      allowedTools: ['Read', 'Bash'],
    },
  },
  modes: {
    fast: {
      kind: 'mode',
      name: 'fast',
      description: 'Fast mode',
      source: 'library',
      path: '/modes/fast.md',
      prompt: 'Move quickly',
      constraints: ['be concise'],
      allowedTools: ['Read', 'Bash'],
      approvalPolicy: 'normal',
    },
  },
  chains: {
    repair: {
      kind: 'chain',
      name: 'repair',
      description: 'Repair chain',
      source: 'library',
      path: '/chains/repair.json',
      entry: 'implement-fix',
      steps: [
        {
          id: 'implement-fix',
          agent: 'builder',
          skills: ['typescript'],
          taskType: 'adr',
          description: 'Fix the issue',
          prompt: 'Repair it',
          transitions: { success: 'done', failure: { retry: 1, then: 'done' } },
        },
      ],
    },
  },
  teams: {},
  workflows: {},
}

describe('compiler', () => {
  it('compiles a chain definition with injected skills and contracts', () => {
    const plan = compileChainDefinition({
      catalog,
      projectRoot: '/repo',
      chainName: 'repair',
      task: 'Fix a regression',
      domainSkill: 'typescript',
      modeSkill: 'fast',
    })

    expect(plan.definition.name).toBe('repair')
    expect(plan.entrypoint).toBe('implement-fix')
    expect(plan.compiledSteps?.[0]?.stepType).toBe('implement')
    expect(plan.compiledSteps?.[0]?.taskType).toBe('adr')
    expect(plan.compiledSteps?.[0]?.domainSkill).toBe('typescript')
    expect(plan.compiledSteps?.[0]?.instructions).toContain('Task Type: adr')
    expect(plan.compiledSteps?.[0]?.instructions).toContain('ADR Rules')
    expect(plan.compiledSteps?.[0]?.outputContract.requiredFields).toEqual([
      'summary',
      'status',
      'files_changed',
      'tests_passed',
    ])
  })

  it('validates structured step output', () => {
    expect(
      validateStepOutput('implement', {
        summary: 'done',
        status: 'ok',
        files_changed: ['src/index.ts'],
        tests_passed: true,
      }).valid,
    ).toBe(true)

    expect(validateStepOutput('implement', { summary: 'done' }).valid).toBe(false)
  })
})
