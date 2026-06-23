# Release Verification Matrix

> **Last updated:** 2026-06-23
> **Issue:** [#350](https://github.com/rluisb/lazyai/issues/350)
> **Status:** Documented matrix with CI automation for Linux, macOS, and Windows; macOS arm64 also verified locally.

## Platform Coverage

| Platform | CI Runner | Go Test | Build | MkDocs | Smoke (init/validate/compile) | Status |
|---|---|---|---|---|---|---|
| **Linux (ubuntu-latest)** | `.github/workflows/cross-platform.yml`, existing CI | Yes | Yes | Yes | Yes | CI workflow added; existing Linux CI already active |
| **macOS (darwin/arm64)** | `.github/workflows/cross-platform.yml` (`macos-latest`) | Yes | Yes | N/A | N/A | Verified locally (2026-06-23); CI workflow added |
| **macOS (darwin/amd64)** | Cross-compile only | Manual | Cross-compile | N/A | Manual | Not runtime-verified |
| **Windows (amd64)** | `.github/workflows/cross-platform.yml` (`windows-latest`) | Yes | Yes | N/A | N/A | CI workflow added; first run pending |

## Verification Commands

Every platform must pass these commands before a release is considered verified:

```bash
# 1. Go tests (all packages)
cd packages/cli && go test ./... -count=1 -timeout 120s
cd packages/diffviewer && go test ./... -count=1 -timeout 60s

# 2. Build
make build

# 3. MkDocs (Python/mkdocs required)
pip install -r docs/requirements.txt
mkdocs build --strict

# 4. CLI smoke test
chmod +x tests/scripts/smoke-test.sh
./tests/scripts/smoke-test.sh
```

## Smoke Scenario

The smoke test (`tests/scripts/smoke-test.sh`) exercises:

1. **Build** — `go build ./cmd/...` succeeds
2. **Doctor** — `doctor --help` shows expected flags
3. **Session** — `session start` creates a session
4. **Ledger** — `ledger init`, `ledger append`, `ledger verify` in a temp dir
5. **Validate** — `validate agents` returns results

For a full release smoke, additionally run against a disposable project:

```bash
# Create a disposable project
TMP=$(mktemp -d)
cd "$TMP"
git init
mkdir -p specs

# Init (non-interactive)
/path/to/lazyai-cli init --no-interactive --scope project --tools opencode

# Validate all
/path/to/lazyai-cli validate --all

# Compile
/path/to/lazyai-cli compile

# Cleanup
rm -rf "$TMP"
```

## Platform-Specific Notes

### macOS

- **darwin/arm64** (Apple Silicon): Verified 2026-06-23. All 43 Go test packages pass, `make build` succeeds, `mkdocs build --strict` passes (info-level "not in nav" warnings only). The disposable-project smoke remains Linux-only until a portable smoke harness exists.
- **darwin/amd64** (Intel): Not verified. Cross-compilation via `make cli-snapshot` confirms build succeeds, but no test execution has been done on Intel macOS.
- **Known macOS-specific code**: `notify.go` (AppleScript notifications), `secret.go` (Keychain), `doctor_health.go` (disk usage via `df`).

### Linux

- **amd64**: Verified in existing CI on ubuntu-latest. The new cross-platform workflow also runs Go tests/builds and the smoke script on Linux.
- **arm64**: Cross-compiled in release workflow only. No CI test execution on Linux arm64 runners.
- **Known Linux-specific code**: `notify.go` (libnotify via `notify-send`), `secret.go` (secret-tool/libsecret).

### Windows

- **amd64**: The new cross-platform workflow runs Go tests and builds on `windows-latest`; first run pending after this PR opens.
- **Known Windows-specific code**: `update-self.go` (`.exe` extension, `.new` rename pattern), `notify.go` (unsupported — returns error), `secret.go` (unsupported — returns error).
- **Shell scripts**: All test scripts use `#!/usr/bin/env bash` and are not compatible with Windows without WSL or Git Bash.
- **MkDocs**: Not part of the Windows verification scope; docs build runs on Ubuntu.

## CI Automation

### Existing CI (Linux only)

| Workflow | What it runs | Platform |
|---|---|---|
| `ci-cli.yml` | `go test`, `go vet`, `go build` on CLI package changes | ubuntu-latest |
| `ci-diffviewer.yml` | `go test`, `go vet`, `go build` on diffviewer changes | ubuntu-latest |
| `ci-integration.yml` | Integration test scripts | ubuntu-latest |
| `test.yml` | Unit tests (race), smoke test, cross-compile build matrix | ubuntu-latest |
| `docs.yml` | `mkdocs build --strict` | ubuntu-latest |
| `lint.yml` | `gofmt`, `go vet`, staticcheck, drift scan, yamllint | ubuntu-latest |

### Cross-Platform CI Workflow

A dedicated cross-platform workflow (`.github/workflows/cross-platform.yml`) has been added to run tests and builds on all three platforms. See that file for details.

## Release Checklist

Before tagging a release, verify:

- [ ] **Linux (CI)**: All CI workflows pass on ubuntu-latest
- [ ] **macOS (CI/manual)**: `go test ./packages/cli/...` and build pass on `macos-latest`; manual darwin/arm64 smoke can be run before release
- [ ] **Cross-compile (CI)**: `make cli-snapshot` succeeds for all 5 platform targets
- [ ] **Windows (CI)**: Go tests/build pass on `windows-latest`
- [ ] **Documentation**: `mkdocs build --strict` passes (checked on Ubuntu)
- [ ] **Release assets**: Release workflow produces binaries for all 5 platform targets

## Follow-Up Issues

| Issue | Description | Priority |
|---|---|---|
| — | Add Windows-compatible smoke test (PowerShell or Go-based) | P3 |
| — | Add macOS disposable-project smoke test once the smoke harness is portable | P3 |
