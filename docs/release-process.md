# Release Process

This document describes the safe release flow for `@ricardoborges-teachable/ai-setup` version `0.2.0`.

> Do **not** publish to npm as part of this process.

## Why this is documented separately

The repository currently includes `.github/workflows/publish.yml`, which publishes to npm on release automation paths. For `v0.2.0`, the goal is to prepare release assets and a GitHub release draft **without** publishing the package.

## Files updated for v0.2.0

- `package.json` → bumped to `0.2.0`
- `package-lock.json` → aligned to `0.2.0`
- `CHANGELOG.md` → added Migration Engine release notes

## Recommended release preparation steps

1. Review the version bump in `package.json` and `package-lock.json`.
2. Review `CHANGELOG.md` for the final release notes.
3. Merge the release preparation branch through the normal PR process.
4. After merge, create a **draft** GitHub release only.
5. Do **not** click publish and do **not** run `npm publish`.

## Draft GitHub release body

Use the following title and notes when creating the draft release:

### Title

`ai-setup v0.2.0`

### Notes

```md
## Migration Engine

- Introduces the new Migration Engine to help teams import existing AI assistant setups into `ai-setup`.
- Supports migration detection for OpenCode, Claude Code, Pi, Gemini CLI, and GitHub Copilot.
- Adds preview mode and multiple merge strategies (`smart`, `preserve`, `replace`, `append`).
- Includes backup-aware execution and migration drift checks through `ai-setup doctor --migration-check`.
- Opens the door for community and custom parser extensions.
```

## Safe draft-only release options

### Option A: GitHub UI

1. Open the repository on GitHub.
2. Go to **Releases**.
3. Choose **Draft a new release**.
4. Create or select tag `v0.2.0`.
5. Set the release title to `ai-setup v0.2.0`.
6. Paste the notes above.
7. Save as draft.
8. Do not publish the release.

### Option B: GitHub CLI

Only use this if you have confirmed that creating a draft release will not trigger unintended publish automation in your repository settings/workflows.

```bash
gh release create v0.2.0 \
  --draft \
  --title "ai-setup v0.2.0" \
  --notes-file CHANGELOG.md
```

## Important safety note

Because `.github/workflows/publish.yml` is configured for release-related automation and npm publishing, treat the GitHub release as a documentation artifact for now. Keep the release in draft state until you intentionally want an npm publication flow.

## Orchestrator binary assets

The release workflow publishes `ai-setup-orchestrator-*` assets for the same supported platforms as the main `ai-setup-*` binary and includes both sets in `checksums.txt` so released installs can verify cached orchestrator downloads. Released `ai-setup` binaries request the orchestrator assets from the matching `v{version}` release tag; development builds fall back to the latest release.
