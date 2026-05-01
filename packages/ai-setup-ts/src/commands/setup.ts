import { createHash } from 'node:crypto'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import type { Command } from 'commander'
import { splitYamlFrontmatter } from '../utils/frontmatter.js'
import { stripJsonComments } from '../utils/jsonc.js'

type SetupScope = 'global' | 'project' | 'workspace'
type SetupToolId = 'claude-code' | 'copilot' | 'opencode'
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
  agents?: ObservedAgent[]
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
  reasons?: string[]
  mcpEntries?: MCPEntry[]
}

interface MCPEntry {
  name: string
  configPath: string
  state: ResourceState
  reasons?: string[]
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
  adopt?: boolean
  import?: boolean
  tool?: string[]
  all?: boolean
  global?: boolean
}

interface SetupSelection {
  explicit: boolean
  requested: Set<SetupToolId>
}

interface SetupOperationResult {
  mode: 'adopt' | 'import' | 'adopt-import'
  registryPath: string
  importRoot: string
  backups?: string[]
  adopted?: OperationTarget[]
  imported?: ImportedResource[]
  skipped?: SkippedOperation[]
}

interface OperationTarget {
  targetId: string
  scope: string
  origin: string
  rootPath: string
}

interface ImportedResource {
  targetId: string
  scope: string
  origin: string
  sourcePath: string
  destinationPath: string
}

interface SkippedOperation {
  targetId: string
  scope: string
  origin: string
  rootPath: string
  state: string
  reason: string
}

interface ScanRegistry {
  version: number
  resources?: ManagedResource[]
  imports?: ImportRecord[]
}

interface ManagedResource {
  targetId: string
  scope: string
  origin: string
  rootPath: string
  state: 'managed' | 'user-owned'
  observedPaths?: RecordedPath[]
  mcpEntries?: RecordedMcpEntry[]
  updatedAt: string
}

interface ImportRecord {
  targetId: string
  scope: string
  origin: string
  rootPath: string
  importedPaths?: string[]
  destinationRoot: string
  updatedAt: string
}

interface RecordedPath {
  relativePath: string
  fingerprint: string
}

interface RecordedMcpEntry {
  name: string
  configPath: string
  fingerprint: string
}

interface ObservedMcpEntry {
  entry: MCPEntry
  fingerprint: string
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
]

const RESOURCE_STATE_ADOPTABLE: ResourceState = 'adoptable'
const RESOURCE_STATE_CONFLICT: ResourceState = 'conflict'
const RESOURCE_STATE_MANAGED: ResourceState = 'managed'
const RESOURCE_STATE_MISSING: ResourceState = 'missing'
const RESOURCE_STATE_USER_OWNED: ResourceState = 'user-owned'
const REUSABLE_AGENT_ID_PATTERN = /^[a-z][a-z0-9-]{0,63}$/
const TOML_VERSION_PATTERN = /^version\s*=\s*"([^"]+)"/m
const SCAN_REGISTRY_VERSION = 1

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
  const registry = loadRegistry(aiSetupHome(homeDir))
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
      .map((root) => buildTargetDetection(root, targetDir, homeDir, tool.id, registry))
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

function buildTargetDetection(root: RegistryRoot, targetDir: string, homeDir: string, toolId: SetupToolId, registry: ScanRegistry): TargetDetection {
  const rootPath = resolveTokenizedPath(root.rootPath, targetDir, homeDir)
  const observedFiles = collectObservedFiles(rootPath, [...root.expectedFiles, ...root.optionalPaths])
  const detected = observedFiles.length > 0 || (root.countRootOnly === true && directoryExists(rootPath))

  const detection: TargetDetection = {
    scope: root.scope,
    origin: root.origin,
    rootPath,
    status: detected ? 'detected' : 'missing',
    state: detected ? RESOURCE_STATE_ADOPTABLE : RESOURCE_STATE_MISSING,
    version: detectVersion(rootPath, root.expectedFiles),
    observedFiles,
  }

  applyDetectionState(detection, toolId, registry)
  return detection
}

