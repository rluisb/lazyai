import { describe, expect, it } from 'vitest'
import { computeLineDiff, computeWordDiff, renderDiffPreview, renderSimpleDiff } from '../utils/diff.js'

describe('computeLineDiff', () => {
  it('returns only context lines for identical strings', () => {
    const diff = computeLineDiff('hello\nworld', 'hello\nworld')
    expect(diff.lines.every(d => d.type === 'context')).toBe(true)
  })

  it('detects added lines', () => {
    const diff = computeLineDiff('hello', 'hello\nworld')
    expect(diff.lines).toContainEqual(expect.objectContaining({ type: 'add', content: 'world' }))
  })

  it('detects removed lines', () => {
    const diff = computeLineDiff('hello\nworld', 'hello')
    expect(diff.lines).toContainEqual(expect.objectContaining({ type: 'remove', content: 'world' }))
  })

  it('handles mixed changes correctly', () => {
    const diff = computeLineDiff('a\nb\nc', 'a\nx\nc')
    const types = diff.lines.map(d => d.type)
    expect(types).toContain('remove')
    expect(types).toContain('add')
    expect(types).toContain('context')
  })

  it('handles empty existing string', () => {
    const diff = computeLineDiff('', 'hello\nworld')
    const addLines = diff.lines.filter(d => d.type === 'add')
    expect(addLines.length).toBeGreaterThan(0)
  })

  it('handles empty incoming string', () => {
    const diff = computeLineDiff('hello\nworld', '')
    const removeLines = diff.lines.filter(d => d.type === 'remove')
    expect(removeLines.length).toBeGreaterThan(0)
  })

  it('computes correct stats', () => {
    const diff = computeLineDiff('a\nb\nc', 'a\nx\nc')
    expect(diff.stats.additions).toBe(1)
    expect(diff.stats.deletions).toBe(1)
    expect(diff.stats.unchanged).toBe(2)
  })

  it('includes line numbers', () => {
    const diff = computeLineDiff('a\nb', 'a\nc')
    const removeLine = diff.lines.find(l => l.type === 'remove')
    const addLine = diff.lines.find(l => l.type === 'add')
    expect(removeLine?.oldLineNum).toBeDefined()
    expect(addLine?.newLineNum).toBeDefined()
  })
})

describe('renderSimpleDiff', () => {
  it('formats context lines with two-space prefix', () => {
    const result = renderSimpleDiff([{ type: 'context', content: 'hello' }])
    expect(result).toBe('  hello')
  })

  it('formats add lines with + prefix', () => {
    const result = renderSimpleDiff([{ type: 'add', content: 'new line' }])
    expect(result).toBe('+ new line')
  })

  it('formats remove lines with - prefix', () => {
    const result = renderSimpleDiff([{ type: 'remove', content: 'old line' }])
    expect(result).toBe('- old line')
  })

  it('joins multiple lines with newlines', () => {
    const result = renderSimpleDiff([
      { type: 'context', content: 'a' },
      { type: 'remove', content: 'b' },
      { type: 'add', content: 'c' },
    ])
    expect(result).toBe('  a\n- b\n+ c')
  })
})

describe('renderDiffPreview', () => {
  it('renders with all options disabled for simple output', () => {
    const diff = computeLineDiff('a\nb', 'a\nc')
    const result = renderDiffPreview(diff, {
      colors: false,
      lineNumbers: false,
      showHeader: false,
      wordDiff: false,
      contextLines: -1,
    })
    expect(result).toContain('  a')
    expect(result).toContain('- b')
    expect(result).toContain('+ c')
  })

  it('includes header when filePath provided', () => {
    const diff = computeLineDiff('a', 'b')
    const result = renderDiffPreview(diff, {
      colors: false,
      lineNumbers: false,
      showHeader: true,
      filePath: 'test.md',
      contextLines: -1,
    })
    expect(result).toContain('test.md')
  })

  it('includes stats line', () => {
    const diff = computeLineDiff('old', 'new')
    const result = renderDiffPreview(diff, {
      colors: false,
      lineNumbers: false,
      showHeader: false,
      contextLines: -1,
    })
    expect(result).toContain('+1 -1')
  })

  it('includes line numbers when enabled', () => {
    const diff = computeLineDiff('a\nb', 'a\nc')
    const result = renderDiffPreview(diff, {
      colors: false,
      lineNumbers: true,
      showHeader: false,
      contextLines: -1,
    })
    expect(result).toContain('│')
  })

  it('handles empty diff', () => {
    const result = renderDiffPreview({ lines: [], stats: { additions: 0, deletions: 0, unchanged: 0 } })
    expect(result).toBe('  (empty diff)')
  })

  it('collapses context with hunk separator', () => {
    // Create a diff with changes far apart
    const existing = 'a\nb\nc\nd\ne\nf\ng\nh\ni\nj\nk'
    const incoming = 'a\nX\nc\nd\ne\nf\ng\nh\ni\nj\nY'
    const diff = computeLineDiff(existing, incoming)
    const result = renderDiffPreview(diff, {
      colors: false,
      lineNumbers: false,
      showHeader: false,
      contextLines: 2,
    })
    expect(result).toContain('...')
  })
})

describe('computeWordDiff', () => {
  it('identifies changed words', () => {
    const result = computeWordDiff('hello world', 'hello universe')
    expect(result.old).toContain('world')
    expect(result.new).toContain('universe')
  })

  it('preserves common words', () => {
    const result = computeWordDiff('the quick fox', 'the slow fox')
    expect(result.old).toContain('the')
    expect(result.new).toContain('the')
    expect(result.old).toContain('fox')
    expect(result.new).toContain('fox')
  })
})
