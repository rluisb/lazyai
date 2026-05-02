import { readFileSync } from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'
import { z } from 'zod'
import { parseSkillFrontmatter } from '../utils/frontmatter.js'
import { findMonorepoLibraryDir } from './test-helpers.js'

const chainVerificationRules = [
  'artifact-presence',
  'requirement-trace',
  'task-evidence',
  'test-evidence',
  'scope-boundary',
  'rollback',
] as const

const chainVerificationStatuses = ['covered', 'partial', 'missing', 'not-applicable'] as const

const repoRelativePathSchema = z
  .string()
  .min(1)
  .refine((value) => !path.isAbsolute(value), 'path must be repo-relative')
  .refine((value) => !value.startsWith('../'), 'path must not escape the repo')
  .refine((value) => !value.includes('\\'), 'path must use POSIX separators')

const artifactRefSchema = repoRelativePathSchema

const chainVerificationLocationSchema = z
  .object({
    file: repoRelativePathSchema,
    section: z.string().min(1).nullable(),
    lineStart: z.number().int().positive(),
    lineEnd: z.number().int().positive(),
  })
  .refine(
    ({ lineStart, lineEnd }) => lineEnd >= lineStart,
    'lineEnd must be greater than or equal to lineStart',
  )

const chainVerificationReportSchema = z
  .object({
    schemaVersion: z.literal('chain-verification-report/v1'),
    verdict: z.enum(['pass', 'warn', 'fail']),
    checkedArtifacts: z.object({
      spec: artifactRefSchema.nullable(),
      plan: artifactRefSchema.nullable(),
      tasks: artifactRefSchema.nullable(),
      taskFiles: z.array(artifactRefSchema),
      implementationEvidence: z.array(artifactRefSchema),
      tests: z.array(artifactRefSchema),
    }),
    traceability: z.array(
      z.object({
        requirementId: z.string().min(1),
        planRefs: z.array(artifactRefSchema),
        taskRefs: z.array(artifactRefSchema),
        evidenceRefs: z.array(artifactRefSchema),
        status: z.enum(chainVerificationStatuses),
      }),
    ),
    findings: z.array(
      z.object({
        rule: z.enum(chainVerificationRules),
        severity: z.enum(['info', 'warn', 'fail']),
        message: z.string().min(1),
        recommendation: z.string().min(1),
        location: chainVerificationLocationSchema,
      }),
    ),
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

describe('ChainVerificationReport contract fixtures', () => {
  it.each([
    ['pass', 'pass'],
    ['warn', 'warn'],
    ['fail', 'fail'],
    ['ambiguous-warn', 'warn'],
  ] as const)('validates the %s fixture against the report schema', (fixtureName, verdict) => {
    const report = chainVerificationReportSchema.parse(loadFixture(fixtureName))

    expect(report.schemaVersion).toBe('chain-verification-report/v1')
    expect(report.verdict).toBe(verdict)
    expect(report.checkedArtifacts.spec).toMatch(/spec-wave2\.md$/)
    expect(report.checkedArtifacts.plan).toMatch(/plan-wave2\.md$/)
    expect(report.traceability.length).toBeGreaterThan(0)
  })

  it('covers checked artifacts, traceability statuses, finding rules, and repo-relative locations', () => {
    const reports = [
      chainVerificationReportSchema.parse(loadFixture('pass')),
      chainVerificationReportSchema.parse(loadFixture('warn')),
      chainVerificationReportSchema.parse(loadFixture('fail')),
    ]
    const statuses = new Set(reports.flatMap((report) => report.traceability.map((trace) => trace.status)))
    const rules = new Set(reports.flatMap((report) => report.findings.map((finding) => finding.rule)))

    expect([...statuses].sort()).toEqual([...chainVerificationStatuses].sort())
    expect([...rules].sort()).toEqual([...chainVerificationRules].sort())
    expect(reports.every((report) => report.checkedArtifacts.taskFiles.every((file) => file.endsWith('.md')))).toBe(true)
    expect(reports.every((report) => report.findings.every((finding) => finding.location.file.endsWith('.md')))).toBe(true)
  })

  it('treats missing optional artifacts as warn-only output', () => {
    const report = chainVerificationReportSchema.parse(loadFixture('warn'))

    expect(report.verdict).toBe('warn')
    expect(report.checkedArtifacts.tasks).toBeNull()
    expect(report.checkedArtifacts.taskFiles).toEqual([])
    expect(report.checkedArtifacts.implementationEvidence).toEqual([])
    expect(report.checkedArtifacts.tests).toEqual([])
    expect(report.findings.every((finding) => finding.severity !== 'fail')).toBe(true)
    expect(report.findings.map((finding) => finding.message).join('\n')).toMatch(/missing optional|optional artifact/i)
  })

  it('treats malformed or ambiguous artifact parsing as warn findings, not parser-driven fail', () => {
    const report = chainVerificationReportSchema.parse(loadFixture('ambiguous-warn'))

    expect(report.verdict).toBe('warn')
    expect(report.findings).not.toHaveLength(0)
    expect(report.findings.every((finding) => finding.severity === 'warn')).toBe(true)
    expect(report.findings.map((finding) => finding.message).join('\n')).toMatch(/malformed|ambiguous|parser/i)
  })
})

describe('chain-verify skill D4 contract', () => {
  const libraryDir = findMonorepoLibraryDir()
  const skillPath = path.join(libraryDir, 'skills', 'chain-verify.md')
  const skill = readFileSync(skillPath, 'utf8')

  it('declares valid bounded skill identity without write permissions', () => {
    const parsed = parseSkillFrontmatter(skill)

    expect(parsed.frontmatter).toMatchObject({
      name: 'chain-verify',
      phase: 'review',
      output: 'ChainVerificationReport',
    })
    expect(parsed.frontmatter.description).toMatch(/verification|trace/i)
    expect(skill).toMatch(/read-only/i)
    expect(skill).not.toMatch(/^\s*writes\s*:/m)
    expect(skill).not.toMatch(/^\s*workspace\s*:/m)
  })

  it('consumes implementation and review artifacts and produces ChainVerificationReport', () => {
    expect(skill).toMatch(/consumes:[\s\S]*spec\.md/i)
    expect(skill).toMatch(/consumes:[\s\S]*plan\.md/i)
    expect(skill).toMatch(/consumes:[\s\S]*tasks\.md/i)
    expect(skill).toMatch(/consumes:[\s\S]*task files/i)
    expect(skill).toMatch(/consumes:[\s\S]*implementation/i)
    expect(skill).toMatch(/consumes:[\s\S]*tests/i)
    expect(skill).toContain('ChainVerificationReport')
    expect(skill).toContain('"schemaVersion": "chain-verification-report/v1"')
    expect(skill).toContain('"verdict": "pass|warn|fail"')
    expect(skill).toContain('"checkedArtifacts"')
    expect(skill).toContain('"traceability"')
    expect(skill).toContain('"findings"')
    expect(skill).toContain('"rule": "artifact-presence|requirement-trace|task-evidence|test-evidence|scope-boundary|rollback"')
    expect(skill).toContain('"status": "covered|partial|missing|not-applicable"')
    expect(skill).toContain('"location"')
  })

  it('documents bounded warning semantics and excludes chain/runtime integration', () => {
    expect(skill).toMatch(/missing optional artifacts?[\s\S]*warn/i)
    expect(skill).toMatch(/malformed|ambiguous|parser uncertainty/i)
    expect(skill).toMatch(/warn[\s\S]*not fail/i)
    expect(skill).toMatch(/human approval|human reviewer/i)
    expect(skill).toMatch(/do not edit[\s\S]*feature\.json/i)
    expect(skill).toMatch(/do not[\s\S]*runtime engine|runtime engine[\s\S]*do not/i)
    expect(skill).toMatch(/no chain integration|do not integrate/i)
  })
})

function loadFixture(name: 'pass' | 'warn' | 'fail' | 'ambiguous-warn'): unknown {
  return JSON.parse(
    readFileSync(new URL(`./fixtures/chain-verification-report/${name}.json`, import.meta.url), 'utf8'),
  )
}