function applyDetectionState(detection: TargetDetection, toolId: SetupToolId, registry: ScanRegistry): void {
  if (detection.status === 'missing') {
    detection.state = RESOURCE_STATE_MISSING
    return
  }

  try {
    const observedPaths = snapshotObservedPaths(detection.rootPath, detection.observedFiles)
    const observedMcpEntries = snapshotMcpEntries(detection.rootPath, detection.observedFiles)
    if (observedMcpEntries.length > 0) {
      detection.mcpEntries = observedMcpEntries.map(({ entry }) => ({ ...entry }))
    }

    const record = findResourceRecord(registry, toolId, detection.scope, detection.origin, detection.rootPath)
    if (!record) {
      detection.state = RESOURCE_STATE_ADOPTABLE
      detection.mcpEntries?.forEach((entry) => {
        entry.state = RESOURCE_STATE_ADOPTABLE
      })
      return
    }

    if (record.state === RESOURCE_STATE_USER_OWNED) {
      detection.state = RESOURCE_STATE_USER_OWNED
      detection.mcpEntries?.forEach((entry) => {
        entry.state = RESOURCE_STATE_USER_OWNED
      })
      return
    }

    if (record.state !== RESOURCE_STATE_MANAGED) {
      detection.state = RESOURCE_STATE_CONFLICT
      detection.reasons = ['unsupported-registry-state']
      detection.mcpEntries?.forEach((entry) => {
        entry.state = RESOURCE_STATE_CONFLICT
      })
      return
    }

    const reasons = compareObservedPaths(record.observedPaths ?? [], observedPaths)
    const mcpReasons = applyMcpEntryStates(detection.mcpEntries, record.mcpEntries ?? [], observedMcpEntries)
    const allReasons = [...reasons, ...mcpReasons]

    if (allReasons.length === 0) {
      detection.state = RESOURCE_STATE_MANAGED
      return
    }

    detection.state = RESOURCE_STATE_CONFLICT
    detection.reasons = allReasons
  } catch (error) {
    detection.state = RESOURCE_STATE_CONFLICT
    detection.reasons = [`snapshot-failed:${error instanceof Error ? error.message : String(error)}`]
  }
}

function compareObservedPaths(recorded: RecordedPath[], observed: RecordedPath[]): string[] {
  const recordedByPath = new Map(recorded.map((entry) => [entry.relativePath, entry.fingerprint]))
  const observedByPath = new Map(observed.map((entry) => [entry.relativePath, entry.fingerprint]))
  const reasons: string[] = []

  for (const [relativePath, fingerprint] of recordedByPath) {
    const observedFingerprint = observedByPath.get(relativePath)
    if (!observedFingerprint) {
      reasons.push(`missing-path:${relativePath}`)
      continue
    }
    if (observedFingerprint !== fingerprint) {
      reasons.push(`changed-path:${relativePath}`)
    }
  }

  for (const relativePath of observedByPath.keys()) {
    if (!recordedByPath.has(relativePath)) {
      reasons.push(`unexpected-path:${relativePath}`)
    }
  }

  return reasons.sort()
}

function applyMcpEntryStates(entries: MCPEntry[] | undefined, recorded: RecordedMcpEntry[], observed: ObservedMcpEntry[]): string[] {
  if (!entries || entries.length === 0) {
    return []
  }

  const recordedByKey = new Map(recorded.map((entry) => [mcpEntryKey(entry.configPath, entry.name), entry]))
  const observedByKey = new Map(observed.map((entry) => [mcpEntryKey(entry.entry.configPath, entry.entry.name), entry]))
  const reasons: string[] = []

  for (const entry of entries) {
    const key = mcpEntryKey(entry.configPath, entry.name)
    const recordedEntry = recordedByKey.get(key)
    if (!recordedEntry) {
      entry.state = RESOURCE_STATE_CONFLICT
      entry.reasons = ['unexpected-mcp-entry']
      reasons.push(`unexpected-mcp-entry:${entry.configPath}:${entry.name}`)
      continue
    }

    const observedEntry = observedByKey.get(key)
    if (!observedEntry || observedEntry.fingerprint !== recordedEntry.fingerprint) {
      entry.state = RESOURCE_STATE_CONFLICT
      entry.reasons = ['mcp-entry-changed']
      reasons.push(`mcp-entry-changed:${entry.configPath}:${entry.name}`)
      continue
    }

    entry.state = RESOURCE_STATE_MANAGED
  }

  for (const recordedEntry of recorded) {
    if (!observedByKey.has(mcpEntryKey(recordedEntry.configPath, recordedEntry.name))) {
      reasons.push(`missing-mcp-entry:${recordedEntry.configPath}:${recordedEntry.name}`)
    }
  }

  return reasons.sort()
}

