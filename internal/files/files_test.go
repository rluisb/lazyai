package files

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWriteFile_CreatesParentDirs(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "a", "b", "c", "test.txt")

	if err := WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if !FileExists(path) {
		t.Error("file not created")
	}
}

func TestReadFile_ReadsBackCorrectly(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello world")

	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	got, err := ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("ReadFile = %q, want %q", got, content)
	}
}

func TestReadFile_NotFound(t *testing.T) {
	t.Parallel()

	_, err := ReadFile("/nonexistent/file.txt")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestFileHash_Consistent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "hashme.txt")
	os.WriteFile(path, []byte("hash this content"), 0o644)

	h1, err := FileHash(path)
	if err != nil {
		t.Fatalf("FileHash first call: %v", err)
	}
	h2, err := FileHash(path)
	if err != nil {
		t.Fatalf("FileHash second call: %v", err)
	}

	if h1 != h2 {
		t.Errorf("hashes differ: %q vs %q", h1, h2)
	}
	if len(h1) != 16 {
		t.Errorf("hash length = %d, want 16", len(h1))
	}
}

func TestFileHash_DifferentContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	os.WriteFile(p1, []byte("content A"), 0o644)
	os.WriteFile(p2, []byte("content B"), 0o644)

	h1, _ := FileHash(p1)
	h2, _ := FileHash(p2)

	if h1 == h2 {
		t.Error("different content should produce different hashes")
	}
}

func TestEnsureDir(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "dir")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	if !DirExists(path) {
		t.Error("directory not created")
	}
}

func TestEnsureDir_Idempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "existing")

	if err := EnsureDir(path); err != nil {
		t.Fatalf("first EnsureDir: %v", err)
	}
	if err := EnsureDir(path); err != nil {
		t.Fatalf("second EnsureDir: %v", err)
	}
}

func TestFileExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	existing := filepath.Join(dir, "exists.txt")
	os.WriteFile(existing, []byte("data"), 0o644)

	if !FileExists(existing) {
		t.Error("FileExists returned false for existing file")
	}
	if FileExists(filepath.Join(dir, "nope.txt")) {
		t.Error("FileExists returned true for nonexistent file")
	}
}

func TestDirExists(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	sub := filepath.Join(dir, "subdir")
	os.MkdirAll(sub, 0o755)

	if !DirExists(sub) {
		t.Error("DirExists returned false for existing directory")
	}
	if DirExists(filepath.Join(dir, "nope")) {
		t.Error("DirExists returned true for nonexistent directory")
	}
}

func TestCopyFile_PreservesContent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "out", "dst.txt")
	content := []byte("copy me exactly")

	os.WriteFile(src, content, 0o644)

	if err := CopyFile(src, dst); err != nil {
		t.Fatalf("CopyFile: %v", err)
	}

	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("dst content = %q, want %q", got, content)
	}
}
