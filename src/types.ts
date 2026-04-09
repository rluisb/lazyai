import type { WizardSelections as StoreWizardSelections } from './store/schema.js'

export type { PresetLevel } from './presets.js'
// Re-export store types (source of truth is zod schemas)
export type { Config, Meta, Operation, StoreData, Sync, TrackedFile } from './store/schema.js'
export type WizardSelections = StoreWizardSelections
export { CURRENT_SCHEMA_VERSION } from './store/schema.js'

export type SetupScope = 'global' | 'workspace' | 'project'
/** @deprecated Use SetupScope */
export type SetupType = SetupScope
export type ToolId = 'opencode' | 'claude-code' | 'gemini' | 'copilot' | 'codex'

export interface SetupConfig {
  setupScope: SetupScope
  /** @deprecated Use setupScope */
  setupType?: SetupType
  tools: ToolId[]
  cliTools?: string[]
  enableServers?: string[]
  projectName: string
  workspaceName?: string
  targetDir: string
  planningRepoPath?: string
  repos?: Array<{ name: string; path: string; type?: string; description?: string }>
  globalRef?: string
  force?: boolean | undefined
  dryRun?: boolean
}

export interface FileRecord {
  path: string
  hash: string
  source: string
  owner?: 'library' | 'user' | 'migrated'
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

export type AgentId = 'builder' | 'documenter' | 'orchestrator' | 'planner' | 'red-team' | 'reviewer' | 'scout'
export type SkillId = 'anti-speculation' | 'extract-standards' | 'implement' | 'iterate' | 'memory-write' | 'parallel-execution' | 'plan' | 'research' | 'tdd-loop'
export type PromptId = 'compact' | 'implement' | 'local-example' | 'plan' | 'research'
export type TemplateId = 'adr' | 'bugfix-rca-template' | 'checklist-template' | 'code-review-template' | 'plan-template' | 'postmortem-template' | 'spec-template' | 'standard' | 'task' | 'tech-debt-template'
export type RuleId = 'access' | 'agent-security' | 'code-style' | 'cost' | 'review' | 'security' | 'testing' | 'tool-use' | 'workflow'
export type ToolAgentId = 'agents-dir' | 'root-dir' | 'skills-dir' | 'templates-dir'
export type InfraId = 'pre-commit' | 'compliance' | 'KNOWLEDGE_MAP' | 'codeowners'

export const ALL_AGENTS: AgentId[] = ['builder', 'documenter', 'orchestrator', 'planner', 'red-team', 'reviewer', 'scout']
export const ALL_SKILLS: SkillId[] = ['anti-speculation', 'extract-standards', 'implement', 'iterate', 'memory-write', 'parallel-execution', 'plan', 'research', 'tdd-loop']
export const ALL_PROMPTS: PromptId[] = ['compact', 'implement', 'local-example', 'plan', 'research']
export const ALL_TEMPLATES: TemplateId[] = ['adr', 'bugfix-rca-template', 'checklist-template', 'code-review-template', 'plan-template', 'postmortem-template', 'spec-template', 'standard', 'task', 'tech-debt-template']
export const ALL_RULES: RuleId[] = ['access', 'agent-security', 'code-style', 'cost', 'review', 'security', 'testing', 'tool-use', 'workflow']
export const ALL_INFRA: InfraId[] = ['pre-commit', 'compliance', 'KNOWLEDGE_MAP', 'codeowners']
export const ALL_SPECS_DIRS: string[] = ['features', 'bugfixes', 'refactors', 'tech-debt', 'adrs', 'memory', 'prompts', 'standards', 'templates', 'rules']

export type ConflictStrategy = 'align' | 'backup-and-replace' | 'skip'

export interface WizardConfig extends SetupConfig {
  selections: WizardSelections
  interactive: boolean
}
