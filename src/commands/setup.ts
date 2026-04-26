import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import type { Command } from 'commander'
import { splitYamlFrontmatter } from '../utils/frontmatter.js'
import { stripJsonComments } from '../utils/jsonc.js'

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

interface SetupInventoryResult {
  currentState: SetupCurrentState
  desiredState: SetupDesiredState
}

interface SetupCurrentState {
  sharedPaths: ObservedPath[]
  targets: ObservedTarget[]
  agents?: ObservedAgent[]
}

interface SetupDesiredState {
  sharedPaths: Array<{ id: string; description: string; path: string }>
  targets: DesiredTarget[]
}

interface ObservedPath {
  id: string
  path: string
  exists: boolean
}

interface ObservedTarget {
  id: SetupToolId
  name: string
  detections: TargetDetection[]
}

interface TargetDetection {
  scope: SetupScope
  origin: SetupScope
  rootPath: string
  status: 'missing' | 'detected'
  state: ResourceState
  version: string
  observedFiles: string[]
}

interface DesiredTarget {
  id: SetupToolId
  name: string
  supportedScopes: SetupScope[]
  candidateRoots: DesiredRoot[]
}

interface DesiredRoot {
  scope: SetupScope
  origin: SetupScope
  rootPath: string
  expectedFiles: string[]
}

interface ObservedAgent {
  id: string
  directory: string
  promptPath?: string
  status: 'detected' | 'invalid'
  title?: string
  description?: string
  tools?: string[]
  mcp?: ObservedAgentMcp
  reasons?: string[]
}

interface ObservedAgentMcp {
  configPath: string
  scoped: boolean
  serverNames?: string[]
  serverCount: number
}

interface SetupOptions {
  scan?: boolean
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

const RESOURCE_STATE_ADOPTABLE: ResourceState = 'adoptable'
const RESOURCE_STATE_MISSING: ResourceState = 'missing'
const REUSABLE_AGENT_ID_PATTERN = /^[a-z][a-z0-9-]{0,63}$/
const TOML_VERSION_PATTERN = /^version\s*=\s*"([^"]+)"/m

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
    if (pathExists(absolutePath)) {
      observed.add(relativePath.replaceAll(path.sep, '/'))
    }
  }

  return [...observed].sort()
}

function pathExists(targetPath: string): boolean {
  return fs.existsSync(targetPath)
}

function directoryExists(targetPath: string): boolean {
  try {
    return fs.statSync(targetPath).isDirectory()
  } catch {
    return false
  }
}

function fileExists(targetPath: string): boolean {
  try {
    return fs.statSync(targetPath).isFile()
  } catch {
    return false
  }
}

function buildScanInventory(targetDir: string, homeDir: string): SetupInventoryResult {
  const desiredSharedPaths = buildSharedPaths(targetDir, homeDir)
  const currentSharedPaths: ObservedPath[] = desiredSharedPaths.map(({ id, path }) => ({
    id,
    path,
    exists: directoryExists(path),
  }))

  const currentTargets: ObservedTarget[] = SETUP_TOOLS.map((tool) => ({
    id: tool.id,
    name: tool.name,
    detections: tool.roots
      .map((root) => buildTargetDetection(root, targetDir, homeDir))
      .sort((left, right) => {
        if (left.origin !== right.origin) {
          return left.origin.localeCompare(right.origin)
        }
        if (left.scope !== right.scope) {
          return left.scope.localeCompare(right.scope)
        }
        return left.rootPath.localeCompare(right.rootPath)
      }),
  }))

  const desiredTargets: DesiredTarget[] = SETUP_TOOLS.map((tool) => ({
    id: tool.id,
    name: tool.name,
    supportedScopes: [...tool.supportedScopes],
    candidateRoots: tool.roots.map((root) => ({
      scope: root.scope,
      origin: root.origin,
      rootPath: resolveTokenizedPath(root.rootPath, targetDir, homeDir),
      expectedFiles: [...root.desiredExpectedFiles].sort(),
    })),
  }))

  const agents = observeAgents(targetDir)

  return {
    currentState: {
      sharedPaths: currentSharedPaths,
      targets: currentTargets,
      ...(agents.length > 0 ? { agents } : {}),
    },
    desiredState: {
      sharedPaths: desiredSharedPaths,
      targets: desiredTargets,
    },
  }
}

