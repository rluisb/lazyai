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

export interface AiSetupConfig {
  version: string
  setupType: SetupType
  tools: ToolId[]
  projectName: string
  installedAt: string
  files: FileRecord[]
}
