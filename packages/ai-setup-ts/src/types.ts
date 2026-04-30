import type { WizardSelections as StoreWizardSelections } from './store/schema.js'

export type { PresetLevel } from './presets.js'
// Re-export store types (source of truth is zod schemas)
export type { Config, Meta, Operation, StoreData, Sync, TrackedFile } from './store/schema.js'
export type WizardSelections = StoreWizardSelections
export { CURRENT_SCHEMA_VERSION } from './store/schema.js'

export type SetupScope = 'global' | 'workspace' | 'project'
/** @deprecated Use SetupScope */
export type SetupType = SetupScope
export type ToolId = 'opencode' | 'claude-code' | 'gemini' | 'copilot' | 'codex' | 'pi'

export interface SetupConfig {
  setupScope: SetupScope
  /** @deprecated Use setupScope */
  setupType?: SetupType
  tools: ToolId[]
  cliTools?: string[]
  enableServers?: string[]
  projectName: string
  workspaceName?: string
  workspaceRoot?: string
  targetDir: string
  planningRepoPath?: string
  planningDir?: string
  housekeeping?: import('./store/schema.js').HousekeepingConfig
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
  globalRef?: string
  driveCLI?: boolean
  localSecrets?: boolean
  force?: boolean | undefined
  dryRun?: boolean
}

export interface FileRecord {
  path: string
  hash: string
  source: string
  owner?: 'library' | 'user' | 'migrated'
  kind?: 'file' | 'symlink'
  linkTarget?: string
}

export interface RepoPermissions {
  read: boolean
  write: boolean
  runCommands: boolean
  runDestructive: boolean
  gitOperations: boolean
}

export const DEFAULT_REPO_PERMISSIONS: RepoPermissions = {
  read: true,
  write: true,
  runCommands: true,
  runDestructive: false,
  gitOperations: false,
}

/**
 * @deprecated Use StoreData from ./store/schema.js.
 * This legacy manifest shape is kept for backward compatibility only.
 */
export interface AiSetupConfig {
  version: string
  setupScope: SetupScope
  /** @deprecated Use setupScope */
  setupType?: SetupType
  tools: ToolId[]
  projectName: string
  installedAt: string
  files: FileRecord[]
  selections?: WizardSelections
}

export type ArtifactType = 'agent' | 'skill' | 'command' | 'prompt' | 'template' | 'workflow' | 'domain' | 'mode'

export type AgentId = 'builder' | 'documenter' | 'implementor' | 'orchestrator' | 'planner' | 'red-team' | 'reviewer' | 'scout'
export type SkillId = 'anti-speculation' | 'bugfix' | 'extract-standards' | 'housekeeping' | 'impact-check' | 'implement' | 'iterate' | 'memory-write' | 'orchestrate' | 'parallel-execution' | 'plan' | 'process-audit' | 'proof-of-concept' | 'research' | 'review' | 'rpi' | 'self-improve' | 'speckit-analyze' | 'speckit-checklist' | 'speckit-clarify' | 'speckit-constitution' | 'speckit-implement' | 'speckit-plan' | 'speckit-specify' | 'speckit-tasks' | 'spike' | 'tdd-loop' | 'update-memory'
export type PromptId = 'compact' | 'implement' | 'local-example' | 'plan' | 'research'
export type CommandId = 'rpi' | 'review' | 'plan' | 'speckit-analyze' | 'speckit-checklist' | 'speckit-clarify' | 'speckit-constitution' | 'speckit-implement' | 'speckit-plan' | 'speckit-specify' | 'speckit-tasks'
export type ChatModeId = 'architect' | 'reviewer'
export type OpenCodeCommandId = 'review' | 'test' | 'commit' | 'speckit.analyze' | 'speckit.checklist' | 'speckit.clarify' | 'speckit.constitution' | 'speckit.implement' | 'speckit.plan' | 'speckit.specify' | 'speckit.tasks'
export type OpenCodeModeId = 'plan' | 'audit'
export type TemplateId = 'adr' | 'bugfix-rca-template' | 'checklist-template' | 'code-review-template' | 'plan-template' | 'postmortem-template' | 'spec-template' | 'standard' | 'task' | 'tech-debt-template'
export type RuleId = 'access' | 'agent-security' | 'code-style' | 'cost' | 'review' | 'security' | 'testing' | 'tool-use' | 'workflow'
export type ToolAgentId = 'agents-dir' | 'root-dir' | 'skills-dir' | 'templates-dir'
export type InfraId = 'pre-commit' | 'compliance' | 'KNOWLEDGE_MAP' | 'codeowners'

