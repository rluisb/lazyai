import type { ToolId } from '../types.js'

export interface OrchestratorHintPaths {
  agentPath: string
  skillPath: string
}

export function orchestratorHintPaths(_tool: ToolId): OrchestratorHintPaths {
  return {
    agentPath: '.opencode/agents/orchestrator.md',
    skillPath: '.opencode/skills/orchestrate/SKILL.md',
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
