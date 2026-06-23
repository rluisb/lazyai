// Package writer applies a plan.Plan to disk using a managed-region-aware,
// atomic file writer and returns the updated lockfile. Whole-file outputs own
// the entire file; managed outputs merge into a marked region, preserving user
// content outside the markers (reusing adapter.MergeManagedBlock).
package writer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/lockfile"
	"github.com/rluisb/lazyai/packages/cli/internal/plan"
)

// Managed-region markers, shared with the adapter root-AGENTS patcher so a file
// written by either path stays compatible.
const (
	StartMarker = adapter.ManagedBlockStartMarker
	EndMarker   = adapter.ManagedBlockEndMarker
)

// Options controls writer behavior.
type Options struct {
	// Force applies drift writes that would otherwise be refused.
	Force bool
	// DryRun plans without touching disk or the lockfile.
	DryRun bool
	// LazyaiVersion is recorded in the lockfile.
	LazyaiVersion string
}

// Result records the outcome of one planned write.
type Result struct {
	Path    string
	Target  string
	Action  plan.Action
	Wrote   bool
	Skipped bool
	Err     error
}

// Apply executes the plan and returns the updated lock and per-write results.
// On any unforced drift it returns a non-nil error after recording all results;
// no drift file is written unless Force is set.
func Apply(p plan.Plan, prev *lockfile.Lock, opts Options) (*lockfile.Lock, []Result, error) {
	if prev == nil {
		prev = &lockfile.Lock{Version: lockfile.SchemaVersion}
	}
	next := &lockfile.Lock{
		Version:       lockfile.SchemaVersion,
		LazyaiVersion: opts.LazyaiVersion,
		CompiledAt:    time.Now().UTC().Format(time.RFC3339),
		Adapters:      cloneAdapters(prev.Adapters),
		Generated:     append([]lockfile.Generated(nil), prev.Generated...),
	}

	results := make([]Result, 0, len(p.Writes))
	var driftErr error

	for _, w := range p.Writes {
		res := Result{Path: w.Path, Target: w.Target, Action: w.Action}

		switch w.Action {
		case plan.Skip:
			res.Skipped = true
			results = append(results, res)
			continue

		case plan.Drift:
			if !opts.Force {
				res.Err = fmt.Errorf("drift: %s (%s); re-run with --force to overwrite", w.Path, w.Reason)
				if driftErr == nil {
					driftErr = res.Err
				}
				results = append(results, res)
				continue
			}
			// Forced: fall through and write as a whole-file/managed update.

		case plan.Create, plan.Update:
			// normal path
		}

		if opts.DryRun {
			results = append(results, res)
			continue
		}

		outHash, err := writeOne(w)
		if err != nil {
			res.Err = err
			results = append(results, res)
			continue
		}
		res.Wrote = true
		next.Upsert(lockfile.Generated{
			Path:       w.Path,
			Target:     w.Target,
			SourceHash: w.SourceHash,
			OutputHash: outHash,
			Managed:    w.Managed,
		})
		results = append(results, res)
	}

	if opts.DryRun {
		return prev, results, driftErr
	}
	return next, results, driftErr
}

// writeOne writes a single planned output atomically and returns the hash of
// the bytes actually written to disk.
func writeOne(w plan.Write) (string, error) {
	if err := os.MkdirAll(filepath.Dir(w.Path), 0o755); err != nil {
		return "", fmt.Errorf("creating parent of %s: %w", w.Path, err)
	}
	final := w.Content
	if w.Managed {
		existing, err := os.ReadFile(w.Path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("reading %s: %w", w.Path, err)
		}
		final = adapter.MergeManagedBlock(existing, w.Content, StartMarker, EndMarker)
	}
	if err := atomicWrite(w.Path, final); err != nil {
		return "", err
	}
	return lockfile.HashBytes(final), nil
}

// atomicWrite writes via a temp file + rename in the destination directory.
// It fsyncs the temp file before rename so the data is on stable storage
// before it replaces the target — crash-safety for the write itself.
func atomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".lazyai-*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp for %s: %w", path, err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("writing temp for %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("syncing temp for %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp for %s: %w", path, err)
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return fmt.Errorf("chmod temp for %s: %w", path, err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("renaming temp into %s: %w", path, err)
	}
	return nil
}

func cloneAdapters(in map[string]lockfile.AdapterLock) map[string]lockfile.AdapterLock {
	out := make(map[string]lockfile.AdapterLock, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
