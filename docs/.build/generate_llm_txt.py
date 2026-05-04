#!/usr/bin/env python3
"""
MkDocs gen-files script: generates llm.txt from curated pages.

This script is called by the mkdocs-gen-files plugin during each build.
It reads curated markdown pages and writes a single concatenated llm.txt
that LLMs can fetch in one request.

The resulting llm.txt is served at https://rluisb.github.io/lazyai/llm.txt
"""

import mkdocs_gen_files
from pathlib import Path
from datetime import date

# Pages to include (relative to docs/ directory).
PAGES = [
    "index.md",
    "getting-started/quick-start.md",
    "getting-started/installation.md",
    "concepts/how-it-works.md",
    "concepts/scopes.md",
    "concepts/presets.md",
    "concepts/tools.md",
    "cli/index.md",
    "cli/reference.md",
    "integration/mcp.md",
    "integration/orchestration.md",
    "troubleshooting/faq.md",
    "migration/ai-setup-to-lazyai.md",
]

HEADER = """\
# LazyAI — LLM-Optimized Documentation

> This file is auto-generated from curated documentation pages.
> Last updated: {date}
> Source: https://github.com/rluisb/lazyai
> Site: https://rluisb.github.io/lazyai/

---

"""

SEPARATOR = "\n\n---\n\n"


def main():
    # The script lives in docs/.build/, so docs/ is one level up.
    docs_dir = Path(__file__).resolve().parent.parent

    sections = []
    for page in PAGES:
        src = docs_dir / page
        if not src.exists():
            print(f"  WARNING: {page} not found, skipping")
            continue

        # Derive a readable section title from the filename
        title = src.stem.replace("-", " ").replace("_", " ").title()
        content = src.read_text(encoding="utf-8")

        # Strip MkDocs frontmatter if present
        if content.startswith("---\n"):
            end = content.find("---\n", 3)
            if end != -1:
                content = content[end + 4:]

        sections.append(f"## {title}\n\n{content}")

    body = SEPARATOR.join(sections)
    full = HEADER.format(date=date.today().isoformat()) + body + "\n"

    # Write via mkdocs-gen-files so it's included in the build output
    with mkdocs_gen_files.open("llm.txt", "w") as f:
        f.write(full)

    print(f"  Generated llm.txt ({len(full):,} bytes, {len(sections)} pages)")


main()