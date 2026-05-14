package domain

// HunkStarts computes the starting diff-line indexes for changed hunks.
func HunkStarts(diffLines []DiffLine) []int {
	hunkStarts := make([]int, 0)
	inHunk := false
	for i, diffLine := range diffLines {
		isChange := diffLine.Type != DiffLineContext
		if isChange && !inHunk {
			hunkStarts = append(hunkStarts, i)
		}
		inHunk = isChange
	}
	return hunkStarts
}
