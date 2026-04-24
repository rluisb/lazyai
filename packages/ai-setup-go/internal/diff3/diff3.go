// Package diff3 provides a 3-way merge algorithm for line-level comparison
// and conflict detection. Mirrors the TS implementation at
// packages/ai-setup-ts/src/migration/diff/diff3.ts so both runtimes produce
// the same merged output (including conflict marker shape) on equivalent input.
package diff3

import "strings"

// DiffLine is a line with its source-array index.
type DiffLine struct {
	Line  string
	Index int
}

// DiffResult is the classified output of a two-way diff.
type DiffResult struct {
	Added     []DiffLine
	Removed   []DiffLine
	Unchanged []DiffLine
}

// Conflict describes a single unresolved region in a 3-way merge.
type Conflict struct {
	LineStart int
	LineEnd   int
	Base      []string
	Ours      []string
	Theirs    []string
}

// Diff3Result is the output of a 3-way merge.
type Diff3Result struct {
	Merged       []string
	Conflicts    []Conflict
	HasConflicts bool
}

// Conflict marker strings (unicode-escaped in TS for Biome reasons; we use
// literal bytes here but the output is byte-identical).
const (
	markerOurs    = "<<<<< OURS"
	markerMiddle  = "====="
	markerTheirs  = ">>>>> THEIRS"
)

type editType int

const (
	editEqual editType = iota
	editInsert
	editDelete
)

type edit struct {
	kind editType
}

// MyersDiff computes the added/removed/unchanged classification between two
// line arrays using an LCS-based edit script. Mirrors TS's `myersDiff`.
func MyersDiff(oldLines, newLines []string) DiffResult {
	result := DiffResult{}

	if len(oldLines) == 0 && len(newLines) == 0 {
		return result
	}
	if len(oldLines) == 0 {
		result.Added = make([]DiffLine, len(newLines))
		for i, line := range newLines {
			result.Added[i] = DiffLine{Line: line, Index: i}
		}
		return result
	}
	if len(newLines) == 0 {
		result.Removed = make([]DiffLine, len(oldLines))
		for i, line := range oldLines {
			result.Removed[i] = DiffLine{Line: line, Index: i}
		}
		return result
	}

	edits := computeEdits(oldLines, newLines)

	oldIndex, newIndex := 0, 0
	for _, e := range edits {
		switch e.kind {
		case editEqual:
			if oldIndex < len(oldLines) {
				result.Unchanged = append(result.Unchanged, DiffLine{Line: oldLines[oldIndex], Index: oldIndex})
			}
			oldIndex++
			newIndex++
		case editDelete:
			if oldIndex < len(oldLines) {
				result.Removed = append(result.Removed, DiffLine{Line: oldLines[oldIndex], Index: oldIndex})
			}
			oldIndex++
		case editInsert:
			if newIndex < len(newLines) {
				result.Added = append(result.Added, DiffLine{Line: newLines[newIndex], Index: newIndex})
			}
			newIndex++
		}
	}

	return result
}

// computeEdits builds an LCS table and backtracks to produce the edit script.
// Matches the TS `computeEdits` implementation byte-for-byte in its choices.
func computeEdits(oldLines, newLines []string) []edit {
	oldLen := len(oldLines)
	newLen := len(newLines)

	// Build LCS table.
	table := make([][]int, oldLen+1)
	for i := range table {
		table[i] = make([]int, newLen+1)
	}
	for i := 1; i <= oldLen; i++ {
		for j := 1; j <= newLen; j++ {
			if oldLines[i-1] == newLines[j-1] {
				table[i][j] = table[i-1][j-1] + 1
			} else {
				if table[i-1][j] > table[i][j-1] {
					table[i][j] = table[i-1][j]
				} else {
					table[i][j] = table[i][j-1]
				}
			}
		}
	}

	// Backtrack to produce edits.
	i, j := oldLen, newLen
	// Build in reverse, then flip.
	var reversed []edit
	for i > 0 || j > 0 {
		switch {
		case i > 0 && j > 0 && oldLines[i-1] == newLines[j-1]:
			reversed = append(reversed, edit{kind: editEqual})
			i--
			j--
		case j > 0 && (i == 0 || table[i][j-1] >= table[i-1][j]):
			reversed = append(reversed, edit{kind: editInsert})
			j--
		default:
			reversed = append(reversed, edit{kind: editDelete})
			i--
		}
	}

	// Reverse in place.
	edits := make([]edit, len(reversed))
	for k, e := range reversed {
		edits[len(reversed)-1-k] = e
	}
	return edits
}

