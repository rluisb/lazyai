import { describe, expect, it } from 'vitest'
import { computeLineDiff, renderDiffPreview } from '../utils/diff.js'

describe('computeLineDiff', () => {
  it('returns only context lines for identical strings', () => {
    const diff = computeLineDiff('hello\nworld', 'hello\nworld')
    expect(diff.every(d => d.type === 'context')).toBe(true)
  })

  it('detects added lines', () => {
    const diff = computeLineDiff('hello', 'hello\nworld')
    expect(diff).toContainEqual({ type: 'add', content: 'world' })
  })

  it('detects removed lines', () => {
    const diff = computeLineDiff('hello\nworld', 'hello')
    expect(diff).toContainEqual({ type: 'remove', content: 'world' })
  })

  it('handles mixed changes correctly', () => {
    const diff = computeLineDiff('a\nb\nc', 'a\nx\nc')
    const types = diff.map(d => d.type)
    expect(types).toContain('remove')
    expect(types).toContain('add')
    expect(types).toContain('context')
  })

  it('handles empty existing string', () => {
    const diff = computeLineDiff('', 'hello\nworld')
    const addLines = diff.filter(d => d.type === 'add')
    expect(addLines.length).toBeGreaterThan(0)
  })

  it('handles empty incoming string', () => {
    const diff = computeLineDiff('hello\nworld', '')
    const removeLines = diff.filter(d => d.type === 'remove')
    expect(removeLines.length).toBeGreaterThan(0)
  })
})

describe('renderDiffPreview', () => {
  it('formats context lines with two-space prefix', () => {
    const result = renderDiffPreview([{ type: 'context', content: 'hello' }])
    expect(result).toBe('  hello')
  })

  it('formats add lines with + prefix', () => {
    const result = renderDiffPreview([{ type: 'add', content: 'new line' }])
    expect(result).toBe('+ new line')
  })

  it('formats remove lines with - prefix', () => {
    const result = renderDiffPreview([{ type: 'remove', content: 'old line' }])
    expect(result).toBe('- old line')
  })

  it('joins multiple lines with newlines', () => {
    const result = renderDiffPreview([
      { type: 'context', content: 'a' },
      { type: 'remove', content: 'b' },
      { type: 'add', content: 'c' },
    ])
    expect(result).toBe('  a\n- b\n+ c')
  })
})
