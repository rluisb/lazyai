import type { ToolId } from '../types.js'
import { ClaudeCodeAdapter } from './claude-code.js'
import { CopilotAdapter } from './copilot.js'
import { OpenCodeAdapter } from './opencode.js'
import type { ToolAdapter } from './types.js'

export class AdapterRegistry {
  private adapters: Map<string, ToolAdapter> = new Map()

  constructor() {
    this.register(new OpenCodeAdapter())
    this.register(new ClaudeCodeAdapter())
    this.register(new CopilotAdapter())
  }

  register(adapter: ToolAdapter): void {
    this.adapters.set(adapter.getToolId(), adapter)
  }

  get(toolId: ToolId): ToolAdapter | undefined {
    return this.adapters.get(toolId)
  }

  getAll(toolIds: ToolId[]): ToolAdapter[] {
    return toolIds
      .map(id => this.get(id))
      .filter((a): a is ToolAdapter => a !== undefined)
  }

  getRegisteredIds(): string[] {
    return [...this.adapters.keys()]
  }
}
