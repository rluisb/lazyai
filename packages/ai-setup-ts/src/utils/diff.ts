/**
 * Diff Utilities
 *
 * Line-level and word-level diff computation with rich terminal rendering.
 */

import * as util from 'node:util'

// styleText was added in Node.js 20.12.0 / 21.7.0
// Provide a fallback for older Node versions
const styleText: (style: string, text: string) => string =
  typeof (util as { styleText?: unknown }).styleText === 'function'
    ? (util as { styleText: (style: string, text: string) => string }).styleText
    : (_style: string, text: string) => text

export interface DiffLine {
  type: 'add' | 'remove' | 'context'
  content: string
  oldLineNum?: number
  newLineNum?: number
}

export interface DiffStats {
  additions: number
  deletions: number
  unchanged: number
}

export interface DiffResult {
  lines: DiffLine[]
  stats: DiffStats
}

function getRequired<T>(value: T | undefined, message: string): T {
  if (value === undefined) {
    throw new Error(message)
  }
  return value
}

/**
 * Compute line-level diff using LCS (Longest Common Subsequence) algorithm
 */
export function computeLineDiff(existing: string, incoming: string): DiffResult {
  const existingLines = existing.split('\n')
  const incomingLines = incoming.split('\n')

  const n = existingLines.length
  const m = incomingLines.length

  // Build LCS table
  const dp: number[][] = Array.from({ length: n + 1 }, () => Array(m + 1).fill(0))

  for (let i = 1; i <= n; i++) {
    const currentRow = getRequired(dp[i], `Missing diff row ${i}`)
    const previousRow = getRequired(dp[i - 1], `Missing diff row ${i - 1}`)

    for (let j = 1; j <= m; j++) {
      const existingLine = getRequired(existingLines[i - 1], `Missing existing line ${i - 1}`)
      const incomingLine = getRequired(incomingLines[j - 1], `Missing incoming line ${j - 1}`)

      if (existingLine === incomingLine) {
        currentRow[j] = getRequired(previousRow[j - 1], `Missing diff cell [${i - 1}, ${j - 1}]`) + 1
      } else {
        currentRow[j] = Math.max(
          getRequired(previousRow[j], `Missing diff cell [${i - 1}, ${j}]`),
          getRequired(currentRow[j - 1], `Missing diff cell [${i}, ${j - 1}]`),
        )
      }
    }
  }

  // Backtrack to build diff
  const reversed: DiffLine[] = []
  let i = n
  let j = m
  let oldLine = n
  let newLine = m

  while (i > 0 && j > 0) {
    const existingLine = getRequired(existingLines[i - 1], `Missing existing line ${i - 1}`)
    const incomingLine = getRequired(incomingLines[j - 1], `Missing incoming line ${j - 1}`)

    if (existingLine === incomingLine) {
      reversed.push({ type: 'context', content: existingLine, oldLineNum: oldLine, newLineNum: newLine })
      i -= 1
      j -= 1
      oldLine -= 1
      newLine -= 1
      continue
    }

    const previousRow = getRequired(dp[i - 1], `Missing diff row ${i - 1}`)
    const currentRow = getRequired(dp[i], `Missing diff row ${i}`)

    if (
      getRequired(previousRow[j], `Missing diff cell [${i - 1}, ${j}]`) >=
      getRequired(currentRow[j - 1], `Missing diff cell [${i}, ${j - 1}]`)
    ) {
      reversed.push({ type: 'remove', content: existingLine, oldLineNum: oldLine })
      i -= 1
      oldLine -= 1
    } else {
      reversed.push({ type: 'add', content: incomingLine, newLineNum: newLine })
      j -= 1
      newLine -= 1
    }
  }

  while (i > 0) {
    reversed.push({ type: 'remove', content: getRequired(existingLines[i - 1], `Missing existing line ${i - 1}`), oldLineNum: oldLine })
    i -= 1
    oldLine -= 1
  }

  while (j > 0) {
    reversed.push({ type: 'add', content: getRequired(incomingLines[j - 1], `Missing incoming line ${j - 1}`), newLineNum: newLine })
    j -= 1
    newLine -= 1
  }

  const lines = reversed.reverse()

  // Compute stats
  const stats: DiffStats = {
    additions: lines.filter((l) => l.type === 'add').length,
    deletions: lines.filter((l) => l.type === 'remove').length,
    unchanged: lines.filter((l) => l.type === 'context').length,
  }

  return { lines, stats }
}

/**
 * Compute word-level diff for a single line pair
 */