export const ALL_AGENTS: AgentId[] = ['builder', 'documenter', 'implementor', 'orchestrator', 'planner', 'red-team', 'reviewer', 'scout']
export const ALL_SKILLS: SkillId[] = ['anti-speculation', 'bugfix', 'extract-standards', 'housekeeping', 'impact-check', 'implement', 'iterate', 'memory-write', 'orchestrate', 'parallel-execution', 'plan', 'process-audit', 'proof-of-concept', 'research', 'review', 'rpi', 'self-improve', 'speckit-analyze', 'speckit-checklist', 'speckit-clarify', 'speckit-constitution', 'speckit-implement', 'speckit-plan', 'speckit-specify', 'speckit-tasks', 'spike', 'tdd-loop', 'update-memory']
export const ALL_PROMPTS: PromptId[] = ['compact', 'implement', 'local-example', 'plan', 'research']
export const ALL_TEMPLATES: TemplateId[] = ['adr', 'bugfix-rca-template', 'checklist-template', 'code-review-template', 'plan-template', 'postmortem-template', 'spec-template', 'standard', 'task', 'tech-debt-template']
export const ALL_RULES: RuleId[] = ['access', 'agent-security', 'code-style', 'cost', 'review', 'security', 'testing', 'tool-use', 'workflow']
export const ALL_INFRA: InfraId[] = ['pre-commit', 'compliance', 'KNOWLEDGE_MAP', 'codeowners']
export const ALL_SPECS_DIRS: string[] = ['features', 'bugfixes', 'refactors', 'tech-debt', 'adrs', 'memory', 'prompts', 'standards', 'templates', 'rules']

export const ALL_COMMANDS: CommandId[] = ['rpi', 'review', 'plan', 'speckit-analyze', 'speckit-checklist', 'speckit-clarify', 'speckit-constitution', 'speckit-implement', 'speckit-plan', 'speckit-specify', 'speckit-tasks']
export const ALL_CHATMODES: ChatModeId[] = ['architect', 'reviewer']
export const ALL_OPENCODE_COMMANDS: OpenCodeCommandId[] = ['review', 'test', 'commit', 'speckit.analyze', 'speckit.checklist', 'speckit.clarify', 'speckit.constitution', 'speckit.implement', 'speckit.plan', 'speckit.specify', 'speckit.tasks']
export const ALL_OPENCODE_MODES: OpenCodeModeId[] = ['plan', 'audit']
export const OPENCODE_DESKTOP_COMMANDER_PLUGIN = '@opencode/desktop-commander'
export const OPENCODE_CONTEXT_FILES_PLUGIN = '@opencode/context-files'
export const OPENCODE_GIT_TOOLS_PLUGIN = '@opencode/git-tools'
export const ALL_OPENCODE_PLUGINS: string[] = [OPENCODE_DESKTOP_COMMANDER_PLUGIN, OPENCODE_CONTEXT_FILES_PLUGIN, OPENCODE_GIT_TOOLS_PLUGIN]

export type ConflictStrategy = 'align' | 'backup-and-replace' | 'skip'

export interface WizardConfig extends SetupConfig {
  selections: WizardSelections
  interactive: boolean
}
