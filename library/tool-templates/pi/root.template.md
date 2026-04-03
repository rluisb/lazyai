# {{PROJECT_NAME}}

{{#include fragments/system-context.xml}}

## AI Assistant Configuration

This project uses Pi Coding Agent with ai-setup integration.

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

## Pi-Specific Notes

- Agents are in `.pi/agents/*.md`
- Templates are in `.pi/templates/*.md`
- Skills are in `.pi/skills/*.md`

## Project-Specific Instructions

{{PROJECT_INSTRUCTIONS}}
