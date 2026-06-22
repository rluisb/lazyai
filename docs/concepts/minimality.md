# Minimality

Minimality is the principle that every byte in the canonical embedded library
should earn its place. The canonical library is compiled into every supported
tool's agent context, so unnecessary content degrades model performance,
increases token costs, and risks hitting per-tool size limits.

## Report-only check

The minimality report at `packages/cli/internal/minimality/report.go` is an
**advisory, report-only** check. It surfaces growth trends without blocking
anything. Run it from the repository root:

```bash
go run ./packages/cli/internal/minimality/cmd/minimality-report
```

The report exits 0 even when advisory thresholds are exceeded. It measures:

- Go files over 500 lines, sorted by line count descending.
- Top-level CLI command registrations (visible, hidden, total).
- Canonical library byte count and the configured token-rent budget.
- Canonical asset counts per category (agents, skills, hooks, prompts,
  templates, rules).

See [Minimality report](minimality-report.md) for full output details and
advisory thresholds.

## Relationship to token rent

Token rent is the **hard gate**; minimality is the **early warning system**.

| Aspect | Token rent | Minimality report |
|---|---|---|
| Mode | Hard gate (exits non-zero on breach) | Advisory (exits 0 always) |
| Scope | Canonical library bytes only | Go file length, CLI commands, canonical assets |
| When to use | Pre-commit, CI | Periodic review, growth awareness |
| Override | `.lazyai/token-rent-override` | None needed (report-only) |

The minimality report reads token-rent budget data from the `tokenrent`
package and displays it, but it never enforces it. The two tools share data
but have different purposes.

## Exclusion and progressive disclosure

Not every asset in the library belongs in the canonical context. The curation
manifest (`packages/cli/library/manifests/curation.yaml`) records why each
asset is kept and whether it is token-rent-relevant.

### Why some assets are excluded from token-rent

Assets with `token_rent_relevant: false` fall into these categories:

1. **Setup-core guidance** — skills, templates, rules, and standards that
   guide the agent during setup or development but are not needed in every
   compiled context. Examples: `library/skills/anti-speculation.md`,
   `library/rules/code-style.md`.

2. **Adapter-support content** — full-parity skills emitted to tool-native
   locations (e.g. `library/skills/diagnose.md` emitted to `.agents/skills/`).
   These are shipped but live outside the canonical context.

3. **Archived or retired assets** — historical content kept for migration or
   reference but not emitted by active adapters. Listed under `exclusions` in
   the curation manifest.

4. **Progressive-disclosure assets** — harness improvement templates, rubrics,
   and trace taxonomy files that are available on demand but not loaded by
   default. Examples: `library/templates/trace-failure.md`,
   `library/rubrics/implementer.rubric.yaml`.

### Why some assets are progressive-disclosure only

Assets that are `token_rent_relevant: false` and `adapter_targets: [none]` are
available in the library but are **not emitted to any tool adapter by
default**. They are:

- **Available on request** — a user can add them to their setup via
  `lazyai add` or by copying them manually.
- **Not in the default context** — they do not consume token-rent budget and
  do not appear in every compiled output.
- **Documented in the library** — the curation manifest records why they exist
  and why they are not emitted by default.

This design keeps the default compiled context small while making the full
library discoverable.

## See also

- [Token rent](token-rent.md) — the hard byte budget for the canonical library.
- [Minimality report](minimality-report.md) — how to run the report and read
  its output.
- [Product boundaries](product-boundaries.md) — where minimality and token-rent
  fit in the dev-harness category.
- [Curation manifest](https://github.com/rluisb/lazyai/blob/main/packages/cli/library/manifests/curation.yaml) — the authoritative record of what is kept, excluded, and why.
