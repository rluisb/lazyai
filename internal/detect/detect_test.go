package detect

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectProjectMetadata(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		expected *ProjectMetadata
	}{
		{
			name:  "NodeJS Project",
			files: []string{"package.json", "pnpm-lock.yaml"},
			expected: &ProjectMetadata{
				PrimaryLanguage: "TypeScript/JavaScript",
				PackageManager:  "pnpm",
				WorkspaceType:   "monorepo",
			},
		},
		{
			name:  "Go Project",
			files: []string{"go.mod"},
			expected: &ProjectMetadata{
				PrimaryLanguage: "Go",
				WorkspaceType:   "single-module",
			},
		},
		{
			name:  "Rust Project",
			files: []string{"Cargo.toml"},
			expected: &ProjectMetadata{
				PrimaryLanguage: "Rust",
			},
		},
		{
			name:  "Ruby Rails Project",
			files: []string{"Gemfile", "config/routes.rb"},
			expected: &ProjectMetadata{
				PrimaryLanguage: "Ruby",
				Framework:       "Rails",
			},
		},
		{
			name:     "Generic Project",
			files:    []string{"README.md"},
			expected: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			for _, f := range tc.files {
				path := filepath.Join(tmpDir, f)
				_ = os.MkdirAll(filepath.Dir(path), 0755)
				_ = os.WriteFile(path, []byte("test"), 0644)
			}

			res := DetectProjectMetadata(tmpDir)
			if (res == nil) != (tc.expected == nil) {
				t.Fatalf("expected %v, got %v", tc.expected, res)
			}
			if res != nil && tc.expected != nil {
				if res.PrimaryLanguage != tc.expected.PrimaryLanguage {
					t.Errorf("expected PrimaryLanguage %s, got %s", tc.expected.PrimaryLanguage, res.PrimaryLanguage)
				}
				if res.PackageManager != tc.expected.PackageManager {
					t.Errorf("expected PackageManager %s, got %s", tc.expected.PackageManager, res.PackageManager)
				}
				if res.Framework != tc.expected.Framework {
					t.Errorf("expected Framework %s, got %s", tc.expected.Framework, res.Framework)
				}
			}
		})
	}
}
