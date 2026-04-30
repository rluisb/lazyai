import type { AgentId, ConflictStrategy, FileRecord, PromptId, SetupScope, SkillId } from '../types.js'

export interface AdapterContext {
  targetDir: string
  setupScope?: SetupScope
  workspaceRoot?: string
  homeDir?: string
  libraryDir: string
  fileRecords: FileRecord[]
  enableServers?: string[]
  localSecrets?: boolean
  driveCLI?: boolean
  driveCli?: boolean
  force?: boolean | undefined
  dryRun?: boolean
  strategy?: ConflictStrategy
  perFileOverrides?: Map<string, ConflictStrategy>
  installMode?: 'copy' | 'symlink'
  selections?: {
    agents?: AgentId[]
    skills?: SkillId[]
    prompts?: PromptId[]
    opencodePlugins?: string[]
  }
}

export interface ToolAdapter {
  getToolId(): string
  install(ctx: AdapterContext): Promise<void>
  remove(ctx: AdapterContext): Promise<void>
}
