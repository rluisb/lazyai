package minimality

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/rluisb/lazyai/packages/cli/internal/tokenrent"
)

const GoFileLineThreshold = 500

var canonicalAssetCategories = []string{"agents", "skills", "hooks", "prompts", "templates", "rules"}

// Report is the deterministic, report-only minimality snapshot for a repository.
type Report struct {
	GoFileLineThreshold int
	GoFiles             []GoFileSummary
	CLICommands         CLICommandSummary
	CanonicalLibrary    CanonicalLibrarySummary
	CanonicalAssets     []CanonicalAssetSummary
}

// GoFileSummary reports one Go file above the advisory line threshold.
type GoFileSummary struct {
	Path  string
	Lines int
}

// CLICommandSummary reports top-level CLI command registrations from source.
type CLICommandSummary struct {
	Visible    int
	Hidden     int
	Registered int
}

// CanonicalLibrarySummary reports canonical library token-rent size inputs.
type CanonicalLibrarySummary struct {
	Path        string
	Bytes       int
	BudgetBytes int
}

// CanonicalAssetSummary reports canonical asset counts for one category.
type CanonicalAssetSummary struct {
	Category string
	Count    int
	Present  bool
}

// AnalyzeProject reads repository files and returns a deterministic minimality report.
func AnalyzeProject(projectRoot string) (Report, error) {
	root, err := filepath.Abs(projectRoot)
	if err != nil {
		return Report{}, fmt.Errorf("resolve project root: %w", err)
	}
	root = filepath.Clean(root)

	goFiles, err := findLargeGoFiles(root, GoFileLineThreshold)
	if err != nil {
		return Report{}, err
	}

	commands, err := countTopLevelCLICommands(filepath.Join(root, "packages", "cli", "cmd"))
	if err != nil {
		return Report{}, err
	}

	tokenResult, budgetBytes, err := readCanonicalLibrary(root)
	if err != nil {
		return Report{}, err
	}

	return Report{
		GoFileLineThreshold: GoFileLineThreshold,
		GoFiles:             goFiles,
		CLICommands:         commands,
		CanonicalLibrary: CanonicalLibrarySummary{
			Path:        tokenrent.CanonicalSubdir,
			Bytes:       tokenResult.TotalBytes,
			BudgetBytes: budgetBytes,
		},
		CanonicalAssets: countCanonicalAssets(root, tokenResult.Files),
	}, nil
}

// Run writes the report and returns nil for advisory threshold and token-rent budget breaches.
func Run(projectRoot string, out io.Writer) error {
	report, err := AnalyzeProject(projectRoot)
	if err != nil {
		return err
	}
	return WriteText(out, report)
}

// WriteText writes a stable plain-text report with no colors or timestamps.
func WriteText(out io.Writer, report Report) error {
	if _, err := fmt.Fprintln(out, "LazyAI minimality report"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "Mode: report-only; advisory thresholds do not change the exit code."); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(out, "Go files over %d lines:\n", report.GoFileLineThreshold); err != nil {
		return err
	}
	if len(report.GoFiles) == 0 {
		if _, err := fmt.Fprintln(out, "  none"); err != nil {
			return err
		}
	} else {
		for _, file := range report.GoFiles {
			if _, err := fmt.Fprintf(out, "  %d %s\n", file.Lines, file.Path); err != nil {
				return err
			}
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "Top-level CLI commands:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  visible: %d\n", report.CLICommands.Visible); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  hidden: %d\n", report.CLICommands.Hidden); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  registered: %d\n", report.CLICommands.Registered); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "Canonical library:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  path: %s\n", report.CanonicalLibrary.Path); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  bytes: %d\n", report.CanonicalLibrary.Bytes); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  token-rent budget: %d\n", report.CanonicalLibrary.BudgetBytes); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "  hard gate: go run ./packages/cli/internal/tokenrent/cmd/token-rent-check"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "Canonical asset counts:"); err != nil {
		return err
	}
	for _, asset := range report.CanonicalAssets {
		state := "present"
		if !asset.Present {
			state = "directory absent"
		}
		if _, err := fmt.Fprintf(out, "  %s: %d (%s)\n", asset.Category, asset.Count, state); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out); err != nil {
		return err
	}

	if _, err := fmt.Fprintln(out, "Advisory thresholds:"); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(out, "  Go file line threshold: >%d lines\n", report.GoFileLineThreshold); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "  CLI command count: report-only; no hard maximum yet"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "  Canonical asset counts: report-only; no hard maximum yet"); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(out, "  Canonical bytes are informational here; token-rent remains the hard gate."); err != nil {
		return err
	}
	return nil
}

