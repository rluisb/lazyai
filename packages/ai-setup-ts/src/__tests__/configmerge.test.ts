import { mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { deepMerge, marshalSortedJson, mergeJsonFile } from '../utils/configmerge.js'
import { fileExists } from '../utils/files.js'

describe('configmerge', () => {
  const tempDirs: string[] = []

  const makeTempDir = (): string => {
    const dir = mkdtempSync(path.join(tmpdir(), 'ai-setup-configmerge-'))
    tempDirs.push(dir)
    return dir
  }

  const readJson = (p: string): Record<string, unknown> =>
    JSON.parse(readFileSync(p, 'utf-8')) as Record<string, unknown>

  afterEach(() => {
    for (const dir of tempDirs) {
      rmSync(dir, { recursive: true, force: true })
    }
    tempDirs.length = 0
  })

  it('writes a new file with no backup', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.json')
    const patch = { permissions: { allow: ['Read'] } }

    const { backupPath } = mergeJsonFile(p, patch)

    expect(backupPath).toBeNull()
    expect(readJson(p)).toEqual(patch)
  })

  it('preserves user keys and creates .bak on first touch', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.json')
    const user = {
      experimental: { foo: true },
      permissions: { allow: ['UserTool'] },
    }
    writeFileSync(p, JSON.stringify(user, null, 2))

    const patch = {
      mcpServers: { orchestrator: { command: 'ai-setup', args: ['server'] } },
    }
    const { backupPath } = mergeJsonFile(p, patch)

    expect(backupPath).not.toBeNull()
    expect(fileExists(backupPath!)).toBe(true)

    const got = readJson(p)
    expect(got.experimental).toEqual(user.experimental)
    expect(got.permissions).toEqual(user.permissions)
    expect(got.mcpServers).toEqual(patch.mcpServers)

    // .bak holds original, not merged
    expect(readJson(backupPath!)).toEqual(user)
  })

  it('is idempotent — re-run with same patch produces identical bytes; .bak never overwritten', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.json')
    const user = { existing: 'yes' }
    writeFileSync(p, JSON.stringify(user, null, 2))

    const patch = { mcpServers: { x: { command: 'c' } } }
    const { backupPath: bak1 } = mergeJsonFile(p, patch)
    const firstContent = readFileSync(p, 'utf-8')
    const firstBak = readFileSync(bak1!, 'utf-8')

    const { backupPath: bak2 } = mergeJsonFile(p, patch)
    const secondContent = readFileSync(p, 'utf-8')
    const secondBak = readFileSync(bak2!, 'utf-8')

    expect(bak1).toBe(bak2)
    expect(firstContent).toBe(secondContent)
    expect(firstBak).toBe(secondBak)

    // Third run with different patch: file updates, .bak unchanged
    const patch2 = { mcpServers: { y: { command: 'd' } } }
    mergeJsonFile(p, patch2)
    const thirdContent = readFileSync(p, 'utf-8')
    const thirdBak = readFileSync(bak1!, 'utf-8')

    expect(thirdContent).not.toBe(secondContent)
    expect(thirdBak).toBe(firstBak)
  })

  it('replaces arrays wholesale (no concatenation)', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.json')
    writeFileSync(p, JSON.stringify({ permissions: { allow: ['A', 'B'] } }, null, 2))

    const patch = { permissions: { allow: ['C'] } }
    mergeJsonFile(p, patch)

    const got = readJson(p)
    expect((got.permissions as { allow: string[] }).allow).toEqual(['C'])
  })

  it('recurses into nested objects; patch wins on leaf collisions', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.json')
    const user = { permission: { edit: 'allow', bash: 'ask' } }
    writeFileSync(p, JSON.stringify(user, null, 2))

    const patch = { permission: { bash: 'deny' } }
    mergeJsonFile(p, patch)

    const got = readJson(p)
    expect(got.permission).toEqual({ edit: 'allow', bash: 'deny' })
  })

  it('strips JSONC comments when reading existing file', () => {
    const dir = makeTempDir()
    const p = path.join(dir, 'settings.jsonc')
    writeFileSync(p, '// user comment\n{\n  "foo": "bar"\n}\n')

    mergeJsonFile(p, { baz: 'qux' })

    const got = readJson(p)
    expect(got).toEqual({ foo: 'bar', baz: 'qux' })
  })
})

describe('deepMerge', () => {
  it('returns a new object (does not mutate inputs)', () => {
    const base = { a: 1, nested: { x: 1 } }
    const patch = { nested: { y: 2 } }

    const result = deepMerge(base, patch)

    expect(result).toEqual({ a: 1, nested: { x: 1, y: 2 } })
    expect(base).toEqual({ a: 1, nested: { x: 1 } })
    expect(patch).toEqual({ nested: { y: 2 } })
  })

  it('treats arrays as leaf values', () => {
    const result = deepMerge({ list: [1, 2] }, { list: [3] })
    expect(result.list).toEqual([3])
  })

  it('handles empty base', () => {
    const result = deepMerge({}, { a: 1 })
    expect(result).toEqual({ a: 1 })
  })
})

describe('marshalSortedJson', () => {
  it('sorts keys alphabetically at every depth', () => {
    const input = { zebra: 1, alpha: { z: 1, a: 2 }, beta: [] }
    const out = marshalSortedJson(input)

    expect(out).toBe('{\n  "alpha": {\n    "a": 2,\n    "z": 1\n  },\n  "beta": [],\n  "zebra": 1\n}\n')
  })

  it('ends with a trailing newline', () => {
    const out = marshalSortedJson({ a: 1 })
    expect(out.endsWith('\n')).toBe(true)
  })
})
