# Task 004: Add Rule-Based Init Recommendations

## Goal

Provide deterministic recommendations for direct-agent, chain, team, or workflow.

## Files

- `src/cli/init.ts`
- `src/__tests__/init-cli.case.ts`

## Rules

1. Review/audit/security task:
   - prefer team when host supports parallel teams and suitable team/agents exist
   - otherwise prefer reviewer/red-team direct agent or review chain
2. Design/architecture/from-scratch task:
   - prefer workflow/chain with architect/planner/implementor
3. Build/implement/refactor task:
   - prefer RPI workflow/chain if available
   - otherwise direct implementor-senior
4. No task:
   - no recommendation; print examples

## Recommendation Shape

```ts
interface InitRecommendation {
  kind: 'direct-agent' | 'chain' | 'team' | 'workflow'
  name?: string
  confidence: 'low' | 'medium' | 'high'
  reason: string
  nextCommand?: string
  alternatives: Array<{ kind: string; name?: string; reason: string }>
}
```

## Done When

- Claude Code + review/multi-agent task can recommend team.
- OpenCode + multi-step task can recommend chain.
- No matching definitions produces a low-confidence direct-agent or examples-only fallback.
- Tests cover at least team vs chain host-aware branching.
