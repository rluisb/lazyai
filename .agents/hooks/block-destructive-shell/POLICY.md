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
- **Claude**: The hook must return exit code 2 (block) or structured deny output.
- **opencode**: The plugin must throw an error to block the command.
- **Pi**: Not supported. No project-local hook mechanism verified.

If the adapter fails silently (e.g., plugin load failure in opencode), the command proceeds. This is a known limitation documented in OPERATIONS.md.

## Implementation Notes

- Each CLI adapter implements this policy natively in its own mechanism.
- The deny list uses exact string/prefix matching, not regex, to avoid false negatives from shell escaping edge cases.
- The canonical policy (this file) is documentation, not a runtime contract. The generated adapters are the runtime artifacts