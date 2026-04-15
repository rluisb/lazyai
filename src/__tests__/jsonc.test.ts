import { describe, expect, it } from 'vitest'
import { stripJsonComments } from '../utils/jsonc.js'

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
