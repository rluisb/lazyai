---
status: accepted
---

# ADR-010: LazyAI owns files, regions and keys, tracked by one committed lock with no journal

**Date:** 2026-10-13. Decided in [Ownership and uninstall model](https://github.com/rluisb/lazyai/issues/641) on the [LazyAI 2.0 map](https://github.com/rluisb/lazyai/issues/632).

LazyAI 2.0 owns three kinds of content: whole files it creates, marked regions inside shared Markdown, and individual keys inside shared config files. Everything else in the project belongs to the user. One lock, committed to git at `.lazyai/io.github.rluisb.lazyai/lock.json`, records each owned output with a hash of only the owned content. Content counts as "unchanged" when it still hashes to the recorded value. Compile refuses to overwrite drifted content unless given `--force`. Uninstall removes only unchanged owned content, and eject keeps everything and deletes the lock.

We rejected the journal and crash-recovery design that agent-context-broker's ADR-013 proposes. Writing each file atomically and writing the lock last leaves a clean state after any interruption, and the project's git history already provides undo. A journal would add rollback machinery that duplicates git without protecting against anything git doesn't.

## Consequences

- This replaces `.ai/lock.json` (ADR-006) and the `.ai-setup` store, which 1.x keeps as two ledgers that disagree.
- In-file markers appear only around regions (`<!-- lazyai:start <id> -->` … `<!-- lazyai:end <id> -->`). The `vibe-lab:managed` whole-file marker is dropped.
- Keys in JSON and TOML carry no marker, so deleting the lock loses ownership of them. Eject depends on exactly that.
- Migration from 1.x has to decide how existing outputs are adopted into the lock.
