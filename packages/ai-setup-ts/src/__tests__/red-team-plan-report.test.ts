import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { parseSkillFrontmatter } from '../utils/frontmatter.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const redTeamCategories = [
  'scope',
  'security',
  'feasibility',
  'rollback',
  'edge-case',
  'assumption',
  'operational',
] as const

const redTeamSeverities = ['low', 'medium', 'high', 'critical'] as const

const redTeamLocationSchema = z
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

const redTeamPlanReportSchema = z.object({
  schemaVersion: z.literal('red-team-plan-report/v1'),
  status: z.enum(['ok', 'soft_fail', 'skipped']),
  findings: z.array(
    z.object({
      category: z.enum(redTeamCategories),
      severity: z.enum(redTeamSeverities),
      message: z.string().min(1),
      recommendation: z.string().min(1),
      location: redTeamLocationSchema,
    }),
  ),
})

describe('RedTeamPlanReport contract fixtures', () => {
  it.each([
    ['ok', 'ok'],
    ['soft-fail', 'soft_fail'],
    ['skipped', 'skipped'],
  ] as const)('validates the %s fixture against the report schema', (fixtureName, status) => {
    const report = redTeamPlanReportSchema.parse(loadFixture(fixtureName))

    expect(report.schemaVersion).toBe('red-team-plan-report/v1')
    expect(report.status).toBe(status)
  })

  it('covers the bounded finding categories and severity values with location metadata', () => {
    const report = redTeamPlanReportSchema.parse(loadFixture('ok'))
    const categories = new Set(report.findings.map((finding) => finding.category))
    const severities = new Set(report.findings.map((finding) => finding.severity))

    expect([...categories].sort()).toEqual([...redTeamCategories].sort())
    expect([...severities].sort()).toEqual([...redTeamSeverities].sort())
    expect(report.findings.every((finding) => finding.message.length > 0)).toBe(true)
    expect(report.findings.every((finding) => finding.recommendation.length > 0)).toBe(true)
    expect(report.findings.every((finding) => finding.location.file.endsWith('.md'))).toBe(true)
  })

  it('represents provider/API outage or unavailable red-team role as soft_fail', () => {
    const report = redTeamPlanReportSchema.parse(loadFixture('soft-fail'))
    const text = report.findings
      .map((finding) => `${finding.category} ${finding.message} ${finding.recommendation}`)
      .join('\n')

    expect(report.status).toBe('soft_fail')
    expect(report.findings).toHaveLength(1)
    expect(report.findings[0]?.category).toBe('operational')
    expect(text).toMatch(/provider|API|red-team role|unavailable|outage/i)
    expect(text).toMatch(/plan-gate|approval gate|proceed/i)
  })
})

describe('red-team-plan skill D17 contract', () => {
  const libraryDir = findMonorepoLibraryDir()
  const skillPath = path.join(libraryDir, 'skills', 'red-team-plan.md')
  const skill = readFileSync(skillPath, 'utf8')

  it('declares valid bounded skill identity without write permissions', () => {
    const parsed = parseSkillFrontmatter(skill)

    expect(parsed.frontmatter).toMatchObject({
      name: 'red-team-plan',
      phase: 'plan',
    })
    expect(parsed.frontmatter.description).toMatch(/adversarial|red-team|design/i)
    expect(skill).toMatch(/read-only/i)
    expect(skill).not.toMatch(/^\s*writes\s*:/m)
    expect(skill).not.toMatch(/^\s*workspace\s*:/m)
  })

  it('consumes plan/spec/research artifacts and produces RedTeamPlanReport', () => {
    expect(skill).toMatch(/consumes:[\s\S]*plan\.md/i)
    expect(skill).toMatch(/consumes:[\s\S]*optional spec\.md/i)
    expect(skill).toMatch(/consumes:[\s\S]*optional research\.md/i)
    expect(skill).toMatch(/produces_for:[\s\S]*plan-gate/i)
    expect(skill).toContain('RedTeamPlanReport')
    expect(skill).toContain('"schemaVersion": "red-team-plan-report/v1"')
    expect(skill).toContain('"status": "ok|soft_fail|skipped"')
    expect(skill).toContain('"category": "scope|security|feasibility|rollback|edge-case|assumption|operational"')
    expect(skill).toContain('"severity": "low|medium|high|critical"')
    expect(skill).toContain('"recommendation"')
    expect(skill).toContain('"location"')
  })

  it('restricts review to plan/spec attack and excludes implementation code review', () => {
    expect(skill).toMatch(/attack|critique/i)
    expect(skill).toMatch(/plan\.md/i)
    expect(skill).toMatch(/spec\.md/i)
    expect(skill).toMatch(/not code review|do not review code|implementation code review/i)
    expect(skill).toMatch(/scope|security|feasibility|rollback|edge-case|assumption|operational/i)
  })

  it('documents outage soft-fail behavior and continuation to plan-gate', () => {
    expect(skill).toMatch(/provider|API|tool outage|red-team role/i)
    expect(skill).toMatch(/soft_fail/i)
    expect(skill).toMatch(/must not halt|do not halt|does not halt|not a chain halt/i)
    expect(skill).toMatch(/proceed|continue/i)
    expect(skill).toMatch(/plan-gate|approval gate/i)
  })
})

function loadFixture(name: 'ok' | 'soft-fail' | 'skipped'): unknown {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/red-team-plan-report/${name}.json`, import.meta.url), 'utf8'),
  )
}
