import { mkdtempSync, rmSync, writeFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import path from 'node:path'
import { afterEach, describe, expect, it } from 'vitest'
import { parseJsonc, readJsoncFile, stripJsonComments } from '../utils/jsonc.js'

describe('stripJsonComments', () => {
  it('strips simple line comments', () => {
    const input = `{
  "a": 1 // trailing comment
}`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ a: 1 })
  })

  it('strips block comments', () => {
    const input = `/* leading */ { "a": /* mid */ 1 }`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ a: 1 })
  })

  it('preserves // inside string literals (URLs)', () => {
    const input = `{ "$schema": "https://opencode.ai/config.json" }`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ $schema: 'https://opencode.ai/config.json' })
  })

  it('preserves /* inside string literals', () => {
    const input = `{ "pattern": "/*.ts" }`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ pattern: '/*.ts' })
  })

  it('handles escaped quotes inside strings', () => {
    const input = `{ "msg": "say \\"hi\\" // now", "x": 1 }`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ msg: 'say "hi" // now', x: 1 })
  })

  it('is a no-op when there are no comments and no strings', () => {
    const input = `{"a":1,"b":[2,3]}`
    expect(stripJsonComments(input)).toBe(input)
  })

  it('handles multiline block comments crossing keys', () => {
    const input = `{
  "a": 1,
  /*
   * "b": 99,
   */
  "c": 3
}`
    const out = stripJsonComments(input)
    expect(JSON.parse(out)).toEqual({ a: 1, c: 3 })
  })

  it('handles a full opencode.jsonc shape with URL + comments', () => {
    const input = `{
  // top-level comment
  "$schema": "https://opencode.ai/config.json",
  "mcp": {
    "orchestrator": {
      "type": "local",
      "command": ["node", "./orchestrator/dist/index.js"] // local dogfood
    }
  }
}`
    const parsed = JSON.parse(stripJsonComments(input)) as {
      $schema: string
      mcp: { orchestrator: { type: string; command: string[] } }
    }
    expect(parsed.$schema).toBe('https://opencode.ai/config.json')
    expect(parsed.mcp.orchestrator.command).toEqual(['node', './orchestrator/dist/index.js'])
  })
})

describe('parseJsonc', () => {
  it('parses JSONC into a plain object', () => {
    const input = `{
  // config
  "$schema": "https://opencode.ai/config.json",
  "permission": { "edit": "ask" }
}`
    expect(parseJsonc(input)).toEqual({
      $schema: 'https://opencode.ai/config.json',
      permission: { edit: 'ask' },
    })
  })

  it('throws on invalid JSON after stripping comments', () => {
    expect(() => parseJsonc('{ not: valid }')).toThrow()
  })
})

describe('readJsoncFile', () => {
  const tempDirs: string[] = []

  afterEach(() => {
    for (const dir of tempDirs) {
      rmSync(dir, { recursive: true, force: true })
    }
    tempDirs.length = 0
  })

  it('reads a JSONC file from disk and parses it', () => {
    const dir = mkdtempSync(path.join(tmpdir(), 'ai-setup-jsonc-'))
    tempDirs.push(dir)
    const filePath = path.join(dir, 'config.jsonc')
    writeFileSync(filePath, '// leading\n{ "foo": 1, "bar": [2, 3] }\n')

    expect(readJsoncFile(filePath)).toEqual({ foo: 1, bar: [2, 3] })
  })
})