function buildTargetDetection(root: RegistryRoot, targetDir: string, homeDir: string): TargetDetection {
  const rootPath = resolveTokenizedPath(root.rootPath, targetDir, homeDir)
  const observedFiles = collectObservedFiles(rootPath, [...root.expectedFiles, ...root.optionalPaths])
  const detected = observedFiles.length > 0 || (root.countRootOnly === true && directoryExists(rootPath))

  return {
    scope: root.scope,
    origin: root.origin,
    rootPath,
    status: detected ? 'detected' : 'missing',
    state: detected ? RESOURCE_STATE_ADOPTABLE : RESOURCE_STATE_MISSING,
    version: detectVersion(rootPath, root.expectedFiles),
    observedFiles,
  }
}

function detectVersion(rootPath: string, candidates: string[]): string {
  for (const relativePath of candidates) {
    const version = detectVersionFromFile(path.join(rootPath, relativePath))
    if (version !== 'unknown') {
      return version
    }
  }

  return 'unknown'
}

function detectVersionFromFile(filePath: string): string {
  if (!fileExists(filePath)) {
    return 'unknown'
  }

  try {
    const raw = fs.readFileSync(filePath, 'utf-8')

    if (filePath.endsWith('.jsonc')) {
      return versionFromRecord(JSON.parse(stripJsonComments(raw)))
    }

    if (filePath.endsWith('.json')) {
      return versionFromRecord(JSON.parse(raw))
    }

    if (filePath.endsWith('.toml')) {
      return TOML_VERSION_PATTERN.exec(raw)?.[1] ?? 'unknown'
    }
  } catch {
    return 'unknown'
  }

  return 'unknown'
}

function versionFromRecord(value: unknown): string {
  if (!value || typeof value !== 'object') {
    return 'unknown'
  }

  const version = (value as Record<string, unknown>).version
  if (typeof version === 'string' && version.trim() !== '') {
    return version
  }
  if (typeof version === 'number') {
    return String(version)
  }

  return 'unknown'
}

function observeAgents(targetDir: string): ObservedAgent[] {
  const agentsRoot = path.join(targetDir, '.ai', 'agents')
  if (!directoryExists(agentsRoot)) {
    return []
  }

  try {
    return fs.readdirSync(agentsRoot, { withFileTypes: true })
      .filter((entry) => entry.isDirectory())
      .map((entry) => observeAgentDirectory(path.join(agentsRoot, entry.name), entry.name))
      .sort((left, right) => left.id.localeCompare(right.id))
  } catch {
    return [{ id: '.ai/agents', directory: agentsRoot, status: 'invalid', reasons: ['invalid-agents-root'] }]
  }
}

function observeAgentDirectory(directory: string, id: string): ObservedAgent {
  const agent: ObservedAgent = {
    id,
    directory,
    promptPath: path.join(directory, 'AGENT.md'),
    status: 'detected',
  }

  if (!REUSABLE_AGENT_ID_PATTERN.test(id)) {
    agent.status = 'invalid'
    agent.reasons = ['invalid-agent-id']
  }

  const hydrated = hydrateAgent(agent)
  if (hydrated !== null) {
    agent.status = 'invalid'
    agent.reasons = [...(agent.reasons ?? []), hydrated]
  }

  if (agent.tools) {
    agent.tools = [...agent.tools].sort()
  }
  if (agent.reasons) {
    agent.reasons = [...agent.reasons].sort()
  }

  return agent
}