function mcpEntryKey(configPath: string, name: string): string {
  return `${configPath}::${name}`
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

function scanRegistryPath(aiSetupHomePath: string): string {
  return path.join(aiSetupHomePath, 'config', 'setup-scan-registry.json')
}

function aiSetupHome(homeDir: string): string {
  return path.join(homeDir, '.ai-setup')
}

function importsRoot(aiSetupHomePath: string): string {
  return path.join(aiSetupHomePath, 'imports')
}

function loadRegistry(aiSetupHomePath: string): ScanRegistry {
  const registryPath = scanRegistryPath(aiSetupHomePath)
  if (!fileExists(registryPath)) {
    return { version: SCAN_REGISTRY_VERSION }
  }

  const parsed = JSON.parse(fs.readFileSync(registryPath, 'utf-8')) as ScanRegistry
  return {
    version: parsed.version || SCAN_REGISTRY_VERSION,
    resources: [...(parsed.resources ?? [])],
    imports: [...(parsed.imports ?? [])],
  }
}

function findResourceRecord(registry: ScanRegistry, toolId: string, scope: string, origin: string, rootPath: string): ManagedResource | undefined {
  return registry.resources?.find((entry) => (
    entry.targetId === toolId
    && entry.scope === scope
    && entry.origin === origin
    && entry.rootPath === rootPath
  ))
}

function upsertResourceRecord(registry: ScanRegistry, record: ManagedResource): void {
  registry.resources ??= []
  const index = registry.resources.findIndex((entry) => (
    entry.targetId === record.targetId
    && entry.scope === record.scope
    && entry.origin === record.origin
    && entry.rootPath === record.rootPath
  ))
  if (index >= 0) {
    registry.resources[index] = record
    return
  }
  registry.resources.push(record)
}

function upsertImportRecord(registry: ScanRegistry, record: ImportRecord): void {
  registry.imports ??= []
  const index = registry.imports.findIndex((entry) => (
    entry.targetId === record.targetId
    && entry.scope === record.scope
    && entry.origin === record.origin
    && entry.rootPath === record.rootPath
  ))
  if (index >= 0) {
    registry.imports[index] = record
    return
  }
  registry.imports.push(record)
}

function operationMode(adopt: boolean, shouldImport: boolean): SetupOperationResult['mode'] {
  if (adopt && shouldImport) {
    return 'adopt-import'
  }
  if (adopt) {
    return 'adopt'
  }
  return 'import'
}

function canImportState(state: ResourceState): boolean {
  return state === RESOURCE_STATE_ADOPTABLE || state === RESOURCE_STATE_MANAGED || state === RESOURCE_STATE_USER_OWNED
}

function snapshotObservedPaths(rootPath: string, observedFiles: string[]): RecordedPath[] {
  return observedFiles
    .map((relativePath) => ({
      relativePath: relativePath.replaceAll(path.sep, '/'),
      fingerprint: pathFingerprint(path.join(rootPath, relativePath)),
    }))
    .sort((left, right) => left.relativePath.localeCompare(right.relativePath))
}

function snapshotMcpEntries(rootPath: string, observedFiles: string[]): ObservedMcpEntry[] {
  const entries: ObservedMcpEntry[] = []

  for (const relativePath of observedFiles) {
    const fullPath = path.join(rootPath, relativePath)
    if (directoryExists(fullPath) || !isMcpConfigCandidate(relativePath)) {
      continue
    }
    const parsed = readJsonLikeMap(fullPath)
    const mcpMap = extractMcpEntriesMap(parsed, path.basename(relativePath))
    if (!mcpMap) {
      continue
    }

    for (const name of Object.keys(mcpMap).sort()) {
      entries.push({
        entry: { name, configPath: relativePath.replaceAll(path.sep, '/'), state: RESOURCE_STATE_ADOPTABLE },
        fingerprint: jsonFingerprint(mcpMap[name]),
      })
    }
  }

  return entries.sort((left, right) => {
    if (left.entry.configPath !== right.entry.configPath) {
      return left.entry.configPath.localeCompare(right.entry.configPath)
    }
    return left.entry.name.localeCompare(right.entry.name)
  })
}

function readJsonLikeMap(filePath: string): Record<string, unknown> {
  const raw = fs.readFileSync(filePath, 'utf-8')
  if (filePath.endsWith('.jsonc')) {
    return JSON.parse(stripJsonComments(raw)) as Record<string, unknown>
  }
  return JSON.parse(raw) as Record<string, unknown>
}

function extractMcpEntriesMap(parsed: Record<string, unknown>, baseName: string): Record<string, unknown> | undefined {
  for (const key of ['mcpServers', 'mcp', 'mcp_servers']) {
    const value = parsed[key]
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      return value as Record<string, unknown>
    }
  }
  if (baseName === 'mcp.json') {
    const servers = parsed.servers
    if (servers && typeof servers === 'object' && !Array.isArray(servers)) {
      return servers as Record<string, unknown>
    }
  }
  return undefined
}

