package theme

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// TestCSSTokenParity verifies bidirectional parity between the design-system
// CSS at .claude/skills/tui-lazy-ai-design-system/colors_and_type.css and the
// exported color constants in this package. Both sides MUST agree on the token
// set and the hex value for each token.
//
// Drift in either direction is a defect:
//
//   - CSS → Go missing: a `--lz-foo: #abcdef;` declaration that has no
//     exported `Foo` constant in `internal/theme/`. Failure mode: TUI cannot
//     render that token.
//   - Go → CSS missing: an exported `lipgloss.Color("#abcdef")` constant in
//     `internal/theme/theme.go` that has no matching `--lz-*` declaration.
//     Failure mode: a new accent was added without updating the design system
//     (FR-017, ADR-001).
//
// This test STRUCTURALLY enforces FR-017 ("no new accents without
// co-change"). When you add a token on either side, this test fails until the
// other side catches up.
func TestCSSTokenParity(t *testing.T) {
	cssTokens := parseCSS(t)
	goTokens := parseGo(t)

	// CSS → Go direction.
	for cssName, cssHex := range cssTokens {
		goName := cssNameToGo(cssName)
		goHex, ok := goTokens[goName]
		if !ok {
			t.Errorf("CSS token --lz-%s (= %s) has no matching Go constant theme.%s — add it to theme.go or remove it from colors_and_type.css", cssName, cssHex, goName)
			continue
		}
		if !equalIgnoreCase(goHex, cssHex) {
			t.Errorf("hex mismatch: --lz-%s = %s, theme.%s = %s", cssName, cssHex, goName, goHex)
		}
	}

	// Go → CSS direction.
	for goName, goHex := range goTokens {
		cssName := goNameToCSS(goName)
		cssHex, ok := cssTokens[cssName]
		if !ok {
			t.Errorf("Go constant theme.%s (= %s) has no matching CSS token --lz-%s — add it to colors_and_type.css or remove it from theme.go (FR-017: no new accents without co-change)", goName, goHex, cssName)
			continue
		}
		if !equalIgnoreCase(goHex, cssHex) {
			// Already reported above when iterating CSS direction; suppress duplicate.
			_ = cssHex
		}
	}
}

// designSystemCSSPath returns the absolute path to colors_and_type.css.
func designSystemCSSPath(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	// file = .../packages/cli/internal/theme/tokens_csstest_test.go
	// Walk up to the repo root: theme -> internal -> cli -> packages -> root
	root := filepath.Join(filepath.Dir(file), "..", "..", "..", "..")
	cssPath := filepath.Join(root, ".claude", "skills", "tui-lazy-ai-design-system", "colors_and_type.css")
	abs, err := filepath.Abs(cssPath)
	if err != nil {
		t.Fatalf("filepath.Abs: %v", err)
	}
	return abs
}

// parseCSS extracts every `--lz-NAME: #HEX;` declaration from the design-system
// CSS file. Returns a map of name (without the `--lz-` prefix) to hex string.
func parseCSS(t *testing.T) map[string]string {
	t.Helper()
	path := designSystemCSSPath(t)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading CSS: %v", err)
	}

	// Match lines like `  --lz-primary:    #7D56F4;` (with optional whitespace
	// and trailing comment).
	re := regexp.MustCompile(`(?m)^\s*--lz-([a-z][a-z0-9-]*)\s*:\s*(#[0-9A-Fa-f]{3,8})\s*;`)
	matches := re.FindAllStringSubmatch(string(content), -1)

	tokens := make(map[string]string, len(matches))
	for _, m := range matches {
		tokens[m[1]] = m[2]
	}
	if len(tokens) == 0 {
		t.Fatalf("no --lz-* tokens found in %s — parser regex may be broken", path)
	}
	return tokens
}

// parseGo walks the AST of theme.go and extracts every exported package-level
// var assigned to `lipgloss.Color("#…")`. Returns a map of constant name to
// hex string.
func parseGo(t *testing.T) map[string]string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	themeGo := filepath.Join(filepath.Dir(file), "theme.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, themeGo, nil, 0)
	if err != nil {
		t.Fatalf("parse theme.go: %v", err)
	}

	tokens := make(map[string]string)

	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}
		for _, spec := range gen.Specs {
			vs, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}
			for i, name := range vs.Names {
				// Only exported names.
				if !name.IsExported() {
					continue
				}
				if i >= len(vs.Values) {
					continue
				}
				// Look for `lipgloss.Color("#XXXXXX")`.
				call, ok := vs.Values[i].(*ast.CallExpr)
				if !ok {
					continue
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok || sel.Sel.Name != "Color" {
					continue
				}
				pkg, ok := sel.X.(*ast.Ident)
				if !ok || pkg.Name != "lipgloss" {
					continue
				}
				if len(call.Args) != 1 {
					continue
				}
				lit, ok := call.Args[0].(*ast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					continue
				}
				hex := strings.Trim(lit.Value, `"`)
				tokens[name.Name] = hex
			}
		}
	}

	if len(tokens) == 0 {
		t.Fatal("no exported lipgloss.Color constants found in theme.go — parser may be broken")
	}
	return tokens
}

// cssNameToGo translates a CSS token name (e.g. `bg-code`) into the Go
// constant name we expect in theme.go (e.g. `BgCode`). Special case: `bg`
// maps to `Background` (historical name).
func cssNameToGo(cssName string) string {
	if cssName == "bg" {
		return "Background"
	}
	// Generic kebab-to-CamelCase: `bg-code` → `BgCode`, `text-2` → `Text2`.
	parts := strings.Split(cssName, "-")
	var out strings.Builder
	for _, p := range parts {
		if p == "" {
			continue
		}
		out.WriteString(strings.ToUpper(p[:1]))
		out.WriteString(p[1:])
	}
	return out.String()
}

// goNameToCSS is the inverse of cssNameToGo. `Background` → `bg`,
// `BgCode` → `bg-code`, `Primary` → `primary`, `Text2` → `text-2`.
func goNameToCSS(goName string) string {
	if goName == "Background" {
		return "bg"
	}
	// CamelCase / digit boundary to kebab-case.
	var out strings.Builder
	for i, r := range goName {
		if i > 0 && (r >= 'A' && r <= 'Z' || r >= '0' && r <= '9') {
			out.WriteRune('-')
		}
		out.WriteRune(toLower(r))
	}
	return out.String()
}

func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// equalIgnoreCase compares two hex strings ignoring case (e.g. "#7D56F4" ==
// "#7d56f4").
func equalIgnoreCase(a, b string) bool {
	return strings.EqualFold(a, b)
}

// Sanity check: package-level types are reachable and `color.Color` is
// satisfied by the existing constants.
var _ color.Color = Primary
var _ = fmt.Sprintf // silence unused-import warning
