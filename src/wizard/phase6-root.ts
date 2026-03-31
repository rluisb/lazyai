import * as p from '@clack/prompts'
import type { ToolId } from '../types.js'

export interface Phase6Result {
  rootFiles: string[] // relative paths of root files that will be created
}

/**
 * Maps each ToolId to its root configuration file(s)
 */
function getRootFilesByTool(tool: ToolId): string[] {
  switch (tool) {
    case 'opencode':
      return ['AGENTS.md']
    case 'pi':
      return ['CLAUDE.md']
    case 'claude-code':
      return ['CLAUDE.md']
    case 'gemini':
      return ['GEMINI.md']
    case 'copilot':
      return ['.github/copilot-instructions.md']
    default:
      const _exhaustive: never = tool
      return _exhaustive
  }
}

export async function runPhase6(opts: {
  interactive: boolean
  tools: ToolId[]
  projectName: string
}): Promise<Phase6Result> {
  // Compute all root files for selected tools
  const allRootFiles = opts.tools.flatMap(tool => getRootFilesByTool(tool))

  // Deduplicate using Set
  const deduplicatedRootFiles = Array.from(new Set(allRootFiles)).sort()

  if (!opts.interactive) {
    return { rootFiles: deduplicatedRootFiles }
  }

  // Interactive mode: show preview and confirm

  // Build mapping of Tool -> File(s) for display
  const toolFileMap: Array<{ tool: ToolId; files: string[] }> = opts.tools.map(
    tool => ({
      tool,
      files: getRootFilesByTool(tool),
    }),
  )

  // Format preview note
  const previewLines = [
    'Root Configuration Files:',
    '',
    ...toolFileMap.map(({ tool, files }) => `  ${tool.padEnd(12)} → ${files.join(', ')}`),
  ]

  p.note(previewLines.join('\n'), 'Preview')

  // Confirm before proceeding
  const confirmed = await p.confirm({
    message: 'Proceed with these root configuration files?',
    initialValue: true,
  })

  if (p.isCancel(confirmed) || !confirmed) {
    p.cancel('Setup cancelled.')
    process.exit(0)
  }

  return { rootFiles: deduplicatedRootFiles }
}
