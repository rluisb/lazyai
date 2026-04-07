import { TemplateCompiler } from "../src/compiler/template-compiler.js";
import type { FragmentContext } from "../src/compiler/fragment-resolver.js";
import type { ToolId } from "../src/types.js";
import path from "node:path";
import { fileURLToPath } from "node:url";
import os from "node:os";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(__dirname, "..");
const LIBRARY = path.join(ROOT, "library");
const TMP_OUTPUT = path.join(os.tmpdir(), "ai-setup-test-output");

// Correct context format per FragmentResolver interface
const context: FragmentContext = {
  projectName: "test-project",
  planningDir: ".planning",
  primaryLanguage: "TypeScript",
  framework: "React",
  workspaceType: "project",
  features: {
    tree_of_thoughts: false,
    agent_harness: false,
    bug_resolution: true
  }
};

console.log("🧪 Testing Multi-Tool Compilation\n");
console.log(`Library: ${LIBRARY}`);
console.log(`Output: ${TMP_OUTPUT}\n`);

// Test each tool
const tools: ToolId[] = ["claude-code", "opencode", "codex", "copilot", "gemini"];

for (const tool of tools) {
  try {
    const compiler = new TemplateCompiler({
      libraryDir: LIBRARY,
      outputDir: TMP_OUTPUT,
      tool,
      context
    });
    
    const result = compiler.compile();
    const firstFile = result.files[0];
    
    if (firstFile) {
      const content = firstFile.content;
      const hasSystemContext = content.includes("<system-context>");
      const hasRpi = content.includes("<rpi-workflow>");
      const hasReasoning = content.includes("<reasoning-protocol>");
      const hasDecision = content.includes("<decision-protocol>");
      const hasPlanningDir = content.includes(".planning");
      const hasProjectName = content.includes("test-project");
      
      console.log(`✅ ${tool}:`);
      console.log(`   Files: ${result.files.map(f => f.relativePath).join(", ")}`);
      console.log(`   Length: ${content.length} chars`);
      console.log(`   Fragment markers: ✓ system-context, ${hasRpi ? '✓' : '✗'} rpi, ${hasReasoning ? '✓' : '✗'} reasoning, ${hasDecision ? '✓' : '✗'} decision`);
      console.log(`   Variables: planningDir=${hasPlanningDir}, projectName=${hasProjectName}`);
    } else {
      console.log(`⚠️ ${tool}: No files generated`);
    }
  } catch (e) {
    console.log(`❌ ${tool}: ${e}`);
  }
}

// Show sample output
console.log("\n📄 Sample output (Claude Code root.md, first 1500 chars):");
console.log("═".repeat(60));
const sampleCompiler = new TemplateCompiler({
  libraryDir: LIBRARY,
  outputDir: TMP_OUTPUT,
  tool: "claude-code",
  context
});
const sample = sampleCompiler.compile();
const rootFile = sample.files.find(f => f.relativePath === "root.md");
if (rootFile) {
  console.log(rootFile.content.substring(0, 1500));
  console.log("═".repeat(60));
  console.log(`\n✅ TOTAL: ${rootFile.content.length} chars compiled successfully`);
}