function isMcpConfigCandidate(relativePath: string): boolean {
  return ['settings.json', 'settings.local.json', 'mcp-config.json', 'mcp.json', 'opencode.json', 'opencode.jsonc'].includes(path.basename(relativePath))
}

function jsonFingerprint(value: unknown): string {
  return createHash('sha256').update(JSON.stringify(value)).digest('hex')
}

function pathFingerprint(targetPath: string): string {
  const stats = fs.statSync(targetPath)
  if (stats.isFile()) {
    return createHash('sha256').update(fs.readFileSync(targetPath)).digest('hex')
  }

  const entries: string[] = []
  walkDirectory(targetPath, targetPath, entries)
  return createHash('sha256').update(entries.sort().join('\n')).digest('hex')
}

function walkDirectory(rootPath: string, currentPath: string, entries: string[]): void {
  for (const entry of fs.readdirSync(currentPath, { withFileTypes: true })) {
    const absolutePath = path.join(currentPath, entry.name)
    const relativePath = path.relative(rootPath, absolutePath).replaceAll(path.sep, '/')
    if (entry.isDirectory()) {
      entries.push(`dir:${relativePath}`)
      walkDirectory(rootPath, absolutePath, entries)
      continue
    }
    entries.push(`file:${relativePath}:${createHash('sha256').update(fs.readFileSync(absolutePath)).digest('hex')}`)
  }
}

function importDirectoryName(scope: string, origin: string, rootPath: string): string {
  return `${scope}-${origin}-${createHash('sha256').update(rootPath).digest('hex').slice(0, 12)}`
}

function ensureDir(targetPath: string): void {
  fs.mkdirSync(targetPath, { recursive: true })
}

function copyPath(sourcePath: string, destinationPath: string): void {
  ensureDir(path.dirname(destinationPath))
  fs.cpSync(sourcePath, destinationPath, { recursive: true, force: true })
}

function pathsMatch(sourcePath: string, destinationPath: string): boolean {
  if (!pathExists(destinationPath)) {
    return false
  }
  return pathFingerprint(sourcePath) === pathFingerprint(destinationPath)
}

function createTimestampedBackup(targetPath: string): string {
  const backupPath = `${targetPath}.bak-${new Date().toISOString().replaceAll(':', '-')}`
  fs.renameSync(targetPath, backupPath)
  return backupPath
}

function importPath(sourcePath: string, destinationPath: string, operation: SetupOperationResult, backupSet: Set<string>): void {
  if (!pathExists(sourcePath)) {
    return
  }
  if (pathsMatch(sourcePath, destinationPath)) {
    return
  }
  if (pathExists(destinationPath)) {
    if (!backupSet.has(destinationPath)) {
      const backupPath = createTimestampedBackup(destinationPath)
      backupSet.add(destinationPath)
      operation.backups = [...(operation.backups ?? []), backupPath]
    } else {
      fs.rmSync(destinationPath, { recursive: true, force: true })
    }
  }
  copyPath(sourcePath, destinationPath)
}

