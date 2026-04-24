import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { clearScanCache } from '../catalog/host-scanner.js'
import { scanClaudeProject } from '../catalog/host-scanner.js'

const tempDirs: string[] = []

afterEach(() => {
  clearScanCache()
  for (const dir of tempDirs.splice(0)) {
    fs.rmSync(dir, { recursive: true, force: true })
  }
})

function makeTempDir(prefix: string): string {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), prefix))
  tempDirs.push(dir)
  return dir
}

describe('host-scanner — claude project scan', () => {
  it('reads agents from <project>/.claude/agents/', () => {
    const projectRoot = makeTempDir('orchestrator-scanner-')
    const agentsDir = path.join(projectRoot, '.claude', 'agents')
    fs.mkdirSync(agentsDir, { recursive: true })
    fs.writeFileSync(
      path.join(agentsDir, 'Reviewer.md'),
      ['---', 'name: Reviewer', 'model: sonnet', '---', '', '# Reviewer', '', 'Review the code.'].join('\n'),
    )

    const scan = scanClaudeProject(projectRoot)
    expect(scan.agents.reviewer).toBeDefined()
    expect(scan.agents.reviewer?.source).toBe('user_project')
    expect(scan.agents.reviewer?.modelHint).toBe('sonnet')
    expect(scan.agents.reviewer?.displayName).toBe('Reviewer')
  })

  it('reads skills from <project>/.claude/skills/ — flat .md layout', () => {
    const projectRoot = makeTempDir('orchestrator-scanner-skill-')
    const skillsDir = path.join(projectRoot, '.claude', 'skills')
    fs.mkdirSync(skillsDir, { recursive: true })
    fs.writeFileSync(
      path.join(skillsDir, 'typescript.md'),
      ['---', 'name: TypeScript', 'description: TS skill', '---', '', 'When applying this skill:', '- Use strict types'].join('\n'),
    )

    const scan = scanClaudeProject(projectRoot)
    expect(scan.skills.typescript).toBeDefined()
    expect(scan.skills.typescript?.source).toBe('user_project')
    expect(scan.skills.typescript?.constraints).toContain('Use strict types')
  })

  it('reads skills from <project>/.claude/skills/ — directory SKILL.md layout', () => {
    const projectRoot = makeTempDir('orchestrator-scanner-skill-dir-')
    const skillDir = path.join(projectRoot, '.claude', 'skills', 'implement')
    fs.mkdirSync(skillDir, { recursive: true })
    fs.writeFileSync(
      path.join(skillDir, 'SKILL.md'),
      ['---', 'name: implement', 'description: Implement skill', '---', '', 'You should:', '- Write tests first'].join('\n'),
    )

    const scan = scanClaudeProject(projectRoot)
    expect(scan.skills.implement).toBeDefined()
    expect(scan.skills.implement?.constraints).toContain('Write tests first')
  })

  it('returns empty catalog when .claude/ does not exist', () => {
    const projectRoot = makeTempDir('orchestrator-scanner-empty-')
    const scan = scanClaudeProject(projectRoot)
    expect(Object.keys(scan.agents)).toHaveLength(0)
    expect(Object.keys(scan.skills)).toHaveLength(0)
  })

  it('uses mtime cache — second call with unchanged dir returns same reference', () => {
    const projectRoot = makeTempDir('orchestrator-scanner-cache-')
    const agentsDir = path.join(projectRoot, '.claude', 'agents')
    fs.mkdirSync(agentsDir, { recursive: true })
    fs.writeFileSync(
      path.join(agentsDir, 'builder.md'),
      '---\nname: builder\n---\n\nBuilds.',
    )
    const first = scanClaudeProject(projectRoot)
    const second = scanClaudeProject(projectRoot)
    expect(first).toBe(second)
  })
})
