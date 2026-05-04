# Installation

## Prerequisites

- **Node.js** `>=20.12.0`
- **Go 1.26+** (only if building the binary from source)
- **pnpm** `>=9.0.0` (for monorepo development)

## Option 1: Run from GitHub with `npx` (recommended)

The shortest install path uses the GitHub shortcut form:

```bash
npx github:ricardoborges-teachable/ai-setup init
```

This downloads the latest version automatically on each invocation. No local install needed.

If you prefer the explicit package identity:

```bash
npx @ricardoborges-teachable/ai-setup@github:ricardoborges-teachable/ai-setup init
```

## Option 2: Clone + link for development

If you are working on `ai-setup` itself:

```bash
git clone git@github.com:ricardoborges-teachable/ai-setup.git
cd ai-setup
pnpm install
pnpm run build
pnpm --filter @ricardoborges-teachable/ai-setup link --global
```

After linking, the binary is available as:

```bash
ai-setup --help
```

## Option 3: Upgrade from an earlier install

If you already have `ai-setup` linked or installed:

```bash
ai-setup update-self --check   # see if a newer release exists
ai-setup update-self --dry-run # preview what would change
ai-setup update-self             # download and replace the binary
```

After upgrading the binary, refresh managed files:

```bash
ai-setup update --check
ai-setup update
ai-setup doctor
```

## Binary name

The executable name is always `ai-setup`:

```bash
ai-setup init
ai-setup compile
ai-setup doctor
ai-setup status
```

## Symlink mode

By default, `ai-setup` copies library files into tool directories. You can use symlinks instead to keep all tools in sync when the library updates:

```bash
ai-setup init --install-mode symlink --no-interactive
```

Symlinked files are tracked in `.ai-setup.json` with `kind: "symlink"`.

## Next steps

- [Quick Start](quick-start.md)
- [How It Works](../concepts/how-it-works.md)
- [CLI Reference](../cli/reference.md)
