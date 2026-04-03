# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## AI Assistant Configuration

This project uses Gemini CLI with ai-setup integration.

{{#include fragments/context-engineering.xml}}

{{#include fragments/rpi-workflow.xml}}

{{#include fragments/chain-of-thought.xml}}

{{#if features.tree_of_thoughts}}
{{#include fragments/tree-of-thoughts.xml}}
{{/if}}

{{#include fragments/adr-enforcement.xml}}

{{#include fragments/quality-gates.xml}}

{{#if features.gitConventions}}
{{#include fragments/git-conventions.xml}}
{{/if}}

{{#if features.agent_harness}}
{{#include fragments/agent-harness.xml}}
{{/if}}

{{#if features.pivotHandling}}
{{#include fragments/pivot-handling.xml}}
{{/if}}

{{#if features.bugResolution}}
{{#include fragments/bug-resolution.xml}}
{{/if}}

## Gemini-Specific Notes

- Gemini does not have a separate agents concept
- Skills are in `.gemini/skills/*/SKILL.md` and function as pseudo-agents

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}
