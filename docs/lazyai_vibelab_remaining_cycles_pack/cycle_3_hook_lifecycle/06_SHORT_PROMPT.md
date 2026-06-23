# Short Prompt — Cycle 3

Work on:

```text
RPI Cycle 3 — Hook lifecycle catalog and adapter capability mapping
```

Research current hooks and adapter support. Then define a neutral lifecycle:

```text
before_agent
before_model
before_tool
after_tool
after_model
after_agent
on_error
on_compaction
on_handoff
on_human_gate
```

Map each adapter honestly:

```text
supported / partial / instruction_only / unsupported / not_applicable
```

Classify existing hooks:

```text
pre-commit
rpi-gate-check
caveman-memory-promotion
startup-self-heal
block-destructive-shell
objective-workflow-gate
```

Do not add runtime execution or orchestration. This is catalog/capability/validation work only.
