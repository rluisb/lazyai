package files

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"
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

func TestCreateTimestampedBackup_CopiesFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	src := filepath.Join(dir, "config.json")
	content := []byte(`{"managed":true}`)
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatalf("seed src: %v", err)
	}

	backupPath, err := CreateTimestampedBackup(src)
	if err != nil {
		t.Fatalf("CreateTimestampedBackup: %v", err)
	}
	if matched := regexp.MustCompile(`config\.json\.\d{8}T\d{6}Z(\.\d+)?\.bak$`).MatchString(backupPath); !matched {
		t.Fatalf("backup path = %q, want timestamped .bak suffix", backupPath)
	}

	got, err := os.ReadFile(backupPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("backup content = %q, want %q", got, content)
	}
}

func TestCreateTimestampedBackup_CopiesDirectory(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	srcDir := filepath.Join(dir, "settings")
	if err := os.MkdirAll(srcDir, 0o755); err != nil {
		t.Fatalf("mkdir srcDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "settings.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatalf("seed directory: %v", err)
	}

	backupPath, err := CreateTimestampedBackup(srcDir)
	if err != nil {
		t.Fatalf("CreateTimestampedBackup: %v", err)
	}
	if !DirExists(backupPath) {
		t.Fatalf("expected backup directory %q", backupPath)
	}
	if !FileExists(filepath.Join(backupPath, "settings.json")) {
		t.Fatalf("expected copied file inside %q", backupPath)
	}
}

func TestAtomicWriteFile_CreatesFileWithoutBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "no-backup", "config.yaml")
	data := []byte("first")

	backup, err := AtomicWriteFile(path, data, 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile: %v", err)
	}
	if backup != "" {
		t.Fatalf("backup = %q, want empty", backup)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if string(got) != string(data) {
		t.Errorf("written content = %q, want %q", got, data)
	}
}

func TestAtomicWriteFile_ReplacesTargetAndWritesSingleSlotBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "settings.yaml")
	seed := []byte("old")
	if err := os.WriteFile(path, seed, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	backup, err := AtomicWriteFile(path, []byte("new"), 0o644)
	if err != nil {
		t.Fatalf("AtomicWriteFile: %v", err)
	}
	if backup == "" {
		t.Fatalf("expected backup path, got empty")
	}
	if backup != path+".bak" {
		t.Fatalf("backup path = %q, want %q", backup, path+".bak")
	}

	got, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(got) != string(seed) {
		t.Fatalf("backup content = %q, want %q", got, seed)
	}

	got, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(got) != "new" {
		t.Fatalf("target content = %q, want new", got)
	}
}

func TestAtomicWriteFile_OverwritesSingleSlotBackup(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	seed := []byte("old")
	if err := os.WriteFile(path, seed, 0o644); err != nil {
		t.Fatalf("seed file: %v", err)
	}

	if _, err := AtomicWriteFile(path, []byte("first"), 0o644); err != nil {
		t.Fatalf("first AtomicWriteFile: %v", err)
	}
	backup, err := AtomicWriteFile(path, []byte("second"), 0o644)
	if err != nil {
		t.Fatalf("second AtomicWriteFile: %v", err)
	}
	if backup == "" {
		t.Fatalf("expected backup path, got empty")
	}

	backupBytes, err := os.ReadFile(backup)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(backupBytes) != "first" {
		t.Fatalf("backup content = %q, want %q", backupBytes, "first")
	}

	targetBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read target: %v", err)
	}
	if string(targetBytes) != "second" {
		t.Fatalf("target content = %q, want %q", targetBytes, "second")
	}
}

func TestWithFileLock_SerializesConcurrentSections(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lockPath := filepath.Join(dir, "workspace.lock")
	firstAcquired := make(chan time.Time, 1)
	secondAcquired := make(chan time.Time, 1)
	releaseFirst := make(chan struct{})
	errors := make(chan error, 2)

	go func() {
		errors <- WithFileLock(lockPath, 2*time.Second, time.Minute, func() error {
			firstAcquired <- time.Now()
			<-releaseFirst
			return nil
		})
	}()

	<-firstAcquired

	go func() {
		errors <- WithFileLock(lockPath, 2*time.Second, time.Minute, func() error {
			secondAcquired <- time.Now()
			return nil
		})
	}()

	select {
	case <-secondAcquired:
		t.Fatal("second section started before first was released")
	default:
	}

	releasedAt := time.Now()
	close(releaseFirst)
	second := <-secondAcquired
	if second.Before(releasedAt) {
		t.Fatalf("second section started too early: second=%s releasedAt=%s", second, releasedAt)
	}

	for i := 0; i < 2; i++ {
		if err := <-errors; err != nil {
			t.Fatalf("WithFileLock: %v", err)
		}
	}
}

func TestWithFileLock_TimesOutWhenHeld(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lockPath := filepath.Join(dir, "held.lock")
	if err := os.WriteFile(lockPath, []byte("held"), 0o600); err != nil {
		t.Fatalf("seed lock file: %v", err)
	}
	if err := os.Chtimes(lockPath, time.Now(), time.Now()); err != nil {
		t.Fatalf("touch lock: %v", err)
	}

	err := WithFileLock(lockPath, 100*time.Millisecond, time.Minute, func() error {
		return nil
	})
	if err == nil {
		t.Fatalf("expected timeout error, got nil")
	}

	if err.Error() != "acquiring lock "+lockPath+": timeout after 100ms" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWithFileLock_RemovesStaleLock(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	lockPath := filepath.Join(dir, "stale.lock")
	if err := os.WriteFile(lockPath, []byte("stale"), 0o600); err != nil {
		t.Fatalf("seed stale lock: %v", err)
	}
	staleSince := time.Now().Add(-10 * time.Minute)
	if err := os.Chtimes(lockPath, staleSince, staleSince); err != nil {
		t.Fatalf("set stale mtime: %v", err)
	}

	acquired := make(chan time.Time, 2)
	errors := make(chan error, 2)
	release := make(chan struct{}, 2)

	for i := 0; i < 2; i++ {
		go func() {
			errors <- WithFileLock(lockPath, 2*time.Second, time.Minute, func() error {
				acquired <- time.Now()
				<-release
				return nil
			})
		}()
	}

	first := <-acquired
	select {
	case second := <-acquired:
		t.Fatalf("second lock acquired before first was released: first=%s second=%s", first, second)
	default:
	}

	release <- struct{}{}
	second := <-acquired
	if !second.After(first) {
		t.Fatalf("second lock acquired too early: first=%s second=%s", first, second)
	}
	release <- struct{}{}

	for i := 0; i < 2; i++ {
		if err := <-errors; err != nil {
			t.Fatalf("WithFileLock: %v", err)
		}
	}

	if FileExists(lockPath) {
		t.Fatalf("expected stale lock to be removed")
	}
}
