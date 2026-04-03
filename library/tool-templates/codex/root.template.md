# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## AI Assistant Configuration

This project uses OpenAI Codex CLI with ai-setup integration.

{{#if features.contextEngineering}}
{{#include fragments/context-engineering.xml}}
{{/if}}

{{#if features.rpiWorkflow}}
{{#include fragments/rpi-workflow.xml}}
{{/if}}

{{#if features.chainOfThought}}
{{#include fragments/chain-of-thought.xml}}
{{/if}}

{{#if features.treeOfThoughts}}
{{#include fragments/tree-of-thoughts.xml}}
{{/if}}

{{#if features.adrEnforcement}}
{{#include fragments/adr-enforcement.xml}}
{{/if}}

{{#if features.qualityGates}}
{{#include fragments/quality-gates.xml}}
{{/if}}

{{#if features.gitConventions}}
{{#include fragments/git-conventions.xml}}
{{/if}}

{{#if features.agentHarness}}
{{#include fragments/agent-harness.xml}}
{{/if}}

{{#if features.pivotHandling}}
{{#include fragments/pivot-handling.xml}}
{{/if}}

{{#if features.bugResolution}}
{{#include fragments/bug-resolution.xml}}
{{/if}}

## Codex-Specific Notes

- Agents are defined inline in this file (no separate agents directory)
- Skills are in `.codex/skills/*/SKILL.md`

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}
