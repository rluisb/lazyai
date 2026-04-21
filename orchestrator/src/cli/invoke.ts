import type { Db } from '../db/index.js'
import { OrchestratorToolHandlers } from '../tool-handlers.js'

export const INVOKE_HELP = `Usage: ai-setup-orchestrator invoke <agent> <task> [options]

Resolve a named agent from the catalog and print its composed prompt spec.

Arguments:
  <agent>            Agent name (as stored in the catalog)
  <task>             Task description to pass to the agent

Options:
  --domain <name>    Domain skill to apply
  --mode   <name>    Mode skill to apply
  --project <path>   Project root (default: cwd)
  -h, --help         Show this help
`

export interface InvokeCliOptions {
  agent?: string
  task?: string
  domainSkill?: string
  modeSkill?: string
  projectRoot?: string
  help?: boolean
}

export function parseInvokeArgs(args: string[]): InvokeCliOptions {
  const opts: InvokeCliOptions = {}
  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === '--domain' && args[i + 1]) { opts.domainSkill = args[++i] as string; continue }
    if (arg === '--mode' && args[i + 1]) { opts.modeSkill = args[++i] as string; continue }
    if (arg === '--project' && args[i + 1]) { opts.projectRoot = args[++i] as string; continue }
    if (arg === '-h' || arg === '--help') { opts.help = true; continue }
    if (!opts.agent) { opts.agent = arg as string; continue }
    if (!opts.task) { opts.task = arg as string; continue }
  }
  return opts
}

export async function runInvoke(_db: Db, args: string[], out: NodeJS.WriteStream = process.stdout): Promise<void> {
  const opts = parseInvokeArgs(args)

  if (opts.help || !opts.agent || !opts.task) {
    out.write(INVOKE_HELP)
    return
  }

  const handlers = new OrchestratorToolHandlers({
    projectRoot: opts.projectRoot ?? process.cwd(),
  })

  const result = handlers.invokeAgent({
    agent: opts.agent,
    task: opts.task,
    ...(opts.domainSkill !== undefined ? { domainSkill: opts.domainSkill } : {}),
    ...(opts.modeSkill !== undefined ? { modeSkill: opts.modeSkill } : {}),
  })

  out.write(JSON.stringify(result, null, 2) + '\n')
}
