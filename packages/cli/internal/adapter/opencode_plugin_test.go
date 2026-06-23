package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeOpenCodePackageJSON_NoFileNoop(t *testing.T) {
	if err := normalizeOpenCodePackageJSON(t.TempDir()); err != nil {
		t.Fatalf("normalizeOpenCodePackageJSON() error = %v", err)
	}
}

func TestNormalizeOpenCodePackageJSON_AddsTypeModule(t *testing.T) {
	ocDir := t.TempDir()
	packagePath := filepath.Join(ocDir, "package.json")
	if err := os.WriteFile(packagePath, []byte(`{
  "dependencies": {
    "@opencode-ai/plugin": "1.17.7"
  }
}`), 0o644); err != nil {
		t.Fatalf("seed package.json: %v", err)
	}
	if err := normalizeOpenCodePackageJSON(ocDir); err != nil {
		t.Fatalf("normalizeOpenCodePackageJSON() error = %v", err)
	}
	data, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("read package.json: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, `"type": "module"`) {
		t.Fatalf("package.json missing type=module:\n%s", contents)
	}
	if !strings.Contains(contents, `"@opencode-ai/plugin": "1.17.7"`) {
		t.Fatalf("package.json lost dependency:\n%s", contents)
	}
}