function hydrateAgent(agent: ObservedAgent): string | null {
  if (!agent.promptPath || !fileExists(agent.promptPath)) {
    return 'missing-agent-md'
  }

  const prompt = fs.readFileSync(agent.promptPath, 'utf-8')
  const parsed = parseAgentPrompt(prompt)
  if ('error' in parsed) {
    return parsed.error
  }

  const body = parsed.body.trim()
  if (body === '') {
    return 'empty-agent-body'
  }

  const title = firstNonEmpty(parsed.frontmatter.title, parsed.frontmatter.name, extractFirstHeading(body))
  if (title) {
    agent.title = title
  }

  const description = firstNonEmpty(parsed.frontmatter.description, extractFirstParagraph(body))
  if (description) {
    agent.description = description
  }
  if (parsed.frontmatter.tools.length > 0) {
    agent.tools = parsed.frontmatter.tools
  }

  const mcpPath = path.join(agent.directory, 'mcp.json')
  if (!fileExists(mcpPath)) {
    return null
  }

  const mcp = parseAgentMcp(mcpPath)
  if ('error' in mcp) {
    return mcp.error
  }

  agent.mcp = mcp.value
  return null
}

function parseAgentPrompt(content: string):
  | { frontmatter: { title: string; name: string; description: string; tools: string[] }; body: string }
  | { error: string } {
  if (content.startsWith('---')) {
    const split = splitYamlFrontmatter(content)
    if (!split) {
      return { error: 'invalid-agent-frontmatter' }
    }

    const frontmatter = parseAgentFrontmatterFields(split.frontmatterBody)
    if ('error' in frontmatter) {
      return frontmatter
    }

    return {
      frontmatter,
      body: split.body,
    }
  }

  return {
    frontmatter: { title: '', name: '', description: '', tools: [] },
    body: content,
  }
}

function parseAgentFrontmatterFields(frontmatterBody: string):
  | { title: string; name: string; description: string; tools: string[] }
  | { error: string } {
  const fields = {
    title: '',
    name: '',
    description: '',
    tools: [] as string[],
  }

  const lines = frontmatterBody.split('\n')

  for (let index = 0; index < lines.length; index++) {
    const rawLine = lines[index] ?? ''
    const trimmed = rawLine.trim()
    if (trimmed === '' || trimmed.startsWith('#')) {
      continue
    }

    const match = rawLine.match(/^([A-Za-z0-9_-]+)\s*:\s*(.*)$/)
    if (!match) {
      return { error: 'invalid-agent-frontmatter' }
    }

    const [, key, rawValue = ''] = match
    const value = normalizeYamlScalar(rawValue)

    if (key === 'title') {
      fields.title = value
      continue
    }
    if (key === 'name') {
      fields.name = value
      continue
    }
    if (key === 'description') {
      fields.description = value
      continue
    }
    if (key === 'tools') {
      if (value !== '') {
        fields.tools = normalizeTools([value])
        continue
      }

      const listValues: string[] = []
      let nextIndex = index + 1
      while (nextIndex < lines.length) {
        const listLine = lines[nextIndex] ?? ''
        if (listLine.trim() === '') {
          nextIndex++
          continue
        }

        const listMatch = listLine.match(/^\s*-\s*(.*)$/)
        if (!listMatch) {
          break
        }

        listValues.push(normalizeYamlScalar(listMatch[1] ?? ''))
        nextIndex++
      }

      fields.tools = normalizeTools(listValues)
      index = nextIndex - 1
    }
  }

  return fields
}

function normalizeYamlScalar(value: string): string {
  const trimmed = value.trim()
  if (
    (trimmed.startsWith('"') && trimmed.endsWith('"'))
    || (trimmed.startsWith("'") && trimmed.endsWith("'"))
  ) {
    return trimmed.slice(1, -1).trim()
  }

  return trimmed
}

function normalizeTools(values: string[]): string[] {
  return values
    .map((value) => value.trim())
    .filter((value) => value.length > 0)
    .sort()
}

