# [YOUR_PROJECT_NAME] Quality Gates

## Gate 1: Pre-Commit
- [ ] Lint passes
- [ ] Type check passes
- [ ] Unit tests pass

## Gate 2: Pre-Merge
- [ ] All CI checks green
- [ ] Code review approved
- [ ] Coverage thresholds met

## Gate 3: Pre-Deploy
- [ ] Integration tests pass
- [ ] No P0/P1 regressions
- [ ] Rollback plan documented

## Thresholds
| Metric | Target | Minimum |
|--------|--------|---------|
| Line coverage | 90% | 80% |
| Branch coverage | 85% | 75% |
| Build time | <5min | <10min |
