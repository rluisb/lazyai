# Task 001: Merge DB Chain/Team/Workflow Definitions into Runtime Catalog

## Goal

Ensure active internal DB catalog definitions for `chain`, `team`, and `workflow` are available to runtime handlers and the future `init` command.

## Files

- `src/catalog/resolver.ts`
- `src/__tests__/catalog-resolver.case.ts`

## Implementation Notes

Current resolver merges DB definitions for:

- `agent`
- `skill`

Add support for:

- `chain`
- `team`
- `workflow`

Each DB row body should be parsed as JSON and mapped to the relevant definition type while normalizing metadata:

- `kind`
- `name`
- `description`
- `version`
- `source: 'db'`
- `path: 'catalog://<kind>/<name>'` or empty path consistent with current DB agent/skill mapping

## Done When

- DB-created active team appears in `resolveCatalog(...).teams`.
- DB-created active chain appears in `resolveCatalog(...).chains`.
- DB-created active workflow appears in `resolveCatalog(...).workflows`.
- Inactive/deactivated DB definitions are skipped.
- Resolver tests pass.