function sortRegistry(registry: ScanRegistry): void {
  registry.resources?.sort((left, right) => `${left.targetId}:${left.scope}:${left.origin}:${left.rootPath}`.localeCompare(`${right.targetId}:${right.scope}:${right.origin}:${right.rootPath}`))
  registry.imports?.sort((left, right) => `${left.targetId}:${left.scope}:${left.origin}:${left.rootPath}`.localeCompare(`${right.targetId}:${right.scope}:${right.origin}:${right.rootPath}`))
}

function sortOperation(operation: SetupOperationResult): void {
  if (operation.backups) {
    operation.backups.sort()
  }
  if (operation.adopted) {
    operation.adopted.sort((left, right) => `${left.targetId}:${left.scope}:${left.origin}:${left.rootPath}`.localeCompare(`${right.targetId}:${right.scope}:${right.origin}:${right.rootPath}`))
  }
  if (operation.imported) {
    operation.imported.sort((left, right) => `${left.targetId}:${left.scope}:${left.origin}:${left.sourcePath}:${left.destinationPath}`.localeCompare(`${right.targetId}:${right.scope}:${right.origin}:${right.sourcePath}:${right.destinationPath}`))
  }
  if (operation.skipped) {
    operation.skipped.sort((left, right) => `${left.targetId}:${left.scope}:${left.origin}:${left.rootPath}:${left.reason}`.localeCompare(`${right.targetId}:${right.scope}:${right.origin}:${right.rootPath}:${right.reason}`))
  }
}

function writeRegistryIfChanged(registryPath: string, registry: ScanRegistry, operation: SetupOperationResult, backupSet: Set<string>): void {
  registry.version = SCAN_REGISTRY_VERSION
  sortRegistry(registry)
  const serialized = `${JSON.stringify(registry, null, 2)}\n`
  if (fileExists(registryPath) && fs.readFileSync(registryPath, 'utf-8') === serialized) {
    sortOperation(operation)
    return
  }
  ensureDir(path.dirname(registryPath))
  if (fileExists(registryPath) && !backupSet.has(registryPath)) {
    const backupPath = createTimestampedBackup(registryPath)
    backupSet.add(registryPath)
    operation.backups = [...(operation.backups ?? []), backupPath]
  }
  fs.writeFileSync(registryPath, serialized, 'utf-8')
  sortOperation(operation)
}

