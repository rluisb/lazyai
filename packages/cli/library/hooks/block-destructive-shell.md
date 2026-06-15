# block-destructive-shell Policy

## Purpose

Prevent accidental execution of destructive shell commands that could cause irreversible damage to the project, filesystem, or system.

## Denied Commands

The following command patterns are denied:

- `rm -rf /` — recursive root delete
- `rm -rf /*` — recursive root delete (glob)
- `mkfs` — filesystem format
- `dd if=/dev/zero of=` — disk zeroing
- `dd if=/dev/zero of=/dev/` — device zeroing
- `> /dev/sd` — direct device write
- `shutdown`, `poweroff`, `reboot` — system power commands
- `halt` — system halt

## Allow Policy

Commands NOT matching the denied patterns are allowed through without comment.

## Fail-Closed Semantics

If the hook adapter cannot be loaded or fails to execute:

- **Claude Code**: The hook must return exit code 2 (block) or structured deny output.
- **OpenCode**: The plugin must throw an error to block the command.

If the adapter fails silently, the command proceeds. This is a known limitation — document it in project operations notes when this policy is adopted.

## Implementation Notes

- Each CLI adapter implements this policy natively in its own mechanism.
- The deny list uses exact string/prefix matching, not regex, to avoid false negatives from shell escaping edge cases.
- This file is documentation, not a runtime contract. The generated adapters are the runtime artifacts.
