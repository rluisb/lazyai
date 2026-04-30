# Task 003: Add `ai-setup-orchestrator init` CLI Command

## Goal

Add a one-shot read-only command that prints project, host, root context, and catalog inventory.

## Files

- `src/cli/init.ts` (new)
- `src/index.ts`
- `src/__tests__/init-cli.case.ts` (new)

## Flags

```text
--task <text>
--host <host>
--project <path>
--json
--verbose
-h, --help
```

## Output Sections

Human-readable:

1. Project/context
2. Host CLI capability
3. Catalog inventory counts and names
4. Root context files: `AGENTS.md`, `CLAUDE.md`
5. Recommendation if `--task` is provided
6. Examples if no task is provided

JSON:

```ts
{
  projectRoot: string,
  host: CliContext,
  rootFiles: { agentsMd: boolean, claudeMd: boolean },
  inventory: { agents: string[], domains: string[], modes: string[], chains: string[], teams: string[], workflows: string[] },
  recommendation?: InitRecommendation
}
```

## Done When

- `ai-setup-orchestrator init --help` shows help.
- `ai-setup-orchestrator init --json` outputs valid JSON.
- Text output includes inventory and root files.
- Tests pass with injected temp catalog roots.