function runScanOperation(targetDir: string, homeDir: string, adopt: boolean, shouldImport: boolean): SetupInventoryResult & { operation?: SetupOperationResult } {
  const inventory = buildScanInventory(targetDir, homeDir)
  if (!adopt && !shouldImport) {
    return inventory
  }

  const aiSetupHomePath = aiSetupHome(homeDir)
  const registryPath = scanRegistryPath(aiSetupHomePath)
  const importRoot = importsRoot(aiSetupHomePath)
  const registry = loadRegistry(aiSetupHomePath)
  const operation: SetupOperationResult = {
    mode: operationMode(adopt, shouldImport),
    registryPath,
    importRoot,
  }
  const backupSet = new Set<string>()
  const adoptSeenRoots = new Map<string, boolean>()
  const importSeenRoots = new Map<string, boolean>()
  const now = new Date().toISOString()

  if (adopt) {
    for (const target of inventory.currentState.targets) {
      for (const detection of target.detections) {
        const rootKey = `${target.id}::${detection.rootPath}`
        if (adoptSeenRoots.has(rootKey)) {
          continue
        }
        if (detection.state !== RESOURCE_STATE_ADOPTABLE) {
          operation.skipped = [...(operation.skipped ?? []), {
            targetId: target.id,
            scope: detection.scope,
            origin: detection.origin,
            rootPath: detection.rootPath,
            state: detection.state,
            reason: 'not-adoptable',
          }]
          continue
        }
        const mcpEntries = snapshotMcpEntries(detection.rootPath, detection.observedFiles).map(({ entry, fingerprint }) => ({
          name: entry.name,
          configPath: entry.configPath,
          fingerprint,
        }))
        upsertResourceRecord(registry, {
          targetId: target.id,
          scope: detection.scope,
          origin: detection.origin,
          rootPath: detection.rootPath,
          state: 'managed',
          observedPaths: snapshotObservedPaths(detection.rootPath, detection.observedFiles),
          ...(mcpEntries.length > 0 ? { mcpEntries } : {}),
          updatedAt: now,
        })
        adoptSeenRoots.set(rootKey, true)
        operation.adopted = [...(operation.adopted ?? []), {
          targetId: target.id,
          scope: detection.scope,
          origin: detection.origin,
          rootPath: detection.rootPath,
        }]
      }
    }
  }

  if (shouldImport) {
    ensureDir(importRoot)
    for (const target of inventory.currentState.targets) {
      for (const detection of target.detections) {
        const rootKey = `${target.id}::${detection.rootPath}`
        if (importSeenRoots.has(rootKey)) {
          continue
        }
        if (!canImportState(detection.state)) {
          operation.skipped = [...(operation.skipped ?? []), {
            targetId: target.id,
            scope: detection.scope,
            origin: detection.origin,
            rootPath: detection.rootPath,
            state: detection.state,
            reason: 'not-importable',
          }]
          continue
        }

        const destinationRoot = path.join(importRoot, target.id, importDirectoryName(detection.scope, detection.origin, detection.rootPath))
        const importedPaths: string[] = []

        for (const relativePath of detection.observedFiles) {
          const sourcePath = path.join(detection.rootPath, relativePath)
          const destinationPath = path.join(destinationRoot, relativePath)
          importPath(sourcePath, destinationPath, operation, backupSet)
          if (!fileExists(sourcePath) && !directoryExists(sourcePath)) {
            continue
          }
          operation.imported = [...(operation.imported ?? []), {
            targetId: target.id,
            scope: detection.scope,
            origin: detection.origin,
            sourcePath,
            destinationPath,
          }]
          importedPaths.push(relativePath.replaceAll(path.sep, '/'))
        }

        upsertImportRecord(registry, {
          targetId: target.id,
          scope: detection.scope,
          origin: detection.origin,
          rootPath: detection.rootPath,
          importedPaths,
          destinationRoot,
          updatedAt: now,
        })
        importSeenRoots.set(rootKey, true)
      }
    }
  }

  writeRegistryIfChanged(registryPath, registry, operation, backupSet)
  return {
    ...buildScanInventory(targetDir, homeDir),
    operation: {
      ...operation,
      ...(operation.backups && operation.backups.length > 0 ? { backups: operation.backups } : {}),
      ...(operation.adopted && operation.adopted.length > 0 ? { adopted: operation.adopted } : {}),
      ...(operation.imported && operation.imported.length > 0 ? { imported: operation.imported } : {}),
      ...(operation.skipped && operation.skipped.length > 0 ? { skipped: operation.skipped } : {}),
    },
  }
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

  const agents = observeAgents(targetDir)

  return {
    mode: 'list',
    ...(scope ? { scopeFilter: scope } : {}),
    sharedPaths: buildSharedPaths(targetDir, homeDir, scope),
    targets,
    ...(agents.length > 0 ? { agents } : {}),
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
    .option('--adopt', 'Mark adoptable external configs as ai-setup managed')
    .option('--import', 'Import external configs into ~/.ai-setup reference storage')
    .option('--tool <tool>', 'Limit setup planning to specific tools (repeatable)', collectTool, [])
    .option('--all', 'Select all supported setup targets for the requested scope')
    .option('--global', 'Use global scope/home layout where supported')
    .action(async (opts: SetupOptions) => {
      const actions = [opts.scan === true, opts.list === true, opts.dryRun === true].filter(Boolean).length
      if (actions === 0) {
        if (opts.adopt === true || opts.import === true) {
          throw new Error('--adopt and --import require --scan')
        }
        throw new Error('no setup action selected (try: ai-setup setup --scan, --list, or --dry-run)')
      }

      if (actions > 1) {
        throw new Error('select exactly one of --scan, --list, or --dry-run')
      }

      if ((opts.adopt === true || opts.import === true) && opts.scan !== true) {
        throw new Error('--adopt and --import require --scan')
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
        console.log(JSON.stringify(runScanOperation(targetDir, homeDir, opts.adopt === true, opts.import === true), null, 2))
        return
      }

      if (opts.list === true) {
        console.log(JSON.stringify(buildListResult(selection, scope, targetDir, homeDir), null, 2))
        return
      }

      console.log(JSON.stringify(buildDryRunResult(selection, scope ?? 'project', targetDir, homeDir), null, 2))
    })
}
