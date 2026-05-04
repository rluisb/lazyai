import fs from 'node:fs';
import path from 'node:path';

const root = process.cwd();
const docsDir = path.join(root, 'docs');
const mkdocsPath = path.join(root, 'mkdocs.yml');
const requirementsPath = path.join(root, 'docs', 'requirements.txt');
const readmePath = path.join(root, 'README.md');

let exitCode = 0;
const errors = [];
const warnings = [];

function fail(message) {
  errors.push(message);
  exitCode = 1;
}

function stripFragmentAndQuery(raw) {
  return raw.split('#')[0].split('?')[0];
}

function isExternal(raw) {
  return /^(https?:|mailto:|tel:)/.test(raw);
}

function normalizeLinkTarget(raw, sourceRelPath) {
  const stripped = stripFragmentAndQuery(raw.trim());
  if (!stripped || isExternal(stripped) || stripped.startsWith('#')) return null;
  const sourceDir = path.posix.dirname(sourceRelPath);
  return stripped.startsWith('/')
    ? stripped.slice(1)
    : path.posix.normalize(path.posix.join(sourceDir, stripped));
}

function existsFileOrDirectory(baseDir, relTarget) {
  const full = path.join(baseDir, relTarget);
  if (fs.existsSync(full)) return true;
  if (!path.extname(relTarget) && fs.existsSync(full + '.md')) return true;
  if (fs.existsSync(path.join(full, 'index.md'))) return true;
  return false;
}

function collectMarkdownLinks(content) {
  const links = [];
  const markdownLinkPattern = /\[[^\]]*\]\(([^)]+)\)/g;
  const referenceLinkPattern = /^\[[^\]]+\]:\s*(\S+)/gm;
  let match;
  while ((match = markdownLinkPattern.exec(content)) !== null) links.push(match[1]);
  while ((match = referenceLinkPattern.exec(content)) !== null) links.push(match[1]);
  return links;
}

if (!fs.existsSync(requirementsPath)) {
  fail('Missing docs/requirements.txt');
} else {
  const req = fs.readFileSync(requirementsPath, 'utf8');
  if (!req.includes('mkdocs-material')) fail('docs/requirements.txt must include mkdocs-material');
}

if (!fs.existsSync(mkdocsPath)) {
  fail('Missing mkdocs.yml');
} else {
  const mkdocs = fs.readFileSync(mkdocsPath, 'utf8');
  const navFiles = [...mkdocs.matchAll(/^\s*-\s+[^:\n]+:[ \t]+([^#\n]+\.md)\s*$/gm)]
    .map((match) => match[1].trim().replace(/^['"]|['"]$/g, ''));

  if (!navFiles.includes('index.md')) fail('mkdocs.yml nav should include index.md');

  for (const rel of navFiles) {
    if (!fs.existsSync(path.join(docsDir, rel))) fail(`Nav references missing file: docs/${rel}`);
  }

  for (const rel of navFiles) {
    const full = path.join(docsDir, rel);
    if (!fs.existsSync(full)) continue;
    const content = fs.readFileSync(full, 'utf8');
    for (const raw of collectMarkdownLinks(content)) {
      const target = normalizeLinkTarget(raw, rel);
      if (target && !existsFileOrDirectory(docsDir, target)) {
        fail(`Broken docs link in docs/${rel}: ${raw}`);
      }
    }
  }

  const existingNavSet = new Set(navFiles);
  const expectedCorePages = [
    'getting-started/quick-start.md',
    'getting-started/installation.md',
    'concepts/how-it-works.md',
    'concepts/scopes.md',
    'concepts/presets.md',
    'concepts/tools.md',
    'cli/index.md',
    'cli/reference.md',
    'integration/mcp.md',
    'integration/orchestration.md',
    'development/contributing.md',
    'development/release.md',
    'troubleshooting/faq.md',
  ];
  for (const rel of expectedCorePages) {
    if (!existingNavSet.has(rel)) warnings.push(`Core page exists but is not in mkdocs nav: docs/${rel}`);
  }
}

if (!fs.existsSync(readmePath)) {
  fail('Missing README.md');
} else {
  const readme = fs.readFileSync(readmePath, 'utf8');
  for (const raw of collectMarkdownLinks(readme)) {
    const stripped = stripFragmentAndQuery(raw.trim());
    if (!stripped || isExternal(stripped) || stripped.startsWith('#')) continue;
    if (stripped.startsWith('docs/')) {
      if (!existsFileOrDirectory(root, stripped)) fail(`README.md broken link: ${raw}`);
    } else if (stripped === 'LICENSE' && !fs.existsSync(path.join(root, stripped))) {
      warnings.push('README.md links to LICENSE, but LICENSE was not found');
    }
  }
}

if (errors.length) {
  console.error('ERRORS:');
  for (const error of errors) console.error(`  - ${error}`);
}
if (warnings.length) {
  console.warn('WARNINGS:');
  for (const warning of warnings) console.warn(`  - ${warning}`);
}
if (!errors.length && !warnings.length) {
  console.log('All documentation checks passed.');
} else if (!errors.length) {
  console.log('Documentation checks passed with warnings.');
} else {
  console.error('Documentation checks failed.');
}

process.exit(exitCode);
