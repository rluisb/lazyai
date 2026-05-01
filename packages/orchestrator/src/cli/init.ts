import fs from 'node:fs'
import path from 'node:path'
import { getCliContext } from '../compiler.js'
import { loadRuntimeCatalog } from '../catalog/runtime.js'
import type { Db } from '../db/index.js'
import type { CliContext, HostCli, OrchestrationCatalog } from '../types.js'

export const INIT_HELP = `Usage: ai-setup-orchestrator init [options]

Inspect the local project, host CLI capability, and orchestration catalog,
then recommend a deterministic orchestration path for an optional task.

Options:
  --task <text>       Task/request to recommend orchestration for
  --host <host>       claude-code | opencode | copilot (default: opencode)
  --project <path>    Project root (default: cwd)
  --json              Machine-readable JSON output
  --verbose           Include definition source/path details
  -h, --help          Show this help
`

const HOSTS: HostCli[] = ['claude-code', 'opencode', 'copilot']

export interface InitCliOptions {
  task?: string
  host?: HostCli
  projectRoot?: string
  json?: boolean
  verbose?: boolean
  help?: boolean
  libraryOrchestrationRoot?: string
  libraryAgentsRoot?: string
  db?: Db
}

export interface InitRecommendation {
  kind: 'direct-agent' | 'chain' | 'team' | 'workflow'
  name?: string
  confidence: 'low' | 'medium' | 'high'
  reason: string
  nextCommand?: string
  alternatives: Array<{ kind: string; name?: string; reason: string }>
}

export interface InitReport {
  projectRoot: string
  host: CliContext
  rootFiles: { agentsMd: boolean; claudeMd: boolean }
  inventory: {
    agents: string[]
    domains: string[]
    modes: string[]
    chains: string[]
    teams: string[]
    workflows: string[]
  }
  recommendation?: InitRecommendation
}

type TaskClass = 'review' | 'design' | 'build' | 'general'
type RecommendationKind = InitRecommendation['kind']

export function parseInitArgs(args: string[]): InitCliOptions {
  const opts: InitCliOptions = {}

  for (let i = 0; i < args.length; i++) {
    const arg = args[i]
    if (arg === '--task') {
      const value = args[i + 1]
      if (!value) throw new Error('--task <text> is required')
      opts.task = value
      i++
      continue
    }
    if (arg === '--host') {
      const value = args[i + 1]
      if (!value) throw new Error('--host <host> is required')
      opts.host = parseHost(value)
      i++
      continue
    }
    if (arg === '--project') {
      const value = args[i + 1]
      if (!value) throw new Error('--project <path> is required')
      opts.projectRoot = value
      i++
      continue
    }
    if (arg === '--json') { opts.json = true; continue }
    if (arg === '--verbose') { opts.verbose = true; continue }
    if (arg === '-h' || arg === '--help') { opts.help = true; continue }
    throw new Error(`Unknown init argument: ${arg}`)
  }

  return opts
}

export async function runInit(options: InitCliOptions, out: NodeJS.WritableStream = process.stdout): Promise<void> {
  if (options.help) {
    out.write(INIT_HELP)
    return
  }

  const projectRoot = path.resolve(options.projectRoot ?? process.cwd())
  const hostName = options.host ?? 'opencode'
  const host = getCliContext(hostName)
  const catalog = loadRuntimeCatalog({
    projectRoot,
    hostCli: hostName,
    ...(options.libraryOrchestrationRoot ? { libraryOrchestrationRoot: options.libraryOrchestrationRoot } : {}),
    ...(options.libraryAgentsRoot ? { libraryAgentsRoot: options.libraryAgentsRoot } : {}),
    ...(options.db ? { db: options.db } : {}),
  })

  const report = buildInitReport({
    projectRoot,
    host,
    catalog,
    ...(options.task ? { task: options.task } : {}),
  })

  if (options.json) {
    out.write(`${JSON.stringify(report, null, 2)}\n`)
    return
  }

  out.write(formatHumanReport(report, catalog, Boolean(options.verbose)))
}

function parseHost(value: string): HostCli {
  if ((HOSTS as string[]).includes(value)) return value as HostCli
  throw new Error(`Unknown host: ${value}`)
}

