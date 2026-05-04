# LazyAI Wiki Bootstrap Source

This directory contains short-form GitHub Wiki pages for LazyAI. The canonical documentation remains GitHub Pages at <https://rluisb.github.io/lazyai/>; these files are a manually bootstrapped mirror for quick discovery in the repository Wiki.

## Pages included

Copy these files to the root of the Wiki repository:

- `Home.md`
- `Installation.md`
- `Package-Layout.md`
- `Go-Install-Versions.md`
- `Migration-ai-setup-to-LazyAI.md`
- `Troubleshooting.md`
- `Release-Process.md`

Do not copy this `README.md` unless you intentionally want a process page in the Wiki.

## Manual bootstrap / sync process

Use a temporary checkout and review the diff before pushing:

```bash
git clone https://github.com/rluisb/lazyai.wiki.git /tmp/lazyai.wiki
cp docs/wiki/Home.md /tmp/lazyai.wiki/Home.md
cp docs/wiki/Installation.md /tmp/lazyai.wiki/Installation.md
cp docs/wiki/Package-Layout.md /tmp/lazyai.wiki/Package-Layout.md
cp docs/wiki/Go-Install-Versions.md /tmp/lazyai.wiki/Go-Install-Versions.md
cp docs/wiki/Migration-ai-setup-to-LazyAI.md /tmp/lazyai.wiki/Migration-ai-setup-to-LazyAI.md
cp docs/wiki/Troubleshooting.md /tmp/lazyai.wiki/Troubleshooting.md
cp docs/wiki/Release-Process.md /tmp/lazyai.wiki/Release-Process.md
git -C /tmp/lazyai.wiki status --short
git -C /tmp/lazyai.wiki diff --stat
git -C /tmp/lazyai.wiki diff
```

If the diff is correct and a human has approved the Wiki update:

```bash
git -C /tmp/lazyai.wiki add Home.md Installation.md Package-Layout.md Go-Install-Versions.md Migration-ai-setup-to-LazyAI.md Troubleshooting.md Release-Process.md
git -C /tmp/lazyai.wiki commit -m "Bootstrap LazyAI wiki"
git -C /tmp/lazyai.wiki push
```

## Automation status

PAT-backed automated Wiki sync is intentionally deferred. A future workflow would need separate approval, a least-privilege secret with Wiki write access, and review of the security/branch-protection implications. Until then, keep this process manual.
