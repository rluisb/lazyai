# Token rent

Token rent is the hard byte budget for the canonical embedded library at
`packages/cli/library/canonical/`. It prevents the compiled agent context from
growing large enough to degrade model performance or exceed per-tool size limits.

## Budget

The default budget is **50,000 bytes** (`DefaultBudgetBytes`). Every file under
`packages/cli/library/canonical/` counts toward this total, with two
exclusions:

- `.gitkeep` files (directory placeholders)
- Any file whose name starts with `.` (dotfiles)

The budget is enforced by the `tokenrent` package at
`packages/cli/internal/tokenrent/check.go`. The canonical subdirectory constant
is `CanonicalSubdir = "packages/cli/library/canonical"`.

## Enforcement

Token rent is a **hard gate**: the check exits non-zero when the budget is
exceeded without a valid override. Run it from the repository root:

```bash
go run ./packages/cli/internal/tokenrent/cmd/token-rent-check
```

The check is also wired into the pre-commit hook so violations are caught
before push.

## Override mechanism

When the canonical library legitimately needs to exceed the default budget,
add a file at `.lazyai/token-rent-override` with YAML frontmatter:

```yaml
budget: 60000
reason: "New canonical skill for X feature"
approved_by: "user@example.com"
expires: "2026-09-01"
```

- `reason` is **required** and must be non-empty.
- `budget` is optional; if set, it becomes the effective budget ceiling.
- `approved_by` and `expires` are optional metadata for auditability.

An override with an empty `reason` is rejected with an `OverrideError`. If no
override file exists and the budget is exceeded, a `BudgetError` is returned.

## Product boundary

The `tokenrent` package is categorized as `dev-harness` in the product
boundaries inventory. It is a repository-maintainer tool, not a shipped user
command. The `minimality` report references token-rent budget data but does not
enforce it.

## Curation manifest integration

Every entry in the curation manifest
(`packages/cli/library/manifests/curation.yaml`) carries a
`token_rent_relevant` boolean field. This field is **required** and validated
by `ValidateCurationManifest`:

- `true` — the asset lives under `canonical/` and counts toward the token-rent
  budget. Examples: canonical agents, canonical skills, canonical hooks.
- `false` — the asset lives outside `canonical/` (e.g. `library/skills/`,
  `library/templates/`, `library/rules/`) and does not count toward the budget.
  These assets are either setup-core guidance or adapter-support content that
  is emitted to tool-native locations, not embedded in the canonical context.

This field makes exclusion and progressive-disclosure decisions explicit and
machine-verifiable. A `false` value documents a deliberate choice: the asset
is valuable enough to ship but not valuable enough to consume token-rent
budget in every compiled output.

## See also

- [Minimality](minimality.md) — the advisory report-only counterpart that
  surfaces growth trends before they hit the hard gate.
- [Minimality report](minimality-report.md) — how to run the report and read
  its output.
- [Product boundaries](product-boundaries.md) — where token-rent fits in the
  dev-harness category.
