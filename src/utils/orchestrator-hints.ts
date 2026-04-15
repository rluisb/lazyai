import type { ToolId } from '../types.js'

export interface OrchestratorHintPaths {
  agentPath: string
  skillPath: string
}

export function orchestratorHintPaths(tool: ToolId): OrchestratorHintPaths {
  switch (tool) {
    case 'opencode':
      return {
        agentPath: '.opencode/agents/orchestrator.md',
        skillPath: '.opencode/skills/orchestrate/SKILL.md',
      }
    case 'claude-code':
      return {
        agentPath: '.claude/agents/orchestrator.md',
        skillPath: '.claude/skills/orchestrate/SKILL.md',
      }
    case 'gemini':
      return {
        agentPath: '.gemini/skills/orchestrator/SKILL.md',
        skillPath: '.gemini/skills/orchestrate/SKILL.md',
      }
    case 'codex':
      return {
        agentPath: '.agents/skills/orchestrator/SKILL.md',
        skillPath: '.agents/skills/orchestrate/SKILL.md',
      }
    case 'copilot':
      return {
        agentPath: '.github/prompts/orchestrator.prompt.md',
        skillPath: '.github/prompts/orchestrate.prompt.md',
      }
  }
}

export function formatOrchestratorHintBody(tool: ToolId): string {
  const { agentPath, skillPath } = orchestratorHintPaths(tool)
  return [
    'Try it out in your host CLI:',
    '  > Use the orchestrator agent to list available chains.',
    '  > Use the orchestrator to start the feature chain for <your task>.',
    '',
    `Agent file: ${agentPath}`,
    `Procedure skill: ${skillPath}`,
    'Chains, teams, and domains: .ai/orchestration/',
  ].join('\n')
}
