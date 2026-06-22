# RPI Cycle 2 — Testing and Final Report

## Required targeted tests

Run targeted tests first:

```bash
go test -v ./packages/cli/internal/compiler -run 'Test.*Skill|Test.*Agent|Test.*Validate'
go test -v ./packages/cli/internal/validate/...
go test -v ./packages/cli/cmd/...
```

Then run broader tests if safe:

```bash
go test ./...
go vet ./...
```

If `go vet` or `go test ./...` fails for unrelated existing reasons, report the exact failure and whether targeted tests passed.

Check working tree:

```bash
git status --short
```

## Final report format

```markdown
# LazyAI / vibe-lab RPI Cycle 2 Report — Semantic Skill and Agent Validation

## 1. Research summary

- Repo state:
- Current validation paths:
- Current skill format:
- Current agent format:
- Existing tests:
- Constraints confirmed:

## 2. Plan executed

| Item | Status | Notes |
|---|---|---|

## 3. Changes made

| Area | Files changed | Summary |
|---|---|---|

## 4. Skill validation

- New checks:
- Warning rules:
- Error rules:
- Tests:
- Remaining gaps:

## 5. Agent validation

- New checks:
- Warning rules:
- Error rules:
- Tests:
- Remaining gaps:

## 6. Validation output quality

- Rule IDs added:
- Example output:
- Fix suggestions:

## 7. Docs/templates

- Added/updated:
- Linked from:
- Remaining gaps:

## 8. Tests and validation

Commands run:

```bash
...
```

Results:

```text
...
```

## 9. Changed files

```text
...
```

## 10. Risks

| Risk | Impact | Mitigation |
|---|---|---|

## 11. Remaining work

| Item | Why not completed | Suggested next step |
|---|---|---|

## 12. Product boundary confirmation

Confirm explicitly:

- No runtime/orchestration surface was added.
- No old task/workflow/eval command was reintroduced.
- No mandatory judge/scoring engine was added.
- No mandatory trace daemon was added.
- No mandatory RAG core was added.
- LazyAI remains a harness asset manager/compiler.
- vibe-lab remains the process/quality asset layer.
```

## Definition of done

```text
- Skill semantic validation exists or is clearly improved.
- Agent contract validation exists or is clearly improved.
- Validation messages are actionable and include rule IDs.
- Tests cover valid and invalid skill/agent cases.
- Docs/templates explain skill and agent quality.
- Existing shipped assets are not broken.
- Targeted tests pass.
- LazyAI remains a compiler/asset manager, not a runtime.
```
