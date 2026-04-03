# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## AI Assistant Configuration

This project uses GitHub Copilot with ai-setup integration.

{{#include fragments/context-engineering.xml}}

{{#include fragments/rpi-workflow.xml}}

{{#include fragments/chain-of-thought.xml}}

{{#if features.tree_of_thoughts}}
{{#include fragments/tree-of-thoughts.xml}}
{{/if}}

{{#include fragments/adr-enforcement.xml}}

{{#include fragments/quality-gates.xml}}

{{#if features.agent_harness}}
{{#include fragments/agent-harness.xml}}
{{/if}}

## Copilot-Specific Notes

- Agents are in `.github/agents/*.md`
- Prompts are in `.github/prompts/*.prompt.md`

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}
