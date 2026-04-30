# Task 005: Verification

## Goal

Validate implementation end-to-end.

## Commands

Run from `packages/orchestrator`:

```bash
npx vitest run src/__tests__/catalog-resolver.case.ts src/__tests__/init-cli.case.ts
npx vitest run
npx tsc --noEmit
npm run build
```

## Manual Smoke Tests

```bash
ai-setup-orchestrator init
ai-setup-orchestrator init --json
ai-setup-orchestrator init --host claude-code --task "review auth code"
ai-setup-orchestrator init --host opencode --task "build auth from scratch"
```

## Done When

- All tests pass.
- Typecheck passes.
- Build passes.
- Manual smoke output is readable and recommendations are reasonable.
