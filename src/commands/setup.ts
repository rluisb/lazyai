import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import type { Command } from 'commander'

type SetupScope = 'global' | 'project' | 'workspace'
type SetupToolId = 'claude-code' | 'codex' | 'copilot' | 'gemini' | 'opencode' | 'pi'
type ResourceState = 'managed' | 'adoptable' | 'conflict' | 'user-owned' | 'missing'

interface RegistryRoot {
  scope: SetupScope
  origin: SetupScope
  rootPath: string
  countRootOnly?: boolean
  expectedFiles: string[]
  optionalPaths: string[]
  desiredExpectedFiles: string[]
}

interface RegistryTool {
  id: SetupToolId
  name: string
  supportedScopes: SetupScope[]
  roots: RegistryRoot[]
}

interface SetupListTarget {
  id: SetupToolId
  name: string
  supportedScopes: SetupScope[]
  candidateRoots: Array<{
    scope: SetupScope
    origin: SetupScope
    rootPath: string
    expectedFiles: string[]
  }>
}

interface SetupListResult {
  mode: 'list'
  scopeFilter?: SetupScope
  sharedPaths: Array<{ id: string; description: string; path: string }>
  targets: SetupListTarget[]
  agents?: []
}

interface SetupDryRunResult {
  mode: 'dry-run'
  scope: SetupScope
  sharedPaths: Array<{ id: string; description: string; path: string }>
  targets: SetupDryRunTarget[]
}

interface SetupDryRunTarget {
  id: SetupToolId
  name: string
  scope: SetupScope
  origin: SetupScope
  rootPath: string
  expectedFiles: string[]
  observedFiles?: string[]
  existingStatus: 'missing' | 'detected'
  existingState?: ResourceState
  action: 'initialize' | 'preserve-existing'
}

interface SetupOptions {
  list?: boolean
  dryRun?: boolean
  tool?: string[]
  all?: boolean
  global?: boolean
}

interface SetupSelection {
  explicit: boolean
  requested: Set<SetupToolId>
}