function parseAgentMcp(configPath: string): { value: ObservedAgentMcp } | { error: string } {
  try {
    const value = JSON.parse(fs.readFileSync(configPath, 'utf-8')) as Record<string, unknown>

    if (!value || Array.isArray(value) || Object.keys(value).length !== 1 || !('mcpServers' in value)) {
      return { error: 'invalid-agent-mcp-schema' }
    }

    const servers = value.mcpServers
    if (!servers || typeof servers !== 'object' || Array.isArray(servers)) {
      return { error: 'invalid-agent-mcp-schema' }
    }

    const serverNames = Object.keys(servers)
      .map((name) => name.trim())
      .filter((name) => name.length > 0)
      .sort()

    if (serverNames.length !== Object.keys(servers).length) {
      return { error: 'invalid-agent-mcp-server-name' }
    }

    return {
      value: {
        configPath,
        scoped: true,
        serverNames,
        serverCount: serverNames.length,
      },
    }
  } catch {
    return { error: 'invalid-agent-mcp-schema' }
  }
}

function firstNonEmpty(...values: Array<string | undefined>): string | undefined {
  for (const value of values) {
    if (value && value.trim() !== '') {
      return value.trim()
    }
  }

  return undefined
}

function extractFirstHeading(content: string): string | undefined {
  for (const line of content.split('\n')) {
    const trimmed = line.trim()
    if (trimmed.startsWith('# ')) {
      return trimmed.slice(2).trim()
    }
  }

  return undefined
}

function extractFirstParagraph(content: string): string | undefined {
  const paragraphs = content.split(/\n\s*\n/)
  for (const paragraph of paragraphs) {
    const trimmed = paragraph.trim()
    if (trimmed === '' || trimmed.startsWith('#')) {
      continue
    }

    return trimmed.split(/\s+/).join(' ')
  }

  return undefined
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
  const inventory = buildScanInventory(targetDir, homeDir)
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
    const detection = inventory.currentState.targets
      .find(({ id }) => id === target.id)
      ?.detections.find((candidate) => candidate.scope === scope)

    const dryRunTarget: SetupDryRunTarget = {
      id: target.id,
      name: target.name,
      scope,
      origin: root.origin,
      rootPath,
      expectedFiles: [...root.expectedFiles],
      existingStatus: 'missing',
      action: 'initialize',
    }

    if (detection) {
      if (detection.observedFiles.length > 0) {
        dryRunTarget.observedFiles = [...detection.observedFiles]
      }
      dryRunTarget.existingStatus = detection.status
      dryRunTarget.existingState = detection.state
      if (detection.status === 'detected') {
        dryRunTarget.action = 'preserve-existing'
      }
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
    .option('--scan', 'Scan known tool targets and print inventory JSON')
    .option('--list', 'List supported setup targets and reusable setup resources')
    .option('--dry-run', 'Show the setup plan without writing files')
    .option('--tool <tool>', 'Limit setup planning to specific tools (repeatable)', collectTool, [])
    .option('--all', 'Select all supported setup targets for the requested scope')
    .option('--global', 'Use global scope/home layout where supported')
    .action(async (opts: SetupOptions) => {
      const actions = [opts.scan === true, opts.list === true, opts.dryRun === true].filter(Boolean).length
      if (actions === 0) {
        throw new Error('no setup action selected (try: ai-setup setup --scan, --list, or --dry-run)')
      }

      if (actions > 1) {
        throw new Error('select exactly one of --scan, --list, or --dry-run')
      }

      if (opts.scan === true && ((opts.tool?.length ?? 0) > 0 || opts.all === true || opts.global === true)) {
        throw new Error('--tool, --all, and --global are only supported with --list or --dry-run')
      }

      if (opts.all === true && (opts.tool?.length ?? 0) > 0) {
        throw new Error('--all cannot be combined with --tool')
      }

      const selection = resolveSetupSelection(opts.tool)
      const scope: SetupScope | undefined = opts.global === true ? 'global' : undefined
      const targetDir = fs.realpathSync.native(process.cwd())
      const homeDir = os.homedir()

      if (opts.scan === true) {
        console.log(JSON.stringify(buildScanInventory(targetDir, homeDir), null, 2))
        return
      }

      if (opts.list === true) {
        console.log(JSON.stringify(buildListResult(selection, scope, targetDir, homeDir), null, 2))
        return
      }

      console.log(JSON.stringify(buildDryRunResult(selection, scope ?? 'project', targetDir, homeDir), null, 2))
    })
}
