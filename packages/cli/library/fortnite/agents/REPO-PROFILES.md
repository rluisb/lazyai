# Repo Profiles

Quality gates by repository. Used by zero-point and build-mode skills.

| Repo | Stack | Quality Gates |
|------|-------|---------------|
| **fedora** | Ruby/Rails | `bundle exec rubocop`, `bundle exec rspec` |
| **school-plan-service** | Go | `go test ./...`, `go vet ./...` |
| **creator-checkout** | Next.js | `npm run quality`, `npm run build` |
| **mono-frontend** | React/TS | `yarn lint`, `yarn typecheck`, `yarn test`, `yarn build` |
| **oauth-service** | Ruby | `bundle exec rubocop`, `bundle exec rspec` |

## CLI Tool Wiring

| Tool | Purpose | Primary Agent |
|------|---------|---------------|
| `ob` | Obsidian vault research | loot-hawk |
| `qmd` | Vault BM25+vector search | loot-hawk, turbo-crank, shield-audit, loop-driver |
| `rtk` | Session checkpoint/handoff | All agents (via slurp-juice) |
| `colima` | Docker container runtime | rift-deploy |
| `hotctl` | Hotmart infra CLI | wall-builder |
| `dev` | Teachable dev tool | rift-deploy |
