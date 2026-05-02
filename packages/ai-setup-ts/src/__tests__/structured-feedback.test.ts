import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { findMonorepoLibraryDir } from './test-helpers.js'

const repoRelativePathSchema = z
  .string()
  .min(1)
  .refine((value) => !path.isAbsolute(value), 'path must be repo-relative')
  .refine((value) => !value.startsWith('../'), 'path must not escape the repo')
  .refine((value) => !value.includes('\\'), 'path must use POSIX separators')

const feedbackLocationSchema = z
  .object({
    file: repoRelativePathSchema,
    section: z.string().min(1).nullable(),
    lineStart: z.number().int().positive().nullable(),
    lineEnd: z.number().int().positive().nullable(),
  })
  .refine(
    ({ lineStart, lineEnd }) => lineStart === null || lineEnd === null || lineEnd >= lineStart,
    'lineEnd must be greater than or equal to lineStart when both line numbers are available',
  )

const requiredChangeSchema = z.object({
  id: z.string().min(1),
  description: z.string().min(1),
  priority: z.enum(['blocking', 'high', 'medium', 'low']),
  target: z.string().min(1),
  evidence: z.string().min(1),
  location: feedbackLocationSchema.nullable(),
  recommendedNextAction: z.string().min(1),
  blocksProgress: z.boolean(),
})

const suggestionSchema = z.object({
  description: z.string().min(1),
  priority: z.enum(['medium', 'low']),
  target: z.string().min(1),
})

const structuredFeedbackSchema = z
  .object({
    schemaVersion: z.literal('structured-feedback/v1'),
    verdict: z.enum(['approved', 'request_changes', 'rejected', 'comment']),
    summary: z.string().min(1),
    requiredChanges: z.array(requiredChangeSchema),
    suggestions: z.array(suggestionSchema),
    requestedBy: z.enum(['human', 'reviewer', 'red-team', 'planner']),
    targetPhaseOrStep: z.string().min(1).nullable(),
  })
  .superRefine((feedback, ctx) => {
    if (
      (feedback.verdict === 'request_changes' || feedback.verdict === 'rejected') &&
      feedback.requiredChanges.length === 0
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'request_changes and rejected feedback require at least one requiredChanges item',
        path: ['requiredChanges'],
      })
    }

    feedback.requiredChanges.forEach((change, index) => {
      if (change.blocksProgress && change.priority === 'low') {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'blocking feedback cannot use low priority',
          path: ['requiredChanges', index, 'priority'],
        })
      }
    })
  })

describe('StructuredFeedback contract fixtures', () => {
  it.each([
    ['request-changes', 'request_changes'],
    ['comment', 'comment'],
  ] as const)('validates the %s fixture against the schema', (fixtureName, verdict) => {
    const feedback = structuredFeedbackSchema.parse(loadFixture(fixtureName))

    expect(feedback.schemaVersion).toBe('structured-feedback/v1')
    expect(feedback.verdict).toBe(verdict)
  })

  it('captures actionable required changes with source, priority, evidence/location, next action, and blocking flag', () => {
    const feedback = structuredFeedbackSchema.parse(loadFixture('request-changes'))
    const change = feedback.requiredChanges[0]

    expect(feedback.requestedBy).toBe('human')
    expect(feedback.targetPhaseOrStep).toBe('plan-gate')
    expect(change?.priority).toBe('blocking')
    expect(change?.target).toMatch(/phase|task|file/i)
    expect(change?.evidence).toMatch(/line|section|command/i)
    expect(change?.location?.file).toMatch(/\.md$/)
    expect(change?.recommendedNextAction).toMatch(/revise|clarify|fix/i)
    expect(change?.blocksProgress).toBe(true)
  })

  it('rejects request_changes feedback when required-change detail is missing', () => {
    expect(() => structuredFeedbackSchema.parse(loadFixture('missing-required-change'))).toThrow(
      /requiredChanges/i,
    )
  })

  it('rejects blocking feedback downgraded to low priority', () => {
    expect(() => structuredFeedbackSchema.parse(loadFixture('low-priority-blocker'))).toThrow(/low priority/i)
  })
})

describe('structured feedback static guidance', () => {
  const libraryDir = findMonorepoLibraryDir()
  const rule = readFileSync(path.join(libraryDir, 'rules', 'structured-feedback.md'), 'utf8')
  const iterateSkill = readFileSync(path.join(libraryDir, 'skills', 'iterate.md'), 'utf8')
  const orchestrateSkill = readFileSync(path.join(libraryDir, 'skills', 'orchestrate.md'), 'utf8')
  const combinedGuidance = `${rule}\n${iterateSkill}\n${orchestrateSkill}`

  it('documents the StructuredFeedback schema and bounded actionable fields', () => {
    expect(rule).toContain('StructuredFeedback')
    expect(rule).toContain('"schemaVersion": "structured-feedback/v1"')
    expect(rule).toContain('"verdict": "approved|request_changes|rejected|comment"')
    expect(rule).toContain('"requiredChanges"')
    expect(rule).toContain('"suggestions"')
    expect(rule).toContain('"requestedBy": "human|reviewer|red-team|planner"')
    expect(rule).toContain('"targetPhaseOrStep"')
    expect(rule).toMatch(/source|requestedBy/i)
    expect(rule).toMatch(/severity|priority/i)
    expect(rule).toMatch(/finding|description/i)
    expect(rule).toMatch(/action|recommended next action/i)
    expect(rule).toMatch(/evidence/i)
    expect(rule).toMatch(/location/i)
    expect(rule).toMatch(/blocksProgress|blocks progress/i)
  })

  it('tells iterate and orchestrate to consume structured feedback and ask clarification on ambiguity', () => {
    for (const content of [iterateSkill, orchestrateSkill]) {
      expect(content).toMatch(/StructuredFeedback/i)
      expect(content).toMatch(/required changes/i)
      expect(content).toMatch(/suggestions/i)
      expect(content).toMatch(/priority/i)
      expect(content).toMatch(/evidence/i)
      expect(content).toMatch(/target phase|target task|targetPhaseOrStep/i)
      expect(content).toMatch(/clarification/i)
      expect(content).toMatch(/do not guess|must not guess/i)
    }
  })

  it('keeps D9 static and excludes T021 runtime feedback propagation claims', () => {
    expect(combinedGuidance).toMatch(/static|prompt|guidance/i)
    expect(combinedGuidance).toMatch(/T021|separate approval/i)
    expect(combinedGuidance).not.toMatch(/chain-machine\.ts|tool-handlers\.ts|ChainState|StepState/i)
    expect(combinedGuidance).not.toMatch(/new gate engine|runtime conditionals|telemetry|model routing|parallel blocks/i)
  })
})

function loadFixture(
  name: 'request-changes' | 'comment' | 'missing-required-change' | 'low-priority-blocker',
): unknown {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/structured-feedback/${name}.json`, import.meta.url), 'utf8'),
  )
}