function buildInitReport(input: {
  projectRoot: string
  host: CliContext
  catalog: OrchestrationCatalog
  task?: string
}): InitReport {
  const inventory = {
    agents: sortedKeys(input.catalog.agents),
    domains: sortedKeys(input.catalog.domains),
    modes: sortedKeys(input.catalog.modes),
    chains: sortedKeys(input.catalog.chains),
    teams: sortedKeys(input.catalog.teams),
    workflows: sortedKeys(input.catalog.workflows),
  }
  const recommendation = input.task
    ? recommendOrchestration(input.task, input.host, input.catalog, inventory)
    : undefined

  return {
    projectRoot: input.projectRoot,
    host: input.host,
    rootFiles: {
      agentsMd: fs.existsSync(path.join(input.projectRoot, 'AGENTS.md')),
      claudeMd: fs.existsSync(path.join(input.projectRoot, 'CLAUDE.md')),
    },
    inventory,
    ...(recommendation ? { recommendation } : {}),
  }
}

function recommendOrchestration(
  task: string,
  host: CliContext,
  catalog: OrchestrationCatalog,
  inventory: InitReport['inventory'],
): InitRecommendation {
  const taskClass = classifyTask(task)

  if (taskClass === 'review') {
    const team = host.supportsParallelTeams ? pickTeam(catalog, inventory.teams, ['review', 'audit', 'security', 'red-team']) : undefined
    if (team) {
      return recommendation('team', team, 'high', 'Review/audit work benefits from parallel specialist review, and this host supports teams.', task, catalog, inventory)
    }

    const agent = pickName(inventory.agents, ['reviewer', 'red-team', 'security', 'auditor'])
    if (agent) {
      return recommendation('direct-agent', agent, 'medium', 'Review/audit work was detected, but no suitable parallel team is available for this host.', task, catalog, inventory)
    }

    const chain = pickName(inventory.chains, ['review', 'audit', 'security']) ?? first(inventory.chains)
    if (chain) {
      return recommendation('chain', chain, 'medium', 'Review/audit work was detected; a chain is the best available catalog fit.', task, catalog, inventory)
    }
  }

  if (taskClass === 'design') {
    const workflow = pickWorkflow(catalog, inventory.workflows, ['rpi', 'feature', 'design', 'architect', 'plan'])
    if (workflow) {
      return recommendation('workflow', workflow, 'high', 'Design/architecture/from-scratch work benefits from a staged workflow before implementation.', task, catalog, inventory)
    }

    const chain = pickChain(catalog, inventory.chains, ['rpi', 'feature', 'design', 'architect', 'plan'])
    if (chain) {
      return recommendation('chain', chain, 'high', 'Design/architecture/from-scratch work benefits from a planned multi-step chain.', task, catalog, inventory)
    }

    const agent = pickName(inventory.agents, ['architect', 'planner', 'implementor-senior'])
    if (agent) {
      return recommendation('direct-agent', agent, 'medium', 'Design work was detected, but no workflow or chain is available.', task, catalog, inventory)
    }
  }

  if (taskClass === 'build') {
    const workflow = pickWorkflow(catalog, inventory.workflows, ['rpi', 'feature', 'implement', 'delivery'])
    if (workflow) {
      return recommendation('workflow', workflow, 'high', 'Build/implement/refactor work matches an RPI-style workflow in the catalog.', task, catalog, inventory)
    }

    const chain = pickChain(catalog, inventory.chains, ['rpi', 'feature', 'implement', 'refactor', 'repair'])
    if (chain) {
      return recommendation('chain', chain, 'high', 'Build/implement/refactor work matches a multi-step implementation chain.', task, catalog, inventory)
    }

    const agent = pickName(inventory.agents, ['implementor-senior', 'implementor'])
    if (agent) {
      return recommendation('direct-agent', agent, 'medium', 'Build/implement/refactor work was detected, but no workflow or chain is available.', task, catalog, inventory)
    }
  }

  const workflow = first(inventory.workflows)
  if (workflow) return recommendation('workflow', workflow, 'low', 'No specific task pattern matched; using the first available workflow.', task, catalog, inventory)
  const chain = first(inventory.chains)
  if (chain) return recommendation('chain', chain, 'low', 'No specific task pattern matched; using the first available chain.', task, catalog, inventory)
  const agent = pickName(inventory.agents, ['implementor-senior', 'implementor', 'architect', 'reviewer']) ?? first(inventory.agents)
  return recommendation('direct-agent', agent, 'low', 'No specific task pattern matched; using the best available direct agent.', task, catalog, inventory)
}