func readCanonicalLibrary(root string) (tokenrent.Result, int, error) {
	result, err := tokenrent.CheckProject(root)
	budgetBytes := result.BudgetBytes
	if result.Override != nil && result.Override.Budget > 0 {
		budgetBytes = result.Override.Budget
	}
	if err == nil {
		return result, budgetBytes, nil
	}

	var budgetErr *tokenrent.BudgetError
	if errors.As(err, &budgetErr) {
		if budgetErr.BudgetBytes > 0 {
			budgetBytes = budgetErr.BudgetBytes
		}
		return result, budgetBytes, nil
	}
	return result, budgetBytes, err
}

func findLargeGoFiles(root string, threshold int) ([]GoFileSummary, error) {
	var files []GoFileSummary
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipDir(d.Name()) && path != root {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		lines, err := countLines(path)
		if err != nil {
			return err
		}
		if lines <= threshold {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		files = append(files, GoFileSummary{Path: filepath.ToSlash(rel), Lines: lines})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk Go files: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Lines != files[j].Lines {
			return files[i].Lines > files[j].Lines
		}
		return files[i].Path < files[j].Path
	})
	return files, nil
}

func shouldSkipDir(name string) bool {
	if strings.HasPrefix(name, ".") {
		return true
	}
	switch name {
	case "archive", "build", "dist", "node_modules", "tmp", "vendor":
		return true
	default:
		return false
	}
}

func countLines(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	var buf [32 * 1024]byte
	lines := 0
	bytesRead := 0
	lastWasNewline := false
	for {
		n, readErr := file.Read(buf[:])
		if n > 0 {
			bytesRead += n
			for _, b := range buf[:n] {
				if b == '\n' {
					lines++
				}
			}
			lastWasNewline = buf[n-1] == '\n'
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return 0, readErr
		}
	}
	if bytesRead > 0 && !lastWasNewline {
		lines++
	}
	return lines, nil
}

func countTopLevelCLICommands(cmdDir string) (CLICommandSummary, error) {
	commandHidden := make(map[string]bool)
	var registrations []string
	fset := token.NewFileSet()

	err := filepath.WalkDir(cmdDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			return nil
		}

		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.VAR {
				continue
			}
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok {
					continue
				}
				for i, value := range valueSpec.Values {
					if i >= len(valueSpec.Names) {
						continue
					}
					hidden, ok := cobraCommandHidden(value)
					if ok {
						commandHidden[valueSpec.Names[i].Name] = hidden
					}
				}
			}
		}
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || !isRootAddCommand(call) {
				return true
			}
			for _, arg := range call.Args {
				if ident, ok := arg.(*ast.Ident); ok {
					registrations = append(registrations, ident.Name)
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		return CLICommandSummary{}, fmt.Errorf("count top-level CLI commands: %w", err)
	}

	summary := CLICommandSummary{Registered: len(registrations)}
	for _, name := range registrations {
		if commandHidden[name] {
			summary.Hidden++
		} else {
			summary.Visible++
		}
	}
	return summary, nil
}

func cobraCommandHidden(expr ast.Expr) (bool, bool) {
	if unary, ok := expr.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expr = unary.X
	}
	composite, ok := expr.(*ast.CompositeLit)
	if !ok || !isCobraCommandType(composite.Type) {
		return false, false
	}
	for _, elt := range composite.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, ok := kv.Key.(*ast.Ident)
		if !ok || key.Name != "Hidden" {
			continue
		}
		value, ok := kv.Value.(*ast.Ident)
		return ok && value.Name == "true", true
	}
	return false, true
}

func isCobraCommandType(expr ast.Expr) bool {
	selector, ok := expr.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Command" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "cobra"
}

func isRootAddCommand(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "AddCommand" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "rootCmd"
}

func countCanonicalAssets(root string, files []tokenrent.CountedFile) []CanonicalAssetSummary {
	counts := make(map[string]int, len(canonicalAssetCategories))
	for _, file := range files {
		category, _, ok := strings.Cut(file.Path, "/")
		if !ok {
			continue
		}
		if isCanonicalAssetCategory(category) {
			counts[category]++
		}
	}

	assets := make([]CanonicalAssetSummary, 0, len(canonicalAssetCategories))
	for _, category := range canonicalAssetCategories {
		_, err := os.Stat(filepath.Join(root, tokenrent.CanonicalSubdir, category))
		assets = append(assets, CanonicalAssetSummary{
			Category: category,
			Count:    counts[category],
			Present:  err == nil,
		})
	}
	return assets
}

func isCanonicalAssetCategory(category string) bool {
	for _, candidate := range canonicalAssetCategories {
		if category == candidate {
			return true
		}
	}
	return false
}
