## Summary

Make `lazyai-cli` installable through Homebrew so users can install and upgrade without needing to know Go module paths.

Recommended install contract:

```bash
brew tap rluisb/tap
brew install lazyai-cli
```

or one-shot:

```bash
brew install rluisb/tap/lazyai-cli
```

## Context from current implementation inspection

Current LazyAI implementation state supports Homebrew packaging well:

- `lazyai-cli` is a Go CLI and normal usage has no npm/npx dependency (`README.md`).
- The product binary lives at `packages/cli/cmd/lazyai-cli` (`docs/concepts/package-layout.md`).
- Release tags are currently submodule-prefixed: `packages/cli/vX.Y.Z` (`docs/development/release.md`).
- Existing release automation builds raw binaries in `.github/workflows/release-cli.yml` for:
  - `lazyai-cli-darwin-arm64`
  - `lazyai-cli-darwin-amd64`
  - `lazyai-cli-linux-amd64`
  - `lazyai-cli-linux-arm64`
  - `lazyai-cli-windows-amd64.exe`
  - `checksums.txt`
- Release workflow sets version ldflags for:
  - `github.com/rluisb/lazyai/packages/cli/cmd.Version`
  - `github.com/rluisb/lazyai/packages/cli/internal/version.Version`
- There is no current Homebrew tap/formula in the repo (`Formula/`, `homebrew/`, `.goreleaser*` absent during inspection).

## Proposal

Create a dedicated tap repo:

```text
github.com/rluisb/homebrew-tap
```

Add:

```text
Formula/lazyai-cli.rb
```

Use a source-build formula first. This is simplest and idiomatic for a tap.

```ruby
class LazyaiCli < Formula
  desc "Compile canonical AI setup assets into tool-native harness surfaces"
  homepage "https://github.com/rluisb/lazyai"
  url "https://github.com/rluisb/lazyai/archive/refs/tags/packages/cli/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_TARBALL_SHA256"
  license "MIT"

  depends_on "go" => :build

  def install
    cd "packages/cli" do
      ldflags = %W[
        -s -w
        -X github.com/rluisb/lazyai/packages/cli/cmd.Version=#{version}
        -X github.com/rluisb/lazyai/packages/cli/internal/version.Version=#{version}
      ]

      system "go", "build", *std_go_args(ldflags: ldflags), "./cmd/lazyai-cli"
    end
  end

  test do
    assert_match "lazyai-cli", shell_output("#{bin}/lazyai-cli --help")
    # If --version is stable, add:
    # assert_match version.to_s, shell_output("#{bin}/lazyai-cli --version")
  end
end
```

## Release automation

Add workflow support to update the tap when a `packages/cli/v*` tag is released.

Suggested steps:

1. On `packages/cli/vX.Y.Z` release, compute the source tarball SHA:

   ```bash
   curl -L -o source.tar.gz \
     https://github.com/rluisb/lazyai/archive/refs/tags/packages/cli/vX.Y.Z.tar.gz
   shasum -a 256 source.tar.gz
   ```

2. Update `rluisb/homebrew-tap/Formula/lazyai-cli.rb` with new version URL + SHA.
3. Open a PR or commit directly to the tap using `brew bump-formula-pr` / `Homebrew/actions/bump-formula-pr`.
4. Update LazyAI README/install docs to include:

   ```bash
   brew install rluisb/tap/lazyai-cli
   ```

## Known wrinkle

The current release tag format includes a slash:

```text
packages/cli/vX.Y.Z
```

This works for GitHub archive URLs, but it may be awkward for Homebrew livecheck. If needed, either:

1. add explicit `livecheck` regex for the prefixed tag, or
2. publish an alias tag such as `lazyai-cli-vX.Y.Z` specifically for package managers.

Potential livecheck shape:

```ruby
livecheck do
  url :stable
  regex(%r{refs/tags/packages/cli/v?(\d+(?:\.\d+)+)\.tar\.gz}i)
end
```

## Acceptance criteria

- [ ] `rluisb/homebrew-tap` exists.
- [ ] `Formula/lazyai-cli.rb` installs `lazyai-cli` from a released version.
- [ ] `brew install rluisb/tap/lazyai-cli` works on macOS arm64.
- [ ] `brew test lazyai-cli` passes.
- [ ] `brew audit --strict --online lazyai-cli` passes or any warnings are documented.
- [ ] Release automation updates the formula for new `packages/cli/v*` tags.
- [ ] README/docs mention Homebrew install alongside `go install`.

## Notes from implementation-state findings

This packaging issue does not change LazyAI's product boundary. LazyAI remains a Go CLI that owns canonical `.ai/` source management, compile/adapter output, validation/doctor/update/migration, and optional runtime-adjacent local state. Homebrew is only an installation channel for the existing `lazyai-cli` binary.
