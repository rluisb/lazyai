import type { WizardSelections as StoreWizardSelections } from './store/schema.js'

// Re-export store types (source of truth is zod schemas)
export type { StoreData, TrackedFile, Operation, Meta, Config, Sync } from './store/schema.js'
export type WizardSelections = StoreWizardSelections
export { CURRENT_SCHEMA_VERSION } from './store/schema.js'

export type SetupType = 'project' | 'workspace'
export type ToolId = 'pi' | 'opencode' | 'claude-code' | 'gemini' | 'copilot'

export interface SetupConfig {
  setupType: SetupType
  tools: ToolId[]
  projectName: string
  targetDir: string
  force?: boolean | undefined
}

export interface FileRecord {
  path: string
  hash: string
  source: string
}

/**
 * @deprecated Use StoreData from ./store/schema.js.
 * This legacy manifest shape is kept for backward compatibility only.
 */
export interface AiSetupConfig {
  version: string
  setupType: SetupType
  tools: ToolId[]
  projectName: string
  installedAt: string
  files: FileRecord[]
  selections?: WizardSelections
}

export type ArtifactType = 'agent' | 'skill' | 'command' | 'prompt' | 'template' | 'workflow'

export type DocsDirId =
  | 'features'
  | 'bugfixes'
  | 'refactors'
  | 'tech-debt'
  | 'adrs'
  | 'memory'
  | 'prompts'
  | 'standards'
  | 'templates'
  | 'rules'

export type AgentId = 'builder' | 'documenter' | 'planner' | 'red-team' | 'reviewer' | 'scout'
export type SkillId = 'anti-speculation' | 'implement' | 'iterate' | 'lessons-learned' | 'memory-write' | 'parallel-execution' | 'plan' | 'research' | 'tdd-loop'
export type PromptId = 'compact' | 'implement' | 'local-example' | 'plan' | 'research'
export type TemplateId = 'adr' | 'bugfix-rca-template' | 'code-review-template' | 'postmortem-template' | 'prd-template' | 'progress' | 'standard' | 'task' | 'tasks-template' | 'tech-debt-template' | 'techspec-template'
export type RuleId = 'access' | 'code-style' | 'cost' | 'review' | 'security' | 'testing' | 'workflow'
export type ToolAgentId = 'agents-dir' | 'root-dir' | 'skills-dir' | 'templates-dir'
export type InfraId = 'pre-commit' | 'compliance' | 'KNOWLEDGE_MAP'

export type ConflictStrategy = 'align' | 'backup-and-replace' | 'skip'

export interface WizardConfig extends SetupConfig {
  selections: WizardSelections
  interactive: boolean
}
