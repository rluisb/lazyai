import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { loadChainState } from '../persistence.js'
import { OrchestratorToolHandlers } from '../tool-handlers.js'
import type { ChainStepStatus } from '../types.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function setupFixture() {
  const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-handlers-project-'))
  const libraryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-handlers-library-'))
  const agentsRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-handlers-agents-'))
  tempDirs.push(projectRoot, libraryRoot, agentsRoot)

  fs.mkdirSync(path.join(libraryRoot, 'chains'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'teams'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'workflows'), { recursive: true })
  fs.mkdirSync(path.join(libraryRoot, 'skills', 'domains'), { recursive: true })
  fs.mkdirSync(agentsRoot, { recursive: true })

  fs.writeFileSync(
    path.join(agentsRoot, 'builder.md'),
    ['---', 'name: Builder', 'model: sonnet', '---', '', '# Builder', '', '## Constraints', '- Stay scoped'].join('\n'),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'skills', 'domains', 'typescript.md'),
    ['---', 'name: TypeScript', 'description: TS domain', '---', '', 'When applying this skill:', '- Prefer exact types'].join('\n'),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'chains', 'repair.json'),
    JSON.stringify({
      name: 'repair',
      kind: 'chain',
      description: 'Repair chain',
      entry: 'implement-fix',
      steps: [
        {
          id: 'implement-fix',
          agent: 'builder',
          skills: ['typescript'],
          description: 'Fix the issue',
          prompt: 'Repair it',
          transitions: { success: 'done', failure: { retry: 1, then: 'done' } },
        },
      ],
    }),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'teams', 'review-team.json'),
    JSON.stringify({
      name: 'review-team',
      kind: 'team',
      description: 'Review in parallel',
      version: '1.0.0',
      parallel: [
        {
          role: 'correctness-reviewer',
          agent: 'builder',
          skills: ['typescript'],
          focus: 'Correctness',
        },
      ],
      synthesize: {
        agent: 'builder',
        description: 'Merge findings',
      },
    }),
  )
  fs.writeFileSync(
    path.join(libraryRoot, 'workflows', 'delivery.json'),
    JSON.stringify({
      name: 'delivery',
      kind: 'workflow',
      description: 'Deliver with implementation and review',
      version: '1.0.0',
      entry: 'implement',
      phases: [
        {
          id: 'implement',
          kind: 'chain',
          ref: 'repair',
          on: {
            success: 'review',
            failure: 'handoff',
          },
        },
        {
          id: 'review',
          kind: 'team',
          ref: 'review-team',
          on: {
            success: 'complete',
            failure: 'handoff',
          },
        },
        {
          id: 'handoff',
          kind: 'terminal',
        },
        {
          id: 'complete',
          kind: 'terminal',
        },
      ],
    }),
  )

  return new OrchestratorToolHandlers({
    projectRoot,
    libraryOrchestrationRoot: libraryRoot,
    libraryAgentsRoot: agentsRoot,
  })
}

function isChainStepStatus(current: unknown): current is ChainStepStatus {
  return typeof current === 'object' && current !== null && 'stepId' in current
}

describe('tool-handlers', () => {
  it('lists catalog items and composes prompts', () => {
    const handlers = setupFixture()

    const catalog = handlers.listCatalog({ query: 'repair' })
    const composed = handlers.composeAgent({
      base: 'builder',
      domainSkill: 'typescript',
      stepInstructions: 'Fix the bug',
    })

    expect(catalog.items.map((item) => item.name)).toContain('repair')
    expect(composed.prompt).toContain('Fix the bug')
    expect(composed.domainSkill).toBe('typescript')
  })

  it('builds team runs and completes workflow child phases', () => {
    const handlers = setupFixture()

    const team = handlers.buildTeam({
      team: 'review-team',
      task: 'Review auth middleware',
    })

    expect(team.state).toBe('running')
    expect(team.readyTaskIds).toEqual(['review-team:correctness-reviewer'])

    handlers.assignTask({
      teamId: team.teamId,
      taskId: 'review-team:correctness-reviewer',
      assignee: 'reviewer-a',
      claim: true,
    })
    handlers.completeTask({
      teamId: team.teamId,
      taskId: 'review-team:correctness-reviewer',
      outcome: 'success',
      result: { summary: 'Looks good' },
      usage: { totalTokens: 25 },
    })
    const synthesized = handlers.completeTask({
      teamId: team.teamId,
      taskId: 'review-team:synthesize',
      outcome: 'success',
      result: { summary: 'Merged', verdict: 'pass' },
      usage: { totalTokens: 10 },
    })

    expect(synthesized.state).toBe('completed')

    const workflow = handlers.startWorkflow({
      workflow: 'delivery',
      task: 'Deliver auth middleware',
    })

    expect(workflow.currentPhase?.phaseId).toBe('implement')
    expect(workflow.currentPhase?.childRun?.runKind).toBe('chain')

    const chainId = workflow.currentPhase?.childRun?.runId
    if (!chainId) throw new Error('expected child chain id')

    const chainStatus = handlers.getStatus({ runId: chainId, kind: 'chain' })
    if (chainStatus.kind !== 'chain') throw new Error('expected chain status')
    const currentStep = chainStatus.current
    if (!isChainStepStatus(currentStep)) throw new Error('expected chain step status')
    const currentStepId = currentStep?.stepId
    if (!currentStepId) throw new Error('expected child chain step id')

    handlers.advanceChain({
      chainId,
      stepId: currentStepId,
      outcome: 'success',
      output: {
        summary: 'done',
        status: 'ok',
        files_changed: ['src/auth.ts'],
        tests_passed: true,
      },
      usage: { totalTokens: 30 },
    })

    const workflowAfterChain = handlers.advanceWorkflow({
      workflowId: workflow.workflowId,
      outcome: 'success',
    })

    expect(workflowAfterChain.currentPhase?.phaseId).toBe('review')
    expect(workflowAfterChain.currentPhase?.childRun?.runKind).toBe('team')

    const budget = handlers.getBudget({ runId: workflowAfterChain.workflowId, kind: 'workflow' })
    expect(budget.scope).toBe('workflow')
    expect(budget.tokens.consumed).toBeGreaterThanOrEqual(30)
  })

  it('attaches bootstrap and housekeeping reports to chain state', () => {
    const handlers = setupFixture()

    const started = handlers.startChain({
      chain: 'repair',
      task: 'Fix auth regression',
    })

    const initial = handlers.getStatus({ runId: started.chainId, kind: 'chain' })
    if (initial.kind !== 'chain') throw new Error('expected chain status')
    expect(loadChainState(projectRootFromHandlers(handlers), started.chainId).bootstrapReport?.memoryPath).toBe('specs/memory')

    const current = initial.current
    if (!isChainStepStatus(current)) throw new Error('expected chain step status')

    handlers.advanceChain({
      chainId: started.chainId,
      stepId: current.stepId,
      outcome: 'success',
      output: {
        summary: 'done',
        status: 'ok',
        files_changed: ['src/auth.ts'],
        tests_passed: true,
      },
    })

    const completed = handlers.getStatus({ runId: started.chainId, kind: 'chain' })
    if (completed.kind !== 'chain') throw new Error('expected chain status')
    expect(loadChainState(projectRootFromHandlers(handlers), started.chainId).steps[0]?.housekeepingReport?.phase).toBe('combined')
  })
})

function projectRootFromHandlers(handlers: OrchestratorToolHandlers): string {
  return (handlers as unknown as { options: { projectRoot: string } }).options.projectRoot
}
