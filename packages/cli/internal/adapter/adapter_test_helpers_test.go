package adapter

import "sort"

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// baselineAgentSource matches the exact vibe-lab baseline canonical agent shape:
// name and description only, no tier/risk/model/temperature/mode/steps/skills.
func baselineAgentSource(name, description string) []byte {
	return []byte("---\nname: " + name + "\ndescription: " + description + "\n---\n\n# System Prompt\n\nYou are " + name + ".")
}

func sortedStrings(s []string) []string {
	sort.Strings(s)
	return s
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
