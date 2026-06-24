package compiler_test

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestCompilerGolden runs the adapter Install and CompileMCP for projects
// defined in testdata/projects and compares the generated files against testdata/golden.
func TestCompilerGolden(t *testing.T) {
	projectsDir := filepath.Join("..", "..", "testdata", "projects")
	goldenDir := filepath.Join("..", "..", "testdata", "golden")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			t.Skip("testdata/projects does not exist")
		}
		t.Fatalf("ReadDir: %v", err)
	}

	libDir, err := library.FindLibraryDir()
	if err != nil {
		t.Fatalf("FindLibraryDir: %v", err)
	}
	libFS := library.GetLibraryFS()

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		projName := entry.Name()
		t.Run(projName, func(t *testing.T) {
			tmpDir := t.TempDir()
			srcProj := filepath.Join(projectsDir, projName)

			// Copy project source
			err := filepath.WalkDir(srcProj, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				rel, _ := filepath.Rel(srcProj, path)
				if rel == "." {
					return nil
				}
				dest := filepath.Join(tmpDir, rel)
				if d.IsDir() {
					return os.MkdirAll(dest, 0o755)
				}
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				return os.WriteFile(dest, data, 0o644)
			})
			if err != nil {
				t.Fatalf("copying project: %v", err)
			}

			// Read manifest
			mf, err := aimanifest.Load(filepath.Join(tmpDir, ".ai"))
			if err != nil && !os.IsNotExist(err) {
				t.Logf("manifest load error: %v", err)
			}
			var targets []types.ToolId
			if err == nil {
				targets, err = mf.EnabledTargets()
				if err != nil {
					t.Logf("manifest target error: %v", err)
				}
			}

			// We only want to test targets successfully resolved. If an error occurred,
			// it's an invalid manifest case (e.g. codex). We record it in a mock log or
			// we just let it be empty so no files are generated.
			if len(targets) > 0 {
				// Run scaffold
				scaffoldCtx := &scaffold.ScaffoldContext{
					TargetDir:        tmpDir,
					WorkspaceRoot:    tmpDir,
					PlanningRepoPath: tmpDir,
					SetupScope:       types.SetupScopeProject,
					Tools:            targets,
					LibraryDir:       libDir,
					LibraryFS:        libFS,
					Agents:           types.ALL_AGENTS[:],
					Skills:           types.ALL_SKILLS[:],
					Rules:            types.ALL_RULES[:],
				}
				_, err = scaffold.ScaffoldAll(scaffoldCtx)
				if err != nil {
					t.Fatalf("ScaffoldAll: %v", err)
				}

				// Run CompileMCP
				homeDir := t.TempDir()
				database, _ := db.Open(filepath.Join(tmpDir, ".ai-setup.db"))
				defer database.Close()
				compileCtx := adapter.CompileContext{
					TargetDir:  tmpDir,
					HomeDir:    homeDir,
					SetupScope: types.SetupScopeProject,
				}

				reg := adapter.NewRegistry()
				for _, tid := range targets {
					adapt, _ := reg.Get(tid)
					if adapt != nil {
						_, err := adapt.CompileMCP(compileCtx)
						if err != nil {
							t.Fatalf("CompileMCP %s: %v", tid, err)
						}
					}
				}
			}

			projGoldenDir := filepath.Join(goldenDir, projName)

			if os.Getenv("UPDATE_GOLDEN") == "true" {
				_ = os.MkdirAll(projGoldenDir, 0o755)
				_ = os.RemoveAll(projGoldenDir)
				_ = os.MkdirAll(projGoldenDir, 0o755)
				// Copy tmpDir to golden
				filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						return err
					}
					rel, _ := filepath.Rel(tmpDir, path)
					if rel == "." {
						return nil
					}
					// exclude db files
					if strings.Contains(rel, ".ai-setup.db") {
						return nil
					}
					dest := filepath.Join(projGoldenDir, rel)
					if d.IsDir() {
						return os.MkdirAll(dest, 0o755)
					}
					if rel == ".ai/populate-needed" || rel == filepath.Join(".ai", "populate-needed") {
						return nil
					}
					data, _ := os.ReadFile(path)
					return os.WriteFile(dest, data, 0o644)
				})
				return
			}

			if _, err := os.Stat(projGoldenDir); os.IsNotExist(err) {
				t.Logf("no golden directory found at %s", projGoldenDir)
				return
			}

			// Compare tmpDir with golden
			err = filepath.WalkDir(projGoldenDir, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}

				rel, _ := filepath.Rel(projGoldenDir, path)
				goldenData, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				if rel == ".ai/populate-needed" || rel == filepath.Join(".ai", "populate-needed") {
					return nil
				}

				gotData, err := os.ReadFile(filepath.Join(tmpDir, rel))
				if err != nil {
					if os.IsNotExist(err) {
						t.Errorf("golden file %s is missing from output", rel)
					} else {
						t.Errorf("reading generated file %s: %v", rel, err)
					}
					return nil
				}

				if !bytes.Equal(goldenData, gotData) {
					t.Errorf("file %s differs from golden", rel)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walking golden: %v", err)
			}

			// Reverse walk: detect extra files in output not present in golden.
			err = filepath.WalkDir(tmpDir, func(path string, d fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if d.IsDir() {
					return nil
				}
				rel, _ := filepath.Rel(tmpDir, path)
				if shouldIgnoreGoldenExtra(rel) {
					return nil
				}
				goldenPath := filepath.Join(projGoldenDir, rel)
				if _, statErr := os.Stat(goldenPath); os.IsNotExist(statErr) {
					t.Errorf("extra output file %s not in golden fixtures", rel)
				}
				return nil
			})
			if err != nil {
				t.Fatalf("walking output for extras: %v", err)
			}
		})
	}
}

func shouldIgnoreGoldenExtra(rel string) bool {
	// .ai-setup.db is an internal SQLite database tracking setup state; we don't
	// fixture it because its binary format and timestamps change every run.
	if rel == ".ai-setup.db" {
		return true
	}
	if rel == ".ai/populate-needed" || rel == filepath.Join(".ai", "populate-needed") {
		return true
	}
	if rel == ".ai/lazyai.json" || rel == filepath.Join(".ai", "lazyai.json") {
		return true
	}
	if rel == ".ai/mcp.json" || rel == filepath.Join(".ai", "mcp.json") {
		return true
	}
	return strings.HasPrefix(rel, ".specify"+string(filepath.Separator)) || strings.HasPrefix(rel, ".specify/")
}