export function computeWordDiff(oldLine: string, newLine: string): { old: string; new: string } {
  const oldWords = oldLine.split(/(\s+)/)
  const newWords = newLine.split(/(\s+)/)

  // Simple word-level LCS
  const n = oldWords.length
  const m = newWords.length
  const dp: number[][] = Array.from({ length: n + 1 }, () => Array(m + 1).fill(0))

  for (let i = 1; i <= n; i++) {
    const currentRow = getRequired(dp[i], `Missing word diff row ${i}`)
    const previousRow = getRequired(dp[i - 1], `Missing word diff row ${i - 1}`)
    for (let j = 1; j <= m; j++) {
      if (oldWords[i - 1] === newWords[j - 1]) {
        currentRow[j] = getRequired(previousRow[j - 1], `Missing word diff cell [${i - 1}, ${j - 1}]`) + 1
      } else {
        currentRow[j] = Math.max(
          getRequired(previousRow[j], `Missing word diff cell [${i - 1}, ${j}]`),
          getRequired(currentRow[j - 1], `Missing word diff cell [${i}, ${j - 1}]`),
        )
      }
    }
  }

  // Backtrack to find common words
  const commonOld = new Set<number>()
  const commonNew = new Set<number>()
  let i = n
  let j = m

  while (i > 0 && j > 0) {
    if (oldWords[i - 1] === newWords[j - 1]) {
      commonOld.add(i - 1)
      commonNew.add(j - 1)
      i--
      j--
    } else if (
      getRequired(
        getRequired(dp[i - 1], `Missing word diff row ${i - 1}`)[j],
        `Missing word diff cell [${i - 1}, ${j}]`,
      ) >=
      getRequired(
        getRequired(dp[i], `Missing word diff row ${i}`)[j - 1],
        `Missing word diff cell [${i}, ${j - 1}]`,
      )
    ) {
      i--
    } else {
      j--
    }
  }

  // Render with highlighting
  const oldRendered = oldWords
    .map((word, idx) => (commonOld.has(idx) ? word : styleText('strikethrough', word)))
    .join('')

  const newRendered = newWords
    .map((word, idx) => (commonNew.has(idx) ? word : styleText('bold', word)))
    .join('')

  return { old: oldRendered, new: newRendered }
}

export interface RenderOptions {
  /** Show colors (default: true) */
  colors?: boolean
  /** Context lines around changes (default: 3, use -1 for all) */
  contextLines?: number
  /** Show line numbers (default: true) */
  lineNumbers?: boolean
  /** Show header with file path (default: true) */
  showHeader?: boolean
  /** File path for header */
  filePath?: string
  /** Show word-level diff for changed lines (default: true) */
  wordDiff?: boolean
}

/**
 * Find change hunks with context
 */
function findHunks(lines: DiffLine[], contextLines: number): Array<{ start: number; end: number }> {
  if (contextLines < 0) {
    return [{ start: 0, end: lines.length }]
  }

  const changeIndices: number[] = []
  for (let i = 0; i < lines.length; i++) {
    if (getRequired(lines[i], `Missing diff line ${i}`).type !== 'context') {
      changeIndices.push(i)
    }
  }

  if (changeIndices.length === 0) {
    return []
  }
  
  // Merge overlapping hunks
  const hunks: Array<{ start: number; end: number }> = []
  const firstChangeIndex = getRequired(changeIndices[0], 'Missing first change index')
  let currentStart = Math.max(0, firstChangeIndex - contextLines)
  let currentEnd = Math.min(lines.length, firstChangeIndex + contextLines + 1)

  for (let i = 1; i < changeIndices.length; i++) {
    const changeIdx = getRequired(changeIndices[i], `Missing change index ${i}`)
    const newStart = Math.max(0, changeIdx - contextLines)
    const newEnd = Math.min(lines.length, changeIdx + contextLines + 1)

    if (newStart <= currentEnd) {
      // Merge with current hunk
      currentEnd = newEnd
    } else {
      // Start new hunk
      hunks.push({ start: currentStart, end: currentEnd })
      currentStart = newStart
      currentEnd = newEnd
    }
  }

  hunks.push({ start: currentStart, end: currentEnd })
  return hunks
}

/**
 * Format line number column
 */
function formatLineNum(num: number | undefined, width: number): string {
  if (num === undefined) {
    return ' '.repeat(width)
  }
  return String(num).padStart(width, ' ')
}

/**
 * Render diff with colors, line numbers, and context collapsing
 */
