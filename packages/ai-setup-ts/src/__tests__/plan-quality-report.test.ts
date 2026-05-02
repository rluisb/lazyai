import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { parseSkillFrontmatter } from '../utils/frontmatter.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const planQualityLocationSchema = z
  .object({
    file: z.string().min(1),
    section: z.string().min(1).nullable(),
    lineStart: z.number().int().positive().nullable(),
    lineEnd: z.number().int().positive().nullable(),
  })
  .refine(({ file }) => !path.isAbsolute(file), 'location.file must be repo-relative')
  .refine(({ file }) => !file.startsWith('../'), 'location.file must not escape the repo')
  .refine(({ file }) => !file.includes('\\'), 'location.file must use POSIX separators')
  .refine(
    ({ lineStart, lineEnd }) => lineStart === null || lineEnd === null || lineEnd >= lineStart,
    'lineEnd must be greater than or equal to lineStart when both line numbers are available',
  )

const planQualityReportSchema = z
  .object({
    schemaVersion: z.literal('plan-quality-report/v1'),
    verdict: z.enum(['pass', 'warn', 'fail']),
    findings: z.array(
      z.object({
        rule: z.enum(['R1', 'R2', 'R3', 'R4']),
        severity: z.enum(['info', 'warn', 'fail']),
        message: z.string().min(1),
        location: planQualityLocationSchema,
      }),
    ),
    checkedAgainst: z.object({
      spec: z.string().min(1),
      plan: z.string().min(1),
      research: z.string().min(1).nullable(),
      tasks: z.null(),
    }),
  })
  .superRefine((report, ctx) => {
    const expectedVerdict = report.findings.some((finding) => finding.severity === 'fail')
      ? 'fail'
      : report.findings.some((finding) => finding.severity === 'warn')
        ? 'warn'
        : 'pass'

    if (report.verdict !== expectedVerdict) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: `verdict must match highest finding severity (${expectedVerdict})`,
        path: ['verdict'],
      })
    }
  })

describe('PlanQualityReport contract fixtures', () => {
  it.each([
    ['pass', 'pass'],
    ['warn', 'warn'],
    ['fail', 'fail'],
  ] as const)('validates the %s fixture against the report schema', (fixtureName, verdict) => {
    const report = planQualityReportSchema.parse(loadFixture(fixtureName))

    expect(report.verdict).toBe(verdict)
    expect(report.checkedAgainst.tasks).toBeNull()
    expect(report.checkedAgainst.plan).toMatch(/plan\.md$/)
    expect(report.checkedAgainst.spec).toMatch(/spec\.md$/)
  })

  it('keeps malformed or ambiguous Markdown findings as warn-only report output', () => {
    const report = planQualityReportSchema.parse(loadFixture('warn'))

    expect(report.verdict).toBe('warn')
    expect(report.findings).not.toHaveLength(0)
    expect(report.findings.every((finding) => finding.severity === 'warn')).toBe(true)
    expect(report.findings.map((finding) => finding.message).join('\n')).toMatch(/ambiguous|malformed|parser/i)
  })

  it('covers bounded R1-R4 fail recommendations without requiring tasks.md coverage', () => {
    const report = planQualityReportSchema.parse(loadFixture('fail'))
    const findingsByRule = new Map(report.findings.map((finding) => [finding.rule, finding]))

    expect([...findingsByRule.keys()].sort()).toEqual(['R1', 'R2', 'R3', 'R4'])
    expect(findingsByRule.get('R1')?.message).toMatch(/plan\.md/i)
    expect(findingsByRule.get('R1')?.message).toMatch(/does not inspect tasks\.md|tasks\.md.*do not exist/i)
    expect(report.checkedAgainst.tasks).toBeNull()
  })
})

describe('planner skill inline D6 Plan Quality Check contract', () => {
  const libraryDir = findMonorepoLibraryDir()
  const plannerSkill = readFileSync(path.join(libraryDir, 'skills', 'plan.md'), 'utf8')

  it('keeps existing planner skill frontmatter valid while adding the inline check', () => {
    const parsed = parseSkillFrontmatter(plannerSkill)

    expect(parsed.frontmatter).toMatchObject({
      name: 'plan',
      description: 'Plan implementation approach before writing code.',
      phase: 'plan',
    })
    expect(plannerSkill).toContain('## Plan Quality Check')
    expect(plannerSkill).not.toContain('plan-validate')
  })

  it('states the PlanQualityReport JSON schema and checkedAgainst.tasks null contract', () => {
    expect(plannerSkill).toContain('PlanQualityReport')
    expect(plannerSkill).toContain('"schemaVersion": "plan-quality-report/v1"')
    expect(plannerSkill).toContain('"verdict": "pass|warn|fail"')
    expect(plannerSkill).toContain('"rule": "R1|R2|R3|R4"')
    expect(plannerSkill).toContain('"severity": "info|warn|fail"')
    expect(plannerSkill).toContain('"tasks": null')
    expect(plannerSkill).toContain('"location"')
    expect(plannerSkill).toContain('"lineStart"')
    expect(plannerSkill).toContain('"lineEnd"')
  })

  it('documents the bounded R1-R4 checklist and R1 plan-only coverage scope', () => {
    expect(plannerSkill).toMatch(/R1[^\n]+spec\.md[^\n]+AC\/FR[^\n]+plan\.md/i)
    expect(plannerSkill).toMatch(/R1[^\n]+does not inspect tasks\.md/i)
    expect(plannerSkill).toMatch(/R2[^\n]+phase exit criteria/i)
    expect(plannerSkill).toMatch(/R3[^\n]+risk[^\n]+mitigation[^\n]+owner/i)
    expect(plannerSkill).toMatch(/R4[^\n]+Wave 1[^\n]+rollback/i)
  })

  it('warns on parser uncertainty and always proceeds to the human approval gate', () => {
    expect(plannerSkill).toMatch(/malformed Markdown|ambiguous structural parsing|parser uncertainty/i)
    expect(plannerSkill).toMatch(/warn[^\n]+not fail/i)
    expect(plannerSkill).toMatch(/All verdicts[^\n]+proceed[^\n]+human/i)
    expect(plannerSkill).toMatch(/fail[^\n]+recommendation/i)
    expect(plannerSkill).toMatch(/no automatic[^\n]+loop/i)
  })
})

function loadFixture(name: 'pass' | 'warn' | 'fail'): unknown {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/plan-quality-report/${name}.json`, import.meta.url), 'utf8'),
  )
}