function classifyTask(task: string): TaskClass {
  const normalized = task.toLowerCase()
  if (/\b(review|audit|security|vulnerability|threat|red[- ]?team|compliance)\b/.test(normalized)) return 'review'
  if (/\b(design|architecture|architect|plan|planner|from scratch|greenfield)\b/.test(normalized)) return 'design'
  if (/\b(build|implement|refactor|fix|feature|migrate|add|update|repair)\b/.test(normalized)) return 'build'
  return 'general'
}

function recommendation(
  kind: RecommendationKind,
  name: string | undefined,
  confidence: InitRecommendation['confidence'],
  reason: string,
  task: string,
  catalog: OrchestrationCatalog,
  inventory: InitReport['inventory'],
): InitRecommendation {
  return {
    kind,
    ...(name ? { name } : {}),
    confidence,
    reason,
    ...(name ? { nextCommand: buildNextCommand(kind, name, task) } : {}),
    alternatives: buildAlternatives(kind, name, catalog, inventory),
  }
}

function buildAlternatives(
  selectedKind: RecommendationKind,
  selectedName: string | undefined,
  catalog: OrchestrationCatalog,
  inventory: InitReport['inventory'],
): InitRecommendation['alternatives'] {
  const alternatives: InitRecommendation['alternatives'] = []
  const add = (kind: RecommendationKind, name: string | undefined, reason: string): void => {
    if (!name) return
    if (kind === selectedKind && name === selectedName) return
    alternatives.push({ kind, name, reason })
  }

  add('workflow', pickWorkflow(catalog, inventory.workflows, ['rpi', 'feature']) ?? first(inventory.workflows), 'Use a workflow when the task needs staged orchestration.')
  add('chain', pickChain(catalog, inventory.chains, ['rpi', 'feature', 'review']) ?? first(inventory.chains), 'Use a chain for deterministic sequential execution.')
  add('team', pickTeam(catalog, inventory.teams, ['review', 'feature', 'team']) ?? first(inventory.teams), 'Use a team when parallel specialist review is appropriate and the host supports it.')
  add('direct-agent', pickName(inventory.agents, ['implementor-senior', 'implementor', 'architect', 'reviewer']) ?? first(inventory.agents), 'Use a direct agent for focused single-agent work.')

  return alternatives.slice(0, 3)
}

function buildNextCommand(kind: RecommendationKind, name: string, task: string): string {
  const taskPart = `task=${JSON.stringify(task)}`
  if (kind === 'workflow') return `Use MCP tool start_workflow with workflow=${JSON.stringify(name)} and ${taskPart}`
  if (kind === 'chain') return `Use MCP tool start_chain with chain=${JSON.stringify(name)} and ${taskPart}`
  if (kind === 'team') return `Use MCP tool build_team with team=${JSON.stringify(name)} and ${taskPart}`
  return `Use MCP tool invoke_agent with agent=${JSON.stringify(name)} and ${taskPart}`
}

function pickWorkflow(catalog: OrchestrationCatalog, names: string[], tokens: string[]): string | undefined {
  return pickName(names, tokens)
    ?? names.find((name) => catalog.workflows[name]?.phases.some((phase) => phase.ref && tokenMatch(phase.ref, tokens)))
}

function pickChain(catalog: OrchestrationCatalog, names: string[], tokens: string[]): string | undefined {
  return pickName(names, tokens)
    ?? names.find((name) => catalog.chains[name]?.steps.some((step) => tokenMatch(step.agent, ['architect', 'planner', 'implementor']) || tokenMatch(step.id, tokens)))
}

function pickTeam(catalog: OrchestrationCatalog, names: string[], tokens: string[]): string | undefined {
  return pickName(names, tokens)
    ?? names.find((name) => catalog.teams[name]?.parallel.some((member) => tokenMatch(member.agent, tokens) || tokenMatch(member.role, tokens) || tokenMatch(member.focus, tokens)))
}

