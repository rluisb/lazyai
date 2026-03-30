import type { FileRecord } from '../types.js'

export interface AdapterContext {
  targetDir: string
  libraryDir: string
  fileRecords: FileRecord[]
}

export interface ToolAdapter {
  getToolId(): string
  install(ctx: AdapterContext): Promise<void>
  remove(ctx: AdapterContext): Promise<void>
}
