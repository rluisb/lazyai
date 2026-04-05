export interface DiffLine {
  type: 'add' | 'remove' | 'context'
  content: string
}

export function computeLineDiff(existing: string, incoming: string): DiffLine[] {
  const existingLines = existing.split('\n')
  const incomingLines = incoming.split('\n')

  const n = existingLines.length
  const m = incomingLines.length

  const dp: number[][] = Array.from({ length: n + 1 }, () => Array(m + 1).fill(0))

  for (let i = 1; i <= n; i++) {
    const currentRow = dp[i]!
    const previousRow = dp[i - 1]!

    for (let j = 1; j <= m; j++) {
      const existingLine = existingLines[i - 1]!
      const incomingLine = incomingLines[j - 1]!

      if (existingLine === incomingLine) {
        currentRow[j] = previousRow[j - 1]! + 1
      } else {
        currentRow[j] = Math.max(previousRow[j]!, currentRow[j - 1]!)
      }
    }
  }

  const reversed: DiffLine[] = []
  let i = n
  let j = m

  while (i > 0 && j > 0) {
    const existingLine = existingLines[i - 1]!
    const incomingLine = incomingLines[j - 1]!

    if (existingLine === incomingLine) {
      reversed.push({ type: 'context', content: existingLine })
      i -= 1
      j -= 1
      continue
    }

    const previousRow = dp[i - 1]!
    const currentRow = dp[i]!

    if (previousRow[j]! >= currentRow[j - 1]!) {
      reversed.push({ type: 'remove', content: existingLine })
      i -= 1
    } else {
      reversed.push({ type: 'add', content: incomingLine })
      j -= 1
    }
  }

  while (i > 0) {
    reversed.push({ type: 'remove', content: existingLines[i - 1]! })
    i -= 1
  }

  while (j > 0) {
    reversed.push({ type: 'add', content: incomingLines[j - 1]! })
    j -= 1
  }

  return reversed.reverse()
}

export function renderDiffPreview(diff: DiffLine[]): string {
  return diff
    .map(line => {
      if (line.type === 'context') return `  ${line.content}`
      if (line.type === 'add') return `+ ${line.content}`
      return `- ${line.content}`
    })
    .join('\n')
}
