// Package files provides file I/O utilities for the ai-setup project.
// Ported from the TypeScript utilities in src/utils/files.ts.
package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
)

// ReadFile reads the contents of the file at path and returns its bytes.
func ReadFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, aierror.FileNotFound(path)
		}
		if os.IsPermission(err) {
			return nil, aierror.FilePermission(path, "read")
		}
		return nil, aierror.FileCorrupt(path, err)
	}
	return data, nil
}

// WriteFile writes data to the file at path, creating parent directories as needed.
func WriteFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		if os.IsPermission(err) {
			return aierror.FilePermission(path, "write")
		}
		return aierror.Unknown(fmt.Sprintf("write failed: %s", path), err)
	}
	return nil
}

// FileHash returns the first 16 hex characters of the SHA256 hash of the file at path.
func FileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", aierror.FileNotFound(path)
		}
		return "", aierror.FilePermission(path, "hash")
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", aierror.FilePermission(path, "hash")
	}
	full := hex.EncodeToString(h.Sum(nil))
	if len(full) > 16 {
		full = full[:16]
	}
	return full, nil
}

// EnsureDir creates the directory (and any parents) if it doesn't exist.
func EnsureDir(path string) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return aierror.FilePermission(path, "create directory")
	}
	return nil
}

// CopyFile copies a single file from src to dst, creating parent dirs for dst.
func CopyFile(src, dst string) error {
	if err := EnsureDir(filepath.Dir(dst)); err != nil {
		return err
	}

	in, err := os.Open(src)
	if err != nil {
		return aierror.Unknown(fmt.Sprintf("copy failed: %s → %s", src, dst), err)
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return aierror.Unknown(fmt.Sprintf("copy failed: %s → %s", src, dst), err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return aierror.Unknown(fmt.Sprintf("copy failed: %s → %s", src, dst), err)
	}

	// Copy permissions from source.
	info, err := os.Stat(src)
	if err != nil {
		return nil // non-fatal
	}
	os.Chmod(dst, info.Mode())

	return nil
}

// CopyDir recursively copies the directory tree from src to dst.
func CopyDir(src, dst string) error {
	if !DirExists(src) {
		return aierror.DirNotFound(src)
	}

	if err := EnsureDir(dst); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return aierror.FilePermission(src, "readdir")
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if entry.IsDir() {
			if err := CopyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := CopyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}
	return nil
}

// FileExists reports whether a file (or directory) exists at path.
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists reports whether path exists and is a directory.
func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FindPackageRoot walks up from startDir until it finds a directory
// containing a go.mod (Go equivalent of package.json for this project),
// then returns that directory path.
func FindPackageRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if FileExists(filepath.Join(dir, "go.mod")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", aierror.DirNotFound(fmt.Sprintf("Could not find package root from: %s", startDir))
		}
		dir = parent
	}
}

// RemoveAll removes the directory tree at path.
func RemoveAll(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return aierror.FilePermission(path, "remove")
	}
	return nil
}

// ListDir returns the names of files and directories in dirPath.
// Returns an empty slice on error.
func ListDir(dirPath string) []string {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return []string{}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// IsDirectory reports whether path is a directory.
func IsDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// BackupFile creates a backup of filePath inside targetDir/.ai-setup-backup/,
// returning the backup path. If a backup already exists, a timestamp suffix is added.
func BackupFile(filePath string, targetDir string) (string, error) {
	backupRoot := filepath.Join(targetDir, ".ai-setup-backup")
	if err := EnsureDir(backupRoot); err != nil {
		return "", err
	}

	relativePath, err := filepath.Rel(targetDir, filePath)
	if err != nil || startsDotDot(relativePath) {
		relativePath = filepath.Base(filePath)
	}

	// Normalize to forward slashes for consistent display.
	normalizedRelativePath := filepath.ToSlash(relativePath)
	backupPath := filepath.Join(backupRoot, normalizedRelativePath)

	if FileExists(backupPath) {
		backupPath = fmt.Sprintf("%s.%d", backupPath, time.Now().UnixMilli())
	}

	if err := EnsureDir(filepath.Dir(backupPath)); err != nil {
		return "", err
	}

	if err := CopyFile(filePath, backupPath); err != nil {
		return "", err
	}

	return backupPath, nil
}

// CreateTimestampedBackup creates a timestamped sibling backup ending in .bak.
// It supports both files and directories and is intended for overwrite flows.
func CreateTimestampedBackup(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", aierror.FileNotFound(path)
		}
		return "", aierror.FilePermission(path, "backup")
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	backupPath := fmt.Sprintf("%s.%s.bak", path, timestamp)
	for attempt := 1; FileExists(backupPath); attempt++ {
		backupPath = fmt.Sprintf("%s.%s.%d.bak", path, timestamp, attempt)
	}

	if info.IsDir() {
		if err := CopyDir(path, backupPath); err != nil {
			return "", err
		}
		return backupPath, nil
	}

	if err := CopyFile(path, backupPath); err != nil {
		return "", err
	}
	return backupPath, nil
}

// startsDotDot reports whether a path starts with "..".
func startsDotDot(p string) bool {
	return len(p) >= 2 && p[0] == '.' && p[1] == '.'
}

// ---------------------------------------------------------------------------
// fs.FS-based helpers — for reading from embedded or in-memory filesystems
// ---------------------------------------------------------------------------

// ReadFS reads the contents of a file from an fs.FS filesystem.
func ReadFS(fsys fs.FS, path string) ([]byte, error) {
	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, aierror.FileNotFound(path)
		}
		return nil, aierror.FileCorrupt(path, err)
	}
	return data, nil
}

// ExistsFS reports whether a file (or directory) exists in an fs.FS filesystem.
func ExistsFS(fsys fs.FS, path string) bool {
	_, err := fs.Stat(fsys, path)
	return err == nil
}

// ReadDirFS returns directory entries from an fs.FS filesystem.
func ReadDirFS(fsys fs.FS, path string) ([]fs.DirEntry, error) {
	entries, err := fs.ReadDir(fsys, path)
	if err != nil {
		return nil, aierror.FilePermission(path, "readdir")
	}
	return entries, nil
}
