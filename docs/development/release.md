# Release Process

This document describes the safe release flow for `@ricardoborges-teachable/ai-setup`.

## Do not publish to npm

The repository currently includes `.github/workflows/publish.yml`, which publishes to npm on release automation paths. For current releases, the goal is to prepare release assets and a GitHub release draft **without** publishing the package.

## Recommended release preparation steps

1. Review the version bump in `package.json` and `package-lock.json`.
2. Review `CHANGELOG.md` for final release notes.
3. Merge the release preparation branch through the normal PR process.
4. After merge, create a **draft** GitHub release only.
5. Do **not** click publish and do **not** run `npm publish`.

## Draft GitHub release

### Option A: GitHub UI

1. Open the repository on GitHub.
2. Go to **Releases**.
3. Choose **Draft a new release**.
4. Create or select the version tag.
5. Set the release title.
6. Paste notes from `CHANGELOG.md`.
7. Save as draft.
8. Do not publish the release.

### Option B: GitHub CLI

Only use this if you have confirmed that creating a draft release will not trigger unintended publish automation in your repository settings/workflows.

```bash
gh release create vX.Y.Z \
  --draft \
  --title "ai-setup vX.Y.Z" \
  --notes-file CHANGELOG.md
```

## Important safety note

Because `.github/workflows/publish.yml` is configured for release-related automation and npm publishing, treat the GitHub release as a documentation artifact for now. Keep the release in draft state until you intentionally want an npm publication flow.

## Orchestrator binary assets

The release workflow publishes `ai-setup-orchestrator-*` assets for the same supported platforms as the main `ai-setup-*` binary. Both sets are included in `checksums.txt` so released installs can verify cached orchestrator downloads.

Released `ai-setup` binaries request the orchestrator assets from the matching `v{version}` release tag. Development builds fall back to the latest release.

## Upgrading the CLI

End users can upgrade with:

```bash
ai-setup update-self --check
ai-setup update-self --dry-run
ai-setup update-self
```

After upgrading the binary, refresh managed files:

```bash
ai-setup update --check
ai-setup update
ai-setup doctor
```

## Store migration

When the Go binary detects an existing `.ai-setup.json` without a `.ai-setup.db`, it automatically imports the configuration into SQLite. This happens on any command (`init`, `update`, `doctor`, `status`). No user action is needed. The `.ai-setup.json` file is left in place for backward compatibility.
