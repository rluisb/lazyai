// Package plan computes the diff between desired compile outputs and the
// recorded lockfile + on-disk state. It is pure: it reads nothing itself,
// taking a DiskReader so it stays unit-testable. The writer package applies a
// Plan.
package plan

import "github.com/rluisb/lazyai/packages/cli/internal/lockfile"

// Action is the resolved disposition of a single desired output.
type Action string

const (
	// Create writes a brand-new file.
	Create Action = "create"
	// Update rewrites a tracked file (whole-file) or re-merges a managed region.
	Update Action = "update"
	// Skip leaves an up-to-date file untouched.
	Skip Action = "skip"
	// Drift marks a whole-file output that changed outside lazyai or a
	// pre-existing untracked file; the writer refuses it without Force.
	Drift Action = "drift"
)

// Desired is one compile output the adapters want on disk.
type Desired struct {
	Target     string
	Path       string
	SourceHash string
	Content    []byte
	// Managed marks files written via a managed region (user content outside
	// the markers is preserved). Whole-file outputs (Managed=false) own the
	// entire file.
	Managed bool
}

// Write is a planned action for one Desired output.
type Write struct {
	Desired
	OutputHash string
	Action     Action
	Reason     string
}

// Plan is the ordered set of planned writes.
type Plan struct {
	Writes []Write
}

// DiskReader returns the current bytes at path and whether it exists.
type DiskReader func(path string) ([]byte, bool)

// Build computes the plan from desired outputs, the prior lock, and disk state.
//
// Idempotency: for whole-file outputs a Skip requires identical source and
// output hashes and a clean on-disk hash; for managed outputs a Skip requires
// an unchanged source and an on-disk hash matching the lock (which records the
// merged result), so a re-run after a clean compile is a no-op.
func Build(desired []Desired, lock *lockfile.Lock, read DiskReader) Plan {
	if lock == nil {
		lock = &lockfile.Lock{}
	}
	p := Plan{Writes: make([]Write, 0, len(desired))}
	for _, d := range desired {
		w := Write{Desired: d, OutputHash: lockfile.HashBytes(d.Content)}
		diskContent, exists := read(d.Path)
		prev, hasPrev := lock.Find(d.Path)

		switch {
		case !exists:
			w.Action, w.Reason = Create, "file does not exist"
		case !hasPrev:
			if d.Managed {
				w.Action, w.Reason = Update, "adopt: insert managed region into untracked file"
			} else {
				w.Action, w.Reason = Drift, "pre-existing untracked file would be overwritten"
			}
		default:
			diskHash := lockfile.HashBytes(diskContent)
			diskMatchesLock := diskHash == prev.OutputHash
			sameSource := prev.SourceHash == d.SourceHash
			switch {
			case !diskMatchesLock && !d.Managed:
				w.Action, w.Reason = Drift, "file modified outside lazyai since last compile"
			case !diskMatchesLock && d.Managed:
				w.Action, w.Reason = Update, "re-merge managed region (preserving out-of-region edits)"
			case sameSource && (d.Managed || prev.OutputHash == w.OutputHash):
				w.Action, w.Reason = Skip, "up to date"
			default:
				w.Action, w.Reason = Update, "source changed"
			}
		}
		p.Writes = append(p.Writes, w)
	}
	return p
}

// Count returns the number of writes with the given action.
func (p Plan) Count(a Action) int {
	n := 0
	for _, w := range p.Writes {
		if w.Action == a {
			n++
		}
	}
	return n
}

// HasDrift reports whether any planned write is a drift conflict.
func (p Plan) HasDrift() bool { return p.Count(Drift) > 0 }
