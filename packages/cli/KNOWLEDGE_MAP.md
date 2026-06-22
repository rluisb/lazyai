Authoritative knowledge map: specs/KNOWLEDGE_MAP.md. This file is scoped/secondary.

# Project Knowledge Map

> Navigable index of the setup-core CLI implementation.
> Update when a feature changes the compile contract, package seams, or durable terminology.

---

## Architecture Decisions

| ADR | Decision | Feature | Status |
|-----|----------|---------|--------|
| [specs/adrs/004-vibe-lab-alignment-contract.md](../../specs/adrs/004-vibe-lab-alignment-contract.md) | Capability-first conformance to vibe-lab-compatible tool surfaces | 026 | Accepted |
| [specs/adrs/005-core-vs-optional-modules.md](../../specs/adrs/005-core-vs-optional-modules.md) | setup-core is the default product; runtime-adjacent commands are optional modules | 026 | Accepted |
| [specs/adrs/006-manifest-driven-compile-and-seven-target-contract.md](../../specs/adrs/006-manifest-driven-compile-and-seven-target-contract.md) | `.ai/lazyai.json` + `.ai/lock.json` define the V2 compile contract; supported targets are exactly seven; binary stays `lazyai-cli` | 029 | Accepted |

## Active Features

| ID | Name | Status | ADRs |
|----|------|--------|------|
| [specs/029-lazyai-v2/](../../specs/029-lazyai-v2/) | Manifest-driven compile, lockfile, consolidated validators, migration/eject, multi-target plugin bundles, local eval validation, and `.ai/` v1 schema freeze | Done | 004, 005, 006 |

## Active Refactors

| ID | Name | Status | ADRs |
|----|------|--------|------|
| [specs/refactors/026-vibe-lab-alignment/](../../specs/refactors/026-vibe-lab-alignment/) | Exact vibe-lab baseline parity for setup-core surfaces; runtime extras demoted to optional modules | Done | 004, 005 |

## Rules & Standards

| Type | Files | Purpose |
|------|-------|---------|
| Rules | `specs/rules/*.md` | Prescriptive — WHAT to do |
| Standards | `specs/standards/*.md` | Descriptive — HOW we do it |

## Key Modules

| Path | Responsibility | Owner |
|------|---------------|-------|
| `packages/cli/cmd/compile.go` | Manifest-driven compile entrypoint; resolves targets and writes `.ai/lock.json` | setup-core |
| `packages/cli/internal/aimanifest/` | `.ai/lazyai.json` load/save/validate/target resolution | setup-core |
| `packages/cli/internal/lockfile/` | `.ai/lock.json` model, hashing, and generated-output tracking | setup-core |
| `packages/cli/internal/writer/` | Managed-region writes with drift detection and lock updates | setup-core |
| `packages/cli/internal/validate/` | Canonical `.ai/` validators, secret scanning, path/symlink safety, doctor security inputs | setup-core |
| `packages/cli/internal/migration/` + `internal/eject/` | Import from native setups, preserve raw source files, strip LazyAI management metadata on eject | setup-core |
| `packages/cli/internal/plugin/` + `internal/evals/` | Multi-target plugin bundles and local eval validation | setup-core |
| `packages/cli/library/` | Embedded canonical assets and MCP catalog | setup-core |

## Terminology

### Accepted domain terms

| Term | Meaning | Source of truth |
|------|---------|-----------------|
| setup-core | The default lazyai-cli command set (init, compile, update, doctor, add, build-plugin, etc.) | `specs/adrs/005-core-vs-optional-modules.md` |
| runtime-adjacent module | Optional command families (session, message, ledger, memory, auth, cost, metrics, notify, secret, backup, restore-runtime-db, git) | `specs/adrs/005-core-vs-optional-modules.md` |
| manifest v1 | `.ai/lazyai.json` schema version `1.0` contract for targets, profile, source/library, adapter options, and safety | `packages/cli/internal/aimanifest/` |
| compile lockfile | `.ai/lock.json` schema version `1.0` record of adapter metadata and generated-output hashes | `packages/cli/internal/lockfile/` |
| artifact type | A category of generated asset: agent, skill, command, prompt, template, rule, infra, specs-dir | `packages/cli/internal/validation/validation.go` |
| library | Embedded reference content used by lazyai-cli to populate target repos | `packages/cli/library/` |
| curation manifest | YAML manifest of every embedded library asset with provenance metadata | `packages/cli/library/manifests/curation.yaml` |

### Vocabulary source of truth

Runtime must not introduce a dedicated terminology lookup subsystem: the source of truth stays in this Markdown section, with derived enums (e.g. `packages/cli/internal/types/types.go`) hand-mapped from the table above. If a lookup feels necessary, extend the table instead.
