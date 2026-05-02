import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { findMonorepoLibraryDir } from './test-helpers.js'

const completionCriterionSchema = z.object({
  id: z.string().min(1),
  description: z.string().min(1),
  evidence: z.string().min(1).nullable(),
  met: z.boolean(),
})

const completionEnforcementReportSchema = z
  .object({
    schemaVersion: z.literal('completion-enforcement-report/v1'),
    status: z.enum(['done', 'blocked', 'not-done']),
    criteria: z.array(completionCriterionSchema).min(1),
    blockers: z.array(z.string().min(1)),
    outOfScopeChanges: z.array(z.string().min(1)),
  })
  .superRefine((report, ctx) => {
    if (report.status === 'done') {
      report.criteria.forEach((criterion, index) => {
        if (!criterion.met || criterion.evidence === null) {
          ctx.addIssue({
            code: z.ZodIssueCode.custom,
            message: 'done status requires every criterion to be met with evidence',
            path: ['criteria', index],
          })
        }
      })
      if (report.blockers.length > 0) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'done status cannot include unresolved blockers',
          path: ['blockers'],
        })
      }
    }

    if (report.status === 'blocked' && report.blockers.length === 0) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'blocked status requires at least one blocker',
        path: ['blockers'],
      })
    }
  })

describe('CompletionEnforcementReport contract', () => {
  it('accepts done reports only when every criterion has evidence', () => {
    const report = completionEnforcementReportSchema.parse({
      schemaVersion: 'completion-enforcement-report/v1',
      status: 'done',
      criteria: [
        {
          id: 'DW-001',
          description: 'Implement guidance requires evidence for every Done When item.',
          evidence: 'packages/ai-setup-go/library/skills/implement.md#completion-enforcement-checklist',
          met: true,
        },
      ],
      blockers: [],
      outOfScopeChanges: [],
    })

    expect(report.status).toBe('done')
  })

  it('rejects done reports with missing evidence or unresolved blockers', () => {
    expect(() =>
      completionEnforcementReportSchema.parse({
        schemaVersion: 'completion-enforcement-report/v1',
        status: 'done',
        criteria: [{ id: 'DW-001', description: 'Missing evidence example.', evidence: null, met: true }],
        blockers: ['Quality gate unavailable.'],
        outOfScopeChanges: [],
      }),
    ).toThrow(/evidence|blockers/i)
  })

  it('allows blocked reports when unmet criteria include an explicit blocker', () => {
    const report = completionEnforcementReportSchema.parse({
      schemaVersion: 'completion-enforcement-report/v1',
      status: 'blocked',
      criteria: [{ id: 'DW-001', description: 'Run quality gate.', evidence: null, met: false }],
      blockers: ['Dependency install is unavailable in the approved worktree.'],
      outOfScopeChanges: [],
    })

    expect(report.blockers).toHaveLength(1)
  })
})

describe('TillDone completion enforcement guidance', () => {
  const libraryDir = findMonorepoLibraryDir()
  const implementSkill = readFileSync(path.join(libraryDir, 'skills', 'implement.md'), 'utf8')
  const iterateSkill = readFileSync(path.join(libraryDir, 'skills', 'iterate.md'), 'utf8')
  const reviewSkill = readFileSync(path.join(libraryDir, 'skills', 'review.md'), 'utf8')
  const workflowRule = readFileSync(path.join(libraryDir, 'rules', 'workflow.md'), 'utf8')

  it('requires implement and iterate to prove every Done When item before declaring done', () => {
    for (const content of [implementSkill, iterateSkill]) {
      expect(content).toContain('## Completion Enforcement Checklist')
      expect(content).toMatch(/Done When/i)
      expect(content).toMatch(/every task Done When item/i)
      expect(content).toMatch(/evidence/i)
      expect(content).toMatch(/verification evidence|quality gates/i)
      expect(content).toMatch(/CompletionEnforcementReport/i)
      expect(content).toMatch(/status.*done.*blocked.*not-done/i)
      expect(content).toMatch(/blocker/i)
    }
  })

  it('requires review to detect early stops, missing evidence, and scope drift', () => {
    expect(reviewSkill).toContain('## Early-Stop Completion Check')
    expect(reviewSkill).toMatch(/early-stop/i)
    expect(reviewSkill).toMatch(/acceptance criteria|Done When/i)
    expect(reviewSkill).toMatch(/missing evidence/i)
    expect(reviewSkill).toMatch(/tests\/quality gates|quality gates/i)
    expect(reviewSkill).toMatch(/out-of-scope changes|scope drift/i)
    expect(reviewSkill).toMatch(/REQUEST_CHANGES/i)
  })

  it('preserves one-task-per-session boundaries and documents blockers instead of overrunning scope', () => {
    expect(workflowRule).toMatch(/one task per session/i)
    expect(workflowRule).toMatch(/approved task/i)
    expect(workflowRule).toMatch(/does not authorize scope expansion/i)
    expect(workflowRule).toMatch(/documented blocker|blocker\/handoff/i)
    expect(workflowRule).toMatch(/unresolved risks\/assumptions|risks and assumptions/i)
    expect(iterateSkill).toMatch(/max 5 iterations/i)
    expect(iterateSkill).toMatch(/STOP and escalate|documented blocker/i)
  })

  it('stays static and prompt-only without runtime automation promises', () => {
    const combinedGuidance = [implementSkill, iterateSkill, reviewSkill, workflowRule].join('\n')

    expect(combinedGuidance).toMatch(/static|guidance|checklist/i)
    expect(combinedGuidance).not.toMatch(/runtime conditionals/i)
    expect(combinedGuidance).not.toMatch(/auto-retry/i)
    expect(combinedGuidance).not.toMatch(/auto-recovery/i)
    expect(combinedGuidance).not.toMatch(/telemetry/i)
    expect(combinedGuidance).not.toMatch(/model routing/i)
    expect(combinedGuidance).not.toMatch(/\bRAG\b/i)
    expect(combinedGuidance).not.toMatch(/parallel blocks/i)
  })
})
