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

// SafeWriteFile writes data to path via temp-file + sync + rename for crash
// safety, without creating a .bak backup. Use this when the caller manages its
// own backup strategy or the file is regenerable.
func SafeWriteFile(path string, data []byte, perm os.FileMode) error {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return err
	}
	tmpPath := filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s.%d.tmp", filepath.Base(path), time.Now().UnixNano()))
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
	if err != nil {
		return aierror.Unknown(fmt.Sprintf("create temp file for: %s", path), err)
	}
	defer func() { _ = os.Remove(tmpPath) }()
	if _, err := f.Write(data); err != nil {
		f.Close()
		return aierror.Unknown(fmt.Sprintf("write temp file for: %s", path), err)
	}
	if err := f.Sync(); err != nil {
		f.Close()
		return aierror.Unknown(fmt.Sprintf("sync temp file for: %s", path), err)
	}
	if err := f.Close(); err != nil {
		return aierror.Unknown(fmt.Sprintf("close temp file for: %s", path), err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return aierror.Unknown(fmt.Sprintf("rename temp file for: %s", path), err)
	}
	tmpPath = "" // prevent deferred removal of the now-renamed file
	return nil
}

// AtomicWriteFile writes data to path atomically.
//
// - ensures the target directory exists
// - creates/refreshes a single-slot `.bak` file for existing regular files
// - writes through a temporary `.tmp` file, then renames into place
func AtomicWriteFile(path string, data []byte, perm os.FileMode) (string, error) {
	if err := EnsureDir(filepath.Dir(path)); err != nil {
		return "", err
	}

	backupPath := ""
	if info, err := os.Stat(path); err == nil && info.Mode().IsRegular() {
		backupPath = path + ".bak"
		if err := CopyFile(path, backupPath); err != nil {
			return backupPath, err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", aierror.Unknown(fmt.Sprintf("unable to inspect %s", path), err)
	}

	var (
		tmpPath string
		tmpFile *os.File
		err     error
	)
	for {
		tmpPath = filepath.Join(filepath.Dir(path), fmt.Sprintf(".%s.%d.tmp", filepath.Base(path), time.Now().UnixNano()))
		tmpFile, err = os.OpenFile(tmpPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, perm)
		if err == nil {
			break
		}
		if os.IsExist(err) {
			continue
		}
		return backupPath, aierror.Unknown(fmt.Sprintf("create temporary file for: %s", path), err)
	}

	defer func() {
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmpFile.Write(data); err != nil {
		tmpFile.Close()
		return backupPath, aierror.Unknown(fmt.Sprintf("write temporary file for: %s", path), err)
	}
	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return backupPath, aierror.Unknown(fmt.Sprintf("sync temporary file for: %s", path), err)
	}
	if err := tmpFile.Close(); err != nil {
		return backupPath, aierror.Unknown(fmt.Sprintf("close temporary file for: %s", path), err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		return backupPath, aierror.Unknown(fmt.Sprintf("rename temporary file for: %s", path), err)
	}

	tmpPath = ""
	return backupPath, nil
}

// WithFileLock acquires a lock file, retries until timeout, and runs fn while held.
func WithFileLock(lockPath string, timeout time.Duration, staleAfter time.Duration, fn func() error) error {
	if err := EnsureDir(filepath.Dir(lockPath)); err != nil {
		return err
	}

	deadline := time.Now().Add(timeout)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("acquiring lock %s: timeout after %s", lockPath, timeout)
		}

		lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if err == nil {
			payload := fmt.Sprintf("%d %s\n", os.Getpid(), time.Now().UTC().Format(time.RFC3339Nano))
			if _, writeErr := lockFile.WriteString(payload); writeErr != nil {
				_ = lockFile.Close()
				_ = os.Remove(lockPath)
				return aierror.Unknown(fmt.Sprintf("write lock file: %s", lockPath), writeErr)
			}
			if syncErr := lockFile.Sync(); syncErr != nil {
				_ = lockFile.Close()
				_ = os.Remove(lockPath)
				return aierror.Unknown(fmt.Sprintf("sync lock file: %s", lockPath), syncErr)
			}
			if closeErr := lockFile.Close(); closeErr != nil {
				_ = os.Remove(lockPath)
				return aierror.Unknown(fmt.Sprintf("close lock file: %s", lockPath), closeErr)
			}

			defer os.Remove(lockPath)
			return fn()
		}
		if !os.IsExist(err) {
			return aierror.Unknown(fmt.Sprintf("acquire lock %s", lockPath), err)
		}

		info, statErr := os.Stat(lockPath)
		if statErr != nil {
			if !os.IsNotExist(statErr) {
				return aierror.Unknown(fmt.Sprintf("inspect lock for stale cleanup %s", lockPath), statErr)
			}
			continue
		}
		if time.Since(info.ModTime()) <= staleAfter {
			time.Sleep(25 * time.Millisecond)
			continue
		}

		// Stale lock cleanup is serialized via an O_EXCL guard file to
		// prevent two cleaners from racing. The guard itself has a
		// stale timeout so a crashed cleaner doesn't permanently block
		// recovery: if the guard's mtime is older than staleAfter, we
		// remove it and retry. While holding the guard, re-stat the
		// main lock to confirm it's still the same stale file (not
		// recreated by another contender), then rename it to trash
		// and remove. The guard is always cleaned up before continuing.
		guardPath := lockPath + ".stale-cleanup"
		guardFile, guardErr := os.OpenFile(guardPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if guardErr != nil {
			if !os.IsExist(guardErr) {
				return aierror.Unknown(fmt.Sprintf("acquire stale cleanup guard %s", guardPath), guardErr)
			}
			// Guard exists — check if it's stale (cleaner crashed).
			guardInfo, guardStatErr := os.Stat(guardPath)
			if guardStatErr != nil && !os.IsNotExist(guardStatErr) {
				return aierror.Unknown(fmt.Sprintf("inspect stale cleanup guard %s", guardPath), guardStatErr)
			}
			if guardStatErr == nil && time.Since(guardInfo.ModTime()) > staleAfter {
				_ = os.Remove(guardPath)
			}
			time.Sleep(25 * time.Millisecond)
			continue
		}
		// Guard acquired — clean up the stale lock, then remove the guard.
		_ = guardFile.Close()
		if err := func() error {
			defer os.Remove(guardPath)
			refreshedInfo, refreshedErr := os.Stat(lockPath)
			if refreshedErr != nil {
				if !os.IsNotExist(refreshedErr) {
					return aierror.Unknown(fmt.Sprintf("inspect stale lock %s", lockPath), refreshedErr)
				}
				return nil // lock already gone
			}
			if !refreshedInfo.ModTime().Equal(info.ModTime()) {
				return nil // lock was recreated; not stale anymore
			}
			trashPath := filepath.Join(filepath.Dir(lockPath), fmt.Sprintf(".%s.stale.%d.%d.lock", filepath.Base(lockPath), os.Getpid(), time.Now().UnixNano()))
			if renameErr := os.Rename(lockPath, trashPath); renameErr != nil {
				if !os.IsNotExist(renameErr) {
					return aierror.Unknown(fmt.Sprintf("rename stale lock %s", lockPath), renameErr)
				}
				return nil // lock already gone
			}
			_ = os.Remove(trashPath)
			return nil
		}(); err != nil {
			return err
		}
		continue
	}
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

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return aierror.Unknown(fmt.Sprintf("copy failed: %s → %s", src, dst), err)
	}

	if err := out.Sync(); err != nil {
		out.Close()
		return aierror.Unknown(fmt.Sprintf("sync failed: %s → %s", src, dst), err)
	}

	if err := out.Close(); err != nil {
		return aierror.Unknown(fmt.Sprintf("close failed: %s → %s", src, dst), err)
	}

	// Copy permissions from source (non-fatal: best effort).
	info, err := os.Stat(src)
	if err != nil {
		return nil
	}
	_ = os.Chmod(dst, info.Mode()) // best-effort; non-fatal on platforms that ignore mode

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
