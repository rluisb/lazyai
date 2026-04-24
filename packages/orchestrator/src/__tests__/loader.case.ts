import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { loadCatalog } from '../loader.js'

const tempDirs: string[] = []

afterEach(() => {
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

describe('loader', () => {
  it('loads library definitions and applies project overrides', () => {
    const projectRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-loader-project-'))
    const libraryRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-loader-library-'))
    const agentsRoot = fs.mkdtempSync(path.join(os.tmpdir(), 'orchestrator-loader-agents-'))
    tempDirs.push(projectRoot, libraryRoot, agentsRoot)

    fs.mkdirSync(path.join(libraryRoot, 'chains'), { recursive: true })
    fs.mkdirSync(path.join(libraryRoot, 'skills', 'domains'), { recursive: true })
    fs.mkdirSync(path.join(projectRoot, '.ai', 'orchestration', 'chains'), { recursive: true })
    fs.mkdirSync(path.join(projectRoot, '.ai', 'orchestration', 'skills', 'domains'), { recursive: true })
    fs.mkdirSync(agentsRoot, { recursive: true })

    fs.writeFileSync(
      path.join(agentsRoot, 'builder.md'),
      ['---', 'name: Builder', 'model: sonnet', '---', '', '# Builder', '', '## Constraints', '- Stay focused'].join('\n'),
    )
    fs.writeFileSync(
      path.join(libraryRoot, 'skills', 'domains', 'typescript.md'),
      ['---', 'name: TypeScript', 'description: Library TS', 'allowed_tools:', '- Read', '- Bash', '---', '', 'When applying this skill:', '- Use types'].join('\n'),
    )
    fs.writeFileSync(
      path.join(projectRoot, '.ai', 'orchestration', 'skills', 'domains', 'typescript.md'),
      ['---', 'name: TypeScript', 'description: Project TS', '---', '', 'You should:', '- Prefer local patterns'].join('\n'),
    )
    fs.writeFileSync(
      path.join(libraryRoot, 'chains', 'repair.json'),
      JSON.stringify({ name: 'repair', kind: 'chain', description: 'Library chain', entry: 'step-1', steps: [{ id: 'step-1', agent: 'builder', skills: [], description: 'repair', transitions: { success: 'done' } }] }),
    )
    fs.writeFileSync(
      path.join(projectRoot, '.ai', 'orchestration', 'chains', 'repair.json'),
      JSON.stringify({ name: 'repair', kind: 'chain', description: 'Project chain', entry: 'step-1', steps: [{ id: 'step-1', agent: 'builder', skills: [], description: 'repair', transitions: { success: 'done' } }] }),
    )

    const catalog = loadCatalog({
      projectRoot,
      libraryOrchestrationRoot: libraryRoot,
      libraryAgentsRoot: agentsRoot,
    })

    expect(catalog.agents.builder?.displayName).toBe('Builder')
    expect(catalog.domains.typescript?.description).toBe('Project TS')
    expect(catalog.domains.typescript?.constraints).toEqual(['Prefer local patterns'])
    expect(catalog.chains.repair?.description).toBe('Project chain')
  })
})