function pickName(names: string[], tokens: string[]): string | undefined {
  for (const token of tokens) {
    const match = names.find((name) => tokenMatch(name, [token]))
    if (match) return match
  }
  return undefined
}

function tokenMatch(value: string, tokens: string[]): boolean {
  const normalized = value.toLowerCase()
  return tokens.some((token) => normalized.includes(token.toLowerCase()))
}

function sortedKeys(record: Record<string, unknown>): string[] {
  return Object.keys(record).sort((a, b) => a.localeCompare(b))
}

function first(values: string[]): string | undefined {
  return values[0]
}

function formatHumanReport(report: InitReport, catalog: OrchestrationCatalog, verbose: boolean): string {
  const lines: string[] = []
  lines.push('1. Project/context')
  lines.push(`   Project root: ${report.projectRoot}`)
  lines.push('')
  lines.push('2. Host CLI capability')
  lines.push(`   Host: ${report.host.host}`)
  lines.push(`   Dispatch mode: ${report.host.dispatchMode}`)
  lines.push(`   Supports subagents: ${yesNo(report.host.supportsSubagents)}`)
  lines.push(`   Supports parallel teams: ${yesNo(report.host.supportsParallelTeams)}`)
  lines.push(`   Supports structured output: ${yesNo(report.host.supportsStructuredOutput)}`)
  lines.push('')
  lines.push('3. Catalog inventory counts and names')
  lines.push(`   Agents (${report.inventory.agents.length}): ${formatNames(report.inventory.agents, catalog.agents, verbose)}`)
  lines.push(`   Domains (${report.inventory.domains.length}): ${formatNames(report.inventory.domains, catalog.domains, verbose)}`)
  lines.push(`   Modes (${report.inventory.modes.length}): ${formatNames(report.inventory.modes, catalog.modes, verbose)}`)
  lines.push(`   Chains (${report.inventory.chains.length}): ${formatNames(report.inventory.chains, catalog.chains, verbose)}`)
  lines.push(`   Teams (${report.inventory.teams.length}): ${formatNames(report.inventory.teams, catalog.teams, verbose)}`)
  lines.push(`   Workflows (${report.inventory.workflows.length}): ${formatNames(report.inventory.workflows, catalog.workflows, verbose)}`)
  lines.push('')
  lines.push('4. Root context files (`AGENTS.md`, `CLAUDE.md`)')
  lines.push(`   AGENTS.md: ${presentMissing(report.rootFiles.agentsMd)}`)
  lines.push(`   CLAUDE.md: ${presentMissing(report.rootFiles.claudeMd)}`)
  lines.push('')

  if (report.recommendation) {
    lines.push('5. Recommendation')
    lines.push(`   Kind: ${report.recommendation.kind}`)
    if (report.recommendation.name) lines.push(`   Name: ${report.recommendation.name}`)
    lines.push(`   Confidence: ${report.recommendation.confidence}`)
    lines.push(`   Reason: ${report.recommendation.reason}`)
    if (report.recommendation.nextCommand) lines.push(`   Next: ${report.recommendation.nextCommand}`)
    if (report.recommendation.alternatives.length > 0) {
      lines.push('   Alternatives:')
      for (const alternative of report.recommendation.alternatives) {
        lines.push(`   - ${alternative.kind}${alternative.name ? `/${alternative.name}` : ''}: ${alternative.reason}`)
      }
    }
  } else {
    lines.push('5. Examples')
    lines.push('   ai-setup-orchestrator init --task "review auth middleware" --host claude-code')
    lines.push('   ai-setup-orchestrator init --task "build checkout from scratch" --host opencode')
    lines.push('   ai-setup-orchestrator init --json')
  }

  return `${lines.join('\n')}\n`
}

function formatNames<T extends { source: string; path: string }>(names: string[], definitions: Record<string, T>, verbose: boolean): string {
  if (names.length === 0) return '(none)'
  if (!verbose) return names.join(', ')
  return names.map((name) => {
    const definition = definitions[name]
    return definition ? `${name} [${definition.source}: ${definition.path}]` : name
  }).join(', ')
}

function yesNo(value: boolean): string {
  return value ? 'yes' : 'no'
}

function presentMissing(value: boolean): string {
  return value ? 'present' : 'missing'
}
