# Homebrew packaging for lazyai-cli

## Tap strategy

`lazyai-cli` is distributed through a dedicated Homebrew tap:

- **Tap repository:** `rluisb/homebrew-lazyai` (https://github.com/rluisb/homebrew-lazyai)
- **Formula name:** `lazyai-cli`
- **Install command:** `brew install rluisb/lazyai/lazyai-cli`

A dedicated tap is the standard Homebrew convention for projects that are not part of `homebrew/core`. It keeps the formula under the project maintainer's control and avoids the lengthy review process required for `homebrew/core` acceptance.

## Formula source

The canonical formula template lives at `packaging/homebrew/lazyai-cli.rb.tmpl` in this repository. It is a Go template with `{{VERSION}}`, `{{SHA256_DARWIN_ARM64}}`, and `{{SHA256_DARWIN_AMD64}}` placeholders.

## Release artifact flow

1. A `packages/cli/vX.Y.Z` tag is pushed.
2. The `release-cli.yml` workflow cross-compiles binaries and uploads them as GitHub Release assets.
3. A maintainer runs `make homebrew-formula VERSION=vX.Y.Z` (or the equivalent manual steps) to render the formula template with the real checksums from the release.
4. The rendered formula is committed to the `rluisb/homebrew-lazyai` tap repository and pushed.

## Release maintenance

When cutting a new release:

1. Tag and publish the Go release: `git tag packages/cli/vX.Y.Z && git push origin packages/cli/vX.Y.Z`
2. Wait for the release workflow to finish and upload assets.
3. Download `checksums.txt` from the release page, or extract the SHA-256 values for the two Darwin binaries.
4. Render the formula template:

   ```bash
   VERSION="X.Y.Z"
   SHA256_DARWIN_ARM64="..."
   SHA256_DARWIN_AMD64="..."
   sed "s/{{VERSION}}/$VERSION/g; s/{{SHA256_DARWIN_ARM64}}/$SHA256_DARWIN_ARM64/g; s/{{SHA256_DARWIN_AMD64}}/$SHA256_DARWIN_AMD64/g" \
     packaging/homebrew/lazyai-cli.rb.tmpl > lazyai-cli.rb
   ```

5. Commit `lazyai-cli.rb` to the `rluisb/homebrew-lazyai` tap repository and push.
6. Verify: `brew update && brew install rluisb/lazyai/lazyai-cli && lazyai-cli --version`

## Manual setup (one-time)

The tap repository must be created once:

```bash
# Create a new repository on GitHub named "homebrew-lazyai"
# Clone it locally
git clone https://github.com/rluisb/homebrew-lazyai.git
cd homebrew-lazyai
# Add the rendered formula
# Commit and push
```

After the tap exists, users install with:

```bash
brew tap rluisb/lazyai
brew install lazyai-cli
```

Or in a single step:

```bash
brew install rluisb/lazyai/lazyai-cli
```

## Linux

Only macOS (arm64 and amd64) is supported for Homebrew installation. Linuxbrew is not claimed or verified. Linux users should use `go install` or the raw release binaries.