const SETUP_TOOLS: RegistryTool[] = [
  {
    id: 'claude-code',
    name: 'Claude Code',
    supportedScopes: ['global', 'project', 'workspace'],
    roots: [
      {
        scope: 'global',
        origin: 'global',
        rootPath: '<HOME_DIR>/.claude',
        expectedFiles: ['settings.json'],
        optionalPaths: ['settings.local.json', 'agents', 'skills', 'commands', 'output-styles'],
        desiredExpectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
      },
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.claude',
        expectedFiles: ['settings.json'],
        optionalPaths: ['settings.local.json', 'agents', 'skills', 'commands', 'output-styles'],
        desiredExpectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.claude',
        expectedFiles: ['settings.json'],
        optionalPaths: ['settings.local.json', 'agents', 'skills', 'commands', 'output-styles'],
        desiredExpectedFiles: ['agents', 'commands', 'output-styles', 'settings.json', 'settings.local.json', 'skills'],
      },
    ],
  },
  {
    id: 'codex',
    name: 'Codex CLI',
    supportedScopes: ['global', 'project', 'workspace'],
    roots: [
      {
        scope: 'global',
        origin: 'global',
        rootPath: '<HOME_DIR>/.codex',
        expectedFiles: ['config.toml'],
        optionalPaths: [],
        desiredExpectedFiles: ['config.toml'],
      },
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.codex',
        expectedFiles: ['config.toml'],
        optionalPaths: [],
        desiredExpectedFiles: ['config.toml'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.codex',
        expectedFiles: ['config.toml'],
        optionalPaths: [],
        desiredExpectedFiles: ['config.toml'],
      },
    ],
  },
  {
    id: 'copilot',
    name: 'GitHub Copilot CLI',
    supportedScopes: ['global', 'project', 'workspace'],
    roots: [
      {
        scope: 'global',
        origin: 'global',
        rootPath: '<HOME_DIR>/.copilot',
        countRootOnly: true,
        expectedFiles: ['mcp-config.json'],
        optionalPaths: [],
        desiredExpectedFiles: ['mcp-config.json'],
      },
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.github',
        expectedFiles: ['copilot-instructions.md'],
        optionalPaths: ['agents', 'instructions', 'prompts', 'chatmodes'],
        desiredExpectedFiles: ['agents', 'chatmodes', 'copilot-instructions.md', 'instructions', 'prompts'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.github',
        expectedFiles: ['copilot-instructions.md'],
        optionalPaths: ['agents', 'instructions', 'prompts', 'chatmodes'],
        desiredExpectedFiles: ['agents', 'chatmodes', 'copilot-instructions.md', 'instructions', 'prompts'],
      },
    ],
  },
  {
    id: 'gemini',
    name: 'Gemini CLI',
    supportedScopes: ['global', 'project', 'workspace'],
    roots: [
      {
        scope: 'global',
        origin: 'global',
        rootPath: '<HOME_DIR>/.gemini',
        expectedFiles: ['settings.json'],
        optionalPaths: ['commands'],
        desiredExpectedFiles: ['commands', 'settings.json'],
      },
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.gemini',
        expectedFiles: ['settings.json'],
        optionalPaths: ['commands'],
        desiredExpectedFiles: ['commands', 'settings.json'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.gemini',
        expectedFiles: ['settings.json'],
        optionalPaths: ['commands'],
        desiredExpectedFiles: ['commands', 'settings.json'],
      },
    ],
  },
  {
    id: 'opencode',
    name: 'OpenCode',
    supportedScopes: ['global', 'project', 'workspace'],
    roots: [
      {
        scope: 'global',
        origin: 'global',
        rootPath: '<HOME_DIR>/.config/opencode',
        expectedFiles: ['opencode.jsonc'],
        optionalPaths: ['opencode.json', 'agents', 'skills', 'commands', 'modes', 'AGENTS.md'],
        desiredExpectedFiles: ['AGENTS.md', 'agents', 'commands', 'modes', 'opencode.json', 'opencode.jsonc', 'skills'],
      },
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.opencode',
        expectedFiles: ['opencode.jsonc'],
        optionalPaths: ['opencode.json', 'agents', 'skills', 'commands', 'modes', 'AGENTS.md'],
        desiredExpectedFiles: ['AGENTS.md', 'agents', 'commands', 'modes', 'opencode.json', 'opencode.jsonc', 'skills'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.opencode',
        expectedFiles: ['opencode.jsonc'],
        optionalPaths: ['opencode.json', 'agents', 'skills', 'commands', 'modes', 'AGENTS.md'],
        desiredExpectedFiles: ['AGENTS.md', 'agents', 'commands', 'modes', 'opencode.json', 'opencode.jsonc', 'skills'],
      },
    ],
  },
  {
    id: 'pi',
    name: 'Pi',
    supportedScopes: ['project', 'workspace'],
    roots: [
      {
        scope: 'project',
        origin: 'project',
        rootPath: '<TARGET_DIR>/.pi',
        expectedFiles: ['settings.json'],
        optionalPaths: ['skills', 'prompts'],
        desiredExpectedFiles: ['prompts', 'settings.json', 'skills'],
      },
      {
        scope: 'workspace',
        origin: 'workspace',
        rootPath: '<TARGET_DIR>/.pi',
        expectedFiles: ['settings.json'],
        optionalPaths: ['skills', 'prompts'],
        desiredExpectedFiles: ['prompts', 'settings.json', 'skills'],
      },
    ],
  },
]

function collectTool(value: string, previous: string[]): string[] {
  previous.push(value)
  return previous
}

function resolveSetupSelection(names: string[] | undefined): SetupSelection {
  const requested = new Set<SetupToolId>()
  const values = names ?? []

  for (const name of values) {
    const trimmed = name.trim()
    if (!trimmed) {
      continue
    }

    const tool = SETUP_TOOLS.find(({ id }) => id === trimmed)
    if (!tool) {
      throw new Error(`unknown tool ${JSON.stringify(trimmed)}`)
    }

    requested.add(tool.id)
  }

  return {
    explicit: requested.size > 0,
    requested,
  }
}

function resolveTokenizedPath(templatePath: string, targetDir: string, homeDir: string): string {
  return templatePath.replace('<TARGET_DIR>', targetDir).replace('<HOME_DIR>', homeDir)
}

function buildSharedPaths(targetDir: string, homeDir: string, scope?: SetupScope): Array<{ id: string; description: string; path: string }> {
  const sharedPaths = [
    { id: 'global-ai-setup', description: 'Global ai-setup managed directory', path: path.join(homeDir, '.ai-setup') },
    { id: 'project-ai', description: 'Project-local ai-setup directory', path: path.join(targetDir, '.ai') },
  ]

  if (!scope) {
    return sharedPaths
  }

  const prefix = scope === 'global' ? 'global-' : 'project-'
  return sharedPaths.filter(({ id }) => id.startsWith(prefix))
}

