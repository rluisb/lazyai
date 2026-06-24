package scaffold

// deprecatedRootTemplateByFile maps output filenames to their template sources.
// Kept for backward compatibility. Ported from src/scaffold/root-files.ts.
var deprecatedRootTemplateByFile = map[string]string{
	"AGENTS.md":                       "root/AGENTS.template.md",
	".github/copilot-instructions.md": "root/copilot-instructions.template.md",
}
