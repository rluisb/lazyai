import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createProgram } from '../cli.js'

function makeTempRepo(): string {
  const tempDir = fs.mkdtempSync(path.join(os.tmpdir(), 'ai-setup-orchestration-cli-'))
  fs.mkdirSync(path.join(tempDir, '.git'), { recursive: true })
  return tempDir
}

async function runCli(args: string[]): Promise<void> {
  const program = createProgram()
  await program.parseAsync(['node', 'ai-setup', ...args])
}

async function runJsonCommand(args: string[]): Promise<unknown> {
  const logSpy = vi.spyOn(console, 'log').mockImplementation(() => {})

  try {
    await runCli(args)
    const output = logSpy.mock.calls.map((call) => call.map((value) => String(value)).join(' ')).join('\n')
    expect(output).not.toBe('')
    return JSON.parse(output)
  } finally {
    logSpy.mockRestore()
  }
}

describe('orchestration CLI commands', () => {
  let originalCwd: string

  beforeEach(() => {
    originalCwd = process.cwd()
  })

  afterEach(() => {
    process.chdir(originalCwd)
  })

  it('create workflow writes orchestration workflow JSON', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(['create', 'workflow', 'custom-flow', '--chain', 'feature', '--team', 'review-team', '--no-interactive'])

    const workflowPath = path.join(tempDir, '.ai/orchestration/workflows/custom-flow.json')
    expect(fs.existsSync(workflowPath)).toBe(true)

    const workflow = JSON.parse(fs.readFileSync(workflowPath, 'utf-8')) as {
      kind: string
      name: string
      phases: Array<{ kind: string; ref?: string }>
    }

    expect(workflow.kind).toBe('workflow')
    expect(workflow.name).toBe('custom-flow')
    expect(workflow.phases.some((phase) => phase.kind === 'chain' && phase.ref === 'feature')).toBe(true)
    expect(workflow.phases.some((phase) => phase.kind === 'team' && phase.ref === 'review-team')).toBe(true)
  })

  it('create domain and mode write orchestration skill files', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(['create', 'domain', 'backend-go', '--description', 'Go backend patterns', '--no-interactive'])
    await runCli(['create', 'mode', 'strict-review', '--description', 'High-friction approval mode', '--no-interactive'])

    const domainPath = path.join(tempDir, '.ai/orchestration/skills/domains/backend-go.md')
    const modePath = path.join(tempDir, '.ai/orchestration/skills/modes/strict-review.md')

    expect(fs.readFileSync(domainPath, 'utf-8')).toContain('kind: domain-skill')
    expect(fs.readFileSync(domainPath, 'utf-8')).toContain('name: backend-go')
    expect(fs.readFileSync(modePath, 'utf-8')).toContain('kind: mode-skill')
    expect(fs.readFileSync(modePath, 'utf-8')).toContain('name: strict-review')
  })

  it('list merges project-local orchestration artifacts with library defaults', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(['create', 'workflow', 'custom-flow', '--no-interactive'])
    await runCli(['create', 'domain', 'backend-go', '--no-interactive'])

    const workflows = await runJsonCommand(['list', 'workflows', '--json']) as {
      workflows: Array<{ name: string; source: string }>
    }
    const domains = await runJsonCommand(['list', 'domains', '--json']) as {
      domains: Array<{ name: string; source: string }>
    }
    const orchestration = await runJsonCommand(['list', 'orchestration', '--json']) as {
      workflows: Array<{ name: string }>
      domains: Array<{ name: string }>
      chains: Array<{ name: string }>
      teams: Array<{ name: string }>
      modes: Array<{ name: string }>
    }

    expect(workflows.workflows.some((item) => item.name === 'custom-flow' && item.source === 'project')).toBe(true)
    expect(workflows.workflows.some((item) => item.name === 'rpi' && item.source === 'library')).toBe(true)
    expect(domains.domains.some((item) => item.name === 'backend-go' && item.source === 'project')).toBe(true)
    expect(domains.domains.some((item) => item.name === 'backend' && item.source === 'library')).toBe(true)
    expect(orchestration.chains.some((item) => item.name === 'feature')).toBe(true)
    expect(orchestration.teams.some((item) => item.name === 'review-team')).toBe(true)
    expect(orchestration.modes.some((item) => item.name === 'senior')).toBe(true)
  })

  it('info returns structured details for orchestration artifacts', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(['create', 'workflow', 'custom-flow', '--chain', 'feature', '--no-interactive'])

    const workflowInfo = await runJsonCommand(['info', 'custom-flow', '--json']) as {
      type: string
      source: string
      data: { kind: string; entry: string }
    }
    const teamInfo = await runJsonCommand(['info', 'review-team', '--json']) as {
      type: string
      source: string
      data: { kind: string; parallel: unknown[] }
    }
    const domainInfo = await runJsonCommand(['info', 'backend', '--json']) as {
      type: string
      source: string
      metadata: Record<string, unknown>
    }

    expect(workflowInfo.type).toBe('workflow')
    expect(workflowInfo.source).toBe('project')
    expect(workflowInfo.data.kind).toBe('workflow')
    expect(workflowInfo.data.entry).toBeTruthy()

    expect(teamInfo.type).toBe('team')
    expect(teamInfo.source).toBe('library')
    expect(teamInfo.data.kind).toBe('team')
    expect(teamInfo.data.parallel.length).toBeGreaterThan(1)

    expect(domainInfo.type).toBe('domain')
    expect(domainInfo.source).toBe('library')
    expect(domainInfo.metadata.kind).toBe('domain-skill')
  })

  it('orchestration namespace delegates create, list, and status', async () => {
    const tempDir = makeTempRepo()
    process.chdir(tempDir)

    await runCli(['orchestration', 'create', 'domain', 'payments', '--description', 'Payments domain', '--no-interactive'])

    const createdPath = path.join(tempDir, '.ai/orchestration/skills/domains/payments.md')
    expect(fs.existsSync(createdPath)).toBe(true)

    const workflows = await runJsonCommand(['orchestration', 'list', 'workflows', '--json']) as {
      workflows: Array<{ name: string }>
    }
    const status = await runJsonCommand(['orchestration', 'status', '--json']) as {
      scaffolded: boolean
      project: Record<string, number>
      library: Record<string, number>
    }

    expect(workflows.workflows.some((item) => item.name === 'rpi')).toBe(true)
    expect(status.scaffolded).toBe(true)
    expect(status.project.domains).toBe(1)
    expect(status.library.workflows).toBeGreaterThan(0)
  })
})
