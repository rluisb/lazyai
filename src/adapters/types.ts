import type { FileRecord, AgentId, SkillId, PromptId, ConflictStrategy } from '../types.js'

export interface AdapterContext {
  targetDir: string
  libraryDir: string
  fileRecords: FileRecord[]
  force?: boolean | undefined
  strategy?: ConflictStrategy
  perFileOverrides?: Map<string, ConflictStrategy>
  selections?: {
    agents?: AgentId[]
    skills?: SkillId[]
    prompts?: PromptId[]
  }
}

export interface ToolAdapter {
  getToolId(): string
  install(ctx: AdapterContext): Promise<void>
  remove(ctx: AdapterContext): Promise<void>
}