function filterTargets(selection: SetupSelection, scope?: SetupScope): SetupListTarget[] {
  const filtered = SETUP_TOOLS.flatMap((tool) => {
    if (selection.explicit && !selection.requested.has(tool.id)) {
      return []
    }

    if (scope && !tool.supportedScopes.includes(scope)) {
      return []
    }

    const candidateRoots = tool.roots
      .filter((root) => (scope ? root.scope === scope : true))
      .map((root) => ({
        scope: root.scope,
        origin: root.origin,
        rootPath: root.rootPath,
        expectedFiles: [...root.desiredExpectedFiles],
      }))

    return [{
      id: tool.id,
      name: tool.name,
      supportedScopes: scope ? [scope] : [...tool.supportedScopes],
      candidateRoots,
    }]
  })

  if (scope && selection.explicit) {
    for (const targetId of selection.requested) {
      if (!filtered.some((target) => target.id === targetId)) {
        throw new Error(`tool ${JSON.stringify(targetId)} does not support scope ${JSON.stringify(scope)}`)
      }
    }
  }

  return filtered
}

function collectObservedFiles(rootPath: string, candidates: string[]): string[] {
  const observed = new Set<string>()

  for (const relativePath of candidates) {
    if (!relativePath) {
      continue
    }

    const absolutePath = path.join(rootPath, relativePath)
    if (fs.existsSync(absolutePath)) {
      observed.add(relativePath.replaceAll(path.sep, '/'))
    }
  }

  return [...observed].sort()
}

function buildListResult(selection: SetupSelection, scope: SetupScope | undefined, targetDir: string, homeDir: string): SetupListResult {
  const targets = filterTargets(selection, scope).map((target) => ({
    ...target,
    candidateRoots: target.candidateRoots.map((root) => ({
      ...root,
      rootPath: resolveTokenizedPath(root.rootPath, targetDir, homeDir),
    })),
  }))

  return {
    mode: 'list',
    ...(scope ? { scopeFilter: scope } : {}),
    sharedPaths: buildSharedPaths(targetDir, homeDir, scope),
    targets,
  }
}

function buildDryRunResult(selection: SetupSelection, scope: SetupScope, targetDir: string, homeDir: string): SetupDryRunResult {
  const targets = filterTargets(selection, scope).flatMap((target) => {
    const root = target.candidateRoots[0]
    if (!root) {
      return []
    }

    const registryTool = SETUP_TOOLS.find(({ id }) => id === target.id)
    const registryRoot = registryTool?.roots.find((candidate) => candidate.scope === scope)
    if (!registryRoot) {
      return []
    }

    const rootPath = resolveTokenizedPath(root.rootPath, targetDir, homeDir)
    const observedFiles = collectObservedFiles(rootPath, [...registryRoot.expectedFiles, ...registryRoot.optionalPaths])
    const rootExists = fs.existsSync(rootPath)
    const detected = observedFiles.length > 0 || (registryRoot.countRootOnly === true && rootExists)

    const dryRunTarget: SetupDryRunTarget = {
      id: target.id,
      name: target.name,
      scope,
      origin: root.origin,
      rootPath,
      expectedFiles: [...root.expectedFiles],
      existingStatus: detected ? 'detected' : 'missing',
      action: detected ? 'preserve-existing' : 'initialize',
    }

    if (observedFiles.length > 0) {
      dryRunTarget.observedFiles = observedFiles
    }
    if (detected) {
      dryRunTarget.existingState = 'adoptable'
    }

    return [dryRunTarget]
  })

  return {
    mode: 'dry-run',
    scope,
    sharedPaths: buildSharedPaths(targetDir, homeDir, scope),
    targets,
  }
}

export function registerSetup(program: Command): void {
  program
    .command('setup')
    .description('Inspect setup inventory and planning output')
    .option('--list', 'List supported setup targets and reusable setup resources')
    .option('--dry-run', 'Show the setup plan without writing files')
    .option('--tool <tool>', 'Limit setup planning to specific tools (repeatable)', collectTool, [])
    .option('--all', 'Select all supported setup targets for the requested scope')
    .option('--global', 'Use global scope/home layout where supported')
    .action(async (opts: SetupOptions) => {
      const actions = [opts.list === true, opts.dryRun === true].filter(Boolean).length
      if (actions === 0) {
        throw new Error('no setup action selected (try: ai-setup setup --list or --dry-run)')
      }

      if (actions > 1) {
        throw new Error('select exactly one of --list or --dry-run')
      }

      if (opts.all === true && (opts.tool?.length ?? 0) > 0) {
        throw new Error('--all cannot be combined with --tool')
      }

      const selection = resolveSetupSelection(opts.tool)
      const scope: SetupScope | undefined = opts.global === true ? 'global' : undefined
      const targetDir = process.cwd()
      const homeDir = os.homedir()

      if (opts.list === true) {
        console.log(JSON.stringify(buildListResult(selection, scope, targetDir, homeDir), null, 2))
        return
      }

      console.log(JSON.stringify(buildDryRunResult(selection, scope ?? 'project', targetDir, homeDir), null, 2))
    })
}