// Diff3 performs a three-way merge of base, ours, and theirs. Returns the
// merged line array, any unresolved conflict regions, and a `HasConflicts`
// summary flag. Mirrors TS's `diff3`.
func Diff3(base, ours, theirs []string) Diff3Result {
	diffOurs := MyersDiff(base, ours)
	diffTheirs := MyersDiff(base, theirs)

	var merged []string
	var conflicts []Conflict

	baseIndex, oursIndex, theirsIndex := 0, 0, 0

	for baseIndex < len(base) || oursIndex < len(ours) || theirsIndex < len(theirs) {
		oursChanged := isChangedAt(diffOurs, baseIndex)
		theirsChanged := isChangedAt(diffTheirs, baseIndex)

		switch {
		case !oursChanged && !theirsChanged:
			// No changes — take base.
			if baseIndex < len(base) {
				merged = append(merged, base[baseIndex])
				baseIndex++
			}
			oursIndex++
			theirsIndex++

		case oursChanged && !theirsChanged:
			if oursIndex < len(ours) {
				merged = append(merged, ours[oursIndex])
			}
			baseIndex++
			oursIndex++
			theirsIndex++

		case !oursChanged && theirsChanged:
			if theirsIndex < len(theirs) {
				merged = append(merged, theirs[theirsIndex])
			}
			baseIndex++
			oursIndex++
			theirsIndex++

		default:
			// Both sides changed at this position.
			var ourLine, theirLine string
			hasOur := oursIndex < len(ours)
			hasTheir := theirsIndex < len(theirs)
			if hasOur {
				ourLine = ours[oursIndex]
			}
			if hasTheir {
				theirLine = theirs[theirsIndex]
			}

			if hasOur && hasTheir && ourLine == theirLine {
				merged = append(merged, ourLine)
				baseIndex++
				oursIndex++
				theirsIndex++
			} else {
				conflictStart := len(merged)
				merged = append(merged, markerOurs)
				for oursIndex < len(ours) && isChangedAt(diffOurs, baseIndex) {
					merged = append(merged, ours[oursIndex])
					oursIndex++
				}
				merged = append(merged, markerMiddle)
				for theirsIndex < len(theirs) && isChangedAt(diffTheirs, baseIndex) {
					merged = append(merged, theirs[theirsIndex])
					theirsIndex++
				}
				merged = append(merged, markerTheirs)

				baseSlice := []string{}
				if baseIndex < len(base) {
					baseSlice = []string{base[baseIndex]}
				}
				oursSlice := []string{}
				if oursIndex > 0 && oursIndex-1 < len(ours) {
					oursSlice = []string{ours[oursIndex-1]}
				}
				theirsSlice := []string{}
				if theirsIndex > 0 && theirsIndex-1 < len(theirs) {
					theirsSlice = []string{theirs[theirsIndex-1]}
				}

				conflicts = append(conflicts, Conflict{
					LineStart: conflictStart + 1,
					LineEnd:   len(merged),
					Base:      baseSlice,
					Ours:      oursSlice,
					Theirs:    theirsSlice,
				})

				baseIndex++
			}
		}
	}

	return Diff3Result{
		Merged:       merged,
		Conflicts:    conflicts,
		HasConflicts: len(conflicts) > 0,
	}
}

func isChangedAt(d DiffResult, index int) bool {
	for _, l := range d.Added {
		if l.Index == index {
			return true
		}
	}
	for _, l := range d.Removed {
		if l.Index == index {
			return true
		}
	}
	return false
}

// Merge2Way applies a set of changes onto a base, producing a merged array.
// Mirrors TS's `merge2Way`.
func Merge2Way(base, changes []string) []string {
	var result []string
	d := MyersDiff(base, changes)

	baseIndex, changesIndex := 0, 0
	for baseIndex < len(base) || changesIndex < len(changes) {
		isDeleted := false
		for _, l := range d.Removed {
			if l.Index == baseIndex {
				isDeleted = true
				break
			}
		}
		isAdded := false
		for _, l := range d.Added {
			if l.Index == changesIndex {
				isAdded = true
				break
			}
		}

		switch {
		case isAdded:
			if changesIndex < len(changes) {
				result = append(result, changes[changesIndex])
			}
			changesIndex++
		case !isDeleted:
			if baseIndex < len(base) {
				result = append(result, base[baseIndex])
			}
			baseIndex++
			changesIndex++
		default:
			baseIndex++
		}
	}

	return result
}

// HasConflictMarkers returns true if the merged line array contains any of
// the conflict marker prefixes. Mirrors TS's `hasConflicts`.
func HasConflictMarkers(merged []string) bool {
	for _, line := range merged {
		if strings.Contains(line, "<<<<<") || strings.Contains(line, "=====") || strings.Contains(line, ">>>>>") {
			return true
		}
	}
	return false
}

// Resolution picks one side of a conflict.
type Resolution string

const (
	ResolveOurs   Resolution = "ours"
	ResolveTheirs Resolution = "theirs"
	ResolveBase   Resolution = "base"
)

// ResolveConflicts walks the merged line array and emits a conflict-free
// result by taking only lines from the chosen side within each conflict
// region. Mirrors TS's `resolveConflicts`.
func ResolveConflicts(merged []string, resolution Resolution) []string {
	var result []string
	inConflict := false
	var side Resolution

	for _, line := range merged {
		switch {
		case strings.HasPrefix(line, "<<<<<"):
			inConflict = true
			side = ResolveOurs
		case strings.HasPrefix(line, "====="):
			side = ResolveTheirs
		case strings.HasPrefix(line, ">>>>>"):
			inConflict = false
			side = ""
		case inConflict:
			if side == resolution {
				result = append(result, line)
			}
		default:
			result = append(result, line)
		}
	}

	return result
}