export function renderDiffPreview(
  diffOrLines: DiffResult | DiffLine[],
  options: RenderOptions = {}
): string {
  const {
    colors = true,
    contextLines = 3,
    lineNumbers = true,
    showHeader = true,
    filePath,
    wordDiff = true,
  } = options

  // Handle both old API (DiffLine[]) and new API (DiffResult)
  const lines = Array.isArray(diffOrLines) ? diffOrLines : diffOrLines.lines
  const stats = Array.isArray(diffOrLines)
    ? {
        additions: lines.filter((l) => l.type === 'add').length,
        deletions: lines.filter((l) => l.type === 'remove').length,
        unchanged: lines.filter((l) => l.type === 'context').length,
      }
    : diffOrLines.stats

  if (lines.length === 0) {
    return '  (empty diff)'
  }

  const output: string[] = []

  // Header
  if (showHeader && filePath) {
    const headerLine = `─── ${filePath} ───`
    output.push(colors ? styleText('dim', headerLine) : headerLine)
  }

  // Stats line
  if (stats.additions > 0 || stats.deletions > 0) {
    const statsStr = `+${stats.additions} -${stats.deletions}`
    output.push(colors ? styleText('dim', statsStr) : statsStr)
    output.push('')
  }

  // Find hunks
  const hunks = findHunks(lines, contextLines)

  if (hunks.length === 0) {
    return '  (no changes)'
  }

  // Calculate line number width
  const maxOldLine = Math.max(...lines.flatMap((line) => (line.oldLineNum === undefined ? [] : [line.oldLineNum])), 1)
  const maxNewLine = Math.max(...lines.flatMap((line) => (line.newLineNum === undefined ? [] : [line.newLineNum])), 1)
  const lineNumWidth = Math.max(String(maxOldLine).length, String(maxNewLine).length)

  // Group consecutive add/remove pairs for word diff
  const getLinePairs = (
    hunkLines: DiffLine[]
  ): Array<{ type: 'pair'; remove: DiffLine; add: DiffLine } | { type: 'single'; line: DiffLine }> => {
    const result: Array<{ type: 'pair'; remove: DiffLine; add: DiffLine } | { type: 'single'; line: DiffLine }> = []
    let i = 0

    while (i < hunkLines.length) {
      const line = getRequired(hunkLines[i], `Missing hunk line ${i}`)
      const nextLine = i + 1 < hunkLines.length ? hunkLines[i + 1] : undefined

      // Look for remove followed by add (word diff candidate)
      if (wordDiff && line.type === 'remove' && nextLine?.type === 'add') {
        result.push({ type: 'pair', remove: line, add: nextLine })
        i += 2
      } else {
        result.push({ type: 'single', line })
        i++
      }
    }

    return result
  }

  // Render hunks
  for (let hunkIdx = 0; hunkIdx < hunks.length; hunkIdx++) {
    const hunk = getRequired(hunks[hunkIdx], `Missing hunk ${hunkIdx}`)

    // Hunk separator
    if (hunkIdx > 0) {
      const separator = '...'
      output.push(colors ? styleText('dim', separator) : separator)
    }

    const hunkLines = lines.slice(hunk.start, hunk.end)
    const pairs = getLinePairs(hunkLines)

    for (const item of pairs) {
      if (item.type === 'pair') {
        // Word-level diff for paired lines
        const { old: oldRendered, new: newRendered } = computeWordDiff(item.remove.content, item.add.content)

        // Remove line
        const oldNum = lineNumbers ? formatLineNum(item.remove.oldLineNum, lineNumWidth) : ''
        const newNumSpace = lineNumbers ? formatLineNum(undefined, lineNumWidth) : ''
        const removePrefix = lineNumbers ? `${oldNum} ${newNumSpace} │` : ''
        const removeLine = `${removePrefix}- ${oldRendered}`
        output.push(colors ? styleText('red', removeLine) : removeLine)

        // Add line
        const oldNumSpace = lineNumbers ? formatLineNum(undefined, lineNumWidth) : ''
        const newNum = lineNumbers ? formatLineNum(item.add.newLineNum, lineNumWidth) : ''
        const addPrefix = lineNumbers ? `${oldNumSpace} ${newNum} │` : ''
        const addLine = `${addPrefix}+ ${newRendered}`
        output.push(colors ? styleText('green', addLine) : addLine)
      } else {
        const line = item.line
        const oldNum = lineNumbers ? formatLineNum(line.oldLineNum, lineNumWidth) : ''
        const newNum = lineNumbers ? formatLineNum(line.newLineNum, lineNumWidth) : ''
        const prefix = lineNumbers ? `${oldNum} ${newNum} │` : ''

        let rendered: string
        switch (line.type) {
          case 'add':
            rendered = `${prefix}+ ${line.content}`
            output.push(colors ? styleText('green', rendered) : rendered)
            break
          case 'remove':
            rendered = `${prefix}- ${line.content}`
            output.push(colors ? styleText('red', rendered) : rendered)
            break
          case 'context':
            rendered = `${prefix}  ${line.content}`
            output.push(colors ? styleText('dim', rendered) : rendered)
            break
        }
      }
    }
  }

  return output.join('\n')
}

/**
 * Legacy API - simple diff preview without colors
 */
export function renderSimpleDiff(lines: DiffLine[]): string {
  return lines
    .map((line) => {
      if (line.type === 'context') return `  ${line.content}`
      if (line.type === 'add') return `+ ${line.content}`
      return `- ${line.content}`
    })
    .join('\n')
}
