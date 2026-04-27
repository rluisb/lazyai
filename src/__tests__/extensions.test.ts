import { describe, expect, it } from 'vitest'
import { existsSync, mkdirSync, writeFileSync, rmSync } from 'node:fs'
import { join } from 'node:path'
import { tmpdir } from 'node:os'
import { discoverExtensions, getExtendedAvailable } from '../extensions/discovery.js'

function makeTempDir(prefix: string): string {
  const dir = join(tmpdir(), `${prefix}-${Date.now()}`)
  mkdirSync(dir, { recursive: true })
  return dir
}

describe('discoverExtensions', () => {
  it('returns empty when no extensions configured', () => {
    const dir = makeTempDir('ai-setup-ext-empty')
    const results = discoverExtensions(dir)
    expect(results).toEqual([])
    rmSync(dir, { recursive: true, force: true })
  })

  it('discovers local extensions from .ai/extensions/', () => {
    const dir = makeTempDir('ai-setup-ext-local')
    const extDir = join(dir, '.ai', 'extensions', 'my-pack')
    mkdirSync(join(extDir, 'skills'), { recursive: true })
    mkdirSync(join(extDir, 'agents'), { recursive: true })
    writeFileSync(join(extDir, 'skills', 'custom-skill.md'), '# Custom Skill')
    writeFileSync(join(extDir, 'agents', 'my-agent.md'), '---\nname: my-agent\n---')

    const results = discoverExtensions(dir)
    expect(results.length).toBe(1)
    expect(results[0]!.name).toBe('my-pack')
    expect(results[0]!.kind).toBe('local')
    expect(results[0]!.content.skills).toEqual(['custom-skill'])
    expect(results[0]!.content.agents).toEqual(['my-agent'])
    rmSync(dir, { recursive: true, force: true })
  })

  it('discovers directory-based agents in extensions', () => {
    const dir = makeTempDir('ai-setup-ext-agentdir')
    const extDir = join(dir, '.ai', 'extensions', 'team-pack')
    const agentDir = join(extDir, 'agents', 'security-reviewer')
    mkdirSync(agentDir, { recursive: true })
    writeFileSync(join(agentDir, 'AGENT.md'), '---\nname: security-reviewer\n---')
    writeFileSync(join(agentDir, 'mcp.json'), '{}')

    const results = discoverExtensions(dir)
    expect(results.length).toBe(1)
    expect(results[0]!.content.agents).toEqual(['security-reviewer'])
    rmSync(dir, { recursive: true, force: true })
  })

  it('skips empty extension directories', () => {
    const dir = makeTempDir('ai-setup-ext-empty-dir')
    mkdirSync(join(dir, '.ai', 'extensions', 'empty-pack'), { recursive: true })

    const results = discoverExtensions(dir)
    expect(results).toEqual([])
    rmSync(dir, { recursive: true, force: true })
  })
})

describe('getExtendedAvailable', () => {
  it('returns built-in skills when no extensions', () => {
    const dir = process.cwd() // Use actual project dir for library
    if (!existsSync(join(dir, 'library'))) return // Skip if no library

    const skills = getExtendedAvailable(dir, 'skills')
    expect(skills.length).toBeGreaterThan(0)
    expect(skills).toContain('implement')
    expect(skills).toContain('research')
  })

  it('includes extension content alongside built-in', () => {
    const dir = makeTempDir('ai-setup-ext-merged')
    // Set up minimal library
    mkdirSync(join(dir, 'library', 'skills'), { recursive: true })
    writeFileSync(join(dir, 'library', 'skills', 'implement.md'), '# implement')

    // Add extension
    const extDir = join(dir, '.ai', 'extensions', 'extra')
    mkdirSync(join(extDir, 'skills'), { recursive: true })
    writeFileSync(join(extDir, 'skills', 'deploy-check.md'), '# deploy-check')

    const skills = getExtendedAvailable(dir, 'skills')
    expect(skills).toContain('implement')
    expect(skills).toContain('deploy-check')
    rmSync(dir, { recursive: true, force: true })
  })
})
