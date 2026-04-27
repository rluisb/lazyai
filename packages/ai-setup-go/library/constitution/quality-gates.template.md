# [YOUR_PROJECT_NAME] Quality Gates

## Gate 1: Pre-Commit
- [ ] Lint passes
- [ ] Type check passes
- [ ] Unit tests pass

<!-- Example commands:
```bash
npm run lint        # or: bundle exec rubocop
npm run typecheck   # or: tsc --noEmit
npm run test        # or: bundle exec rspec
```
-->

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
| Line coverage | [YOUR_LINE_COVERAGE_TARGET] | [YOUR_LINE_COVERAGE_MINIMUM] |
| Branch coverage | [YOUR_BRANCH_COVERAGE_TARGET] | [YOUR_BRANCH_COVERAGE_MINIMUM] |
| Build time | [YOUR_BUILD_TIME_TARGET] | [YOUR_BUILD_TIME_MAXIMUM] |

<!-- Customize thresholds for your stack. These are starting points.
For Teachable repos: fedora (Rails) uses rubocop + rspec,
creator-checkout (Next.js) targets client 85%/server 90%,
school-plan-service (Go) uses go test + go vet. -->
