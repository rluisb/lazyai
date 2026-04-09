import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { OrchestratorToolHandlers } from '../tool-handlers.js'

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

  return new OrchestratorToolHandlers({
    projectRoot,
    libraryOrchestrationRoot: libraryRoot,
    libraryAgentsRoot: agentsRoot,
  })
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
})
