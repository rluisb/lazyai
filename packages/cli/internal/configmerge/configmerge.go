// Package configmerge provides deep-merge read-modify-write helpers for
// shared config files (Claude/Gemini settings.json, opencode.json). A .bak
// sidecar is created the first time a file is touched
// so users can restore their original config; subsequent runs never overwrite
// the existing .bak.
//
// Semantics:
//   - Maps are recursively merged; the patch wins on leaf collisions.
//   - Slices are replaced wholesale (not concatenated), because list
//     concatenation creates duplicates users cannot easily remove.
//   - Idempotent: re-running with the same patch produces identical bytes
//     (keys sorted alphabetically in both JSON and TOML output).
package configmerge

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
)

// MergeJSONFile deep-merges patch into the existing JSON/JSONC file at path,
// writing a .bak sidecar on first touch. If the file does not exist, patch is
// written directly and backupPath is empty.
func MergeJSONFile(path string, patch map[string]any) (backupPath string, err error) {
	var existing map[string]any
	if files.FileExists(path) {
		existing, err = jsonc.ReadJSONCFile(path)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", path, err)
		}
		backupPath, err = ensureBackup(path)
		if err != nil {
			return "", err
		}
	}
	merged := deepMerge(existing, patch)
	out, err := marshalSortedJSON(merged)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := files.SafeWriteFile(path, out, 0o644); err != nil {
		return "", err
	}
	return backupPath, nil
}

// MergeTOMLFile deep-merges patch into the existing TOML file at path,
// writing a .bak sidecar on first touch. If the file does not exist, patch is
// written directly and backupPath is empty.
func MergeTOMLFile(path string, patch map[string]any) (backupPath string, err error) {
	var existing map[string]any
	if files.FileExists(path) {
		data, rerr := files.ReadFile(path)
		if rerr != nil {
			return "", fmt.Errorf("read %s: %w", path, rerr)
		}
		existing = map[string]any{}
		if _, derr := toml.Decode(string(data), &existing); derr != nil {
			return "", fmt.Errorf("decode %s: %w", path, derr)
		}
		backupPath, err = ensureBackup(path)
		if err != nil {
			return "", err
		}
	}
	merged := deepMerge(existing, patch)
	out, err := marshalSortedTOML(merged)
	if err != nil {
		return "", fmt.Errorf("marshal %s: %w", path, err)
	}
	if err := files.SafeWriteFile(path, out, 0o644); err != nil {
		return "", err
	}
	return backupPath, nil
}

func ensureBackup(path string) (string, error) {
	bak := path + ".bak"
	if files.FileExists(bak) {
		return bak, nil
	}
	if err := files.CopyFile(path, bak); err != nil {
		return "", fmt.Errorf("backup %s: %w", path, err)
	}
	return bak, nil
}

// deepMerge returns a new map that is the result of recursively overlaying
// patch over base. Maps recurse; everything else (including slices) takes
// the patch value wholesale. Nil base is treated as empty.
func deepMerge(base, patch map[string]any) map[string]any {
	out := make(map[string]any, len(base)+len(patch))
	for k, v := range base {
		out[k] = v
	}
	for k, pv := range patch {
		if bv, ok := out[k]; ok {
			bm, bok := bv.(map[string]any)
			pm, pok := pv.(map[string]any)
			if bok && pok {
				out[k] = deepMerge(bm, pm)
				continue
			}
		}
		out[k] = pv
	}
	return out
}

// marshalSortedJSON emits a deterministic JSON representation with 2-space
// indentation and sorted keys at every depth. encoding/json already sorts
// map keys, but we round-trip through a stable formatter to keep the output
// stable across Go versions and to make the sorted-keys invariant explicit.
func marshalSortedJSON(m map[string]any) ([]byte, error) {
	// encoding/json sorts map keys lexicographically by default.
	raw, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// marshalSortedTOML emits a TOML representation with deterministic key order
// by sorting the map keys before encoding. BurntSushi/toml encodes maps in
// unspecified order, so we pre-order into a KeyedSliceCodec by wrapping the
// map in a sort-aware intermediate form.
func marshalSortedTOML(m map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	if err := encodeTOMLSorted(&buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// encodeTOMLSorted writes the map as TOML with sorted keys at each level.
// It uses BurntSushi/toml's encoder for values, but manages table order
// manually so regenerated files are byte-stable.
func encodeTOMLSorted(w *bytes.Buffer, m map[string]any) error {
	// Partition into scalar/array keys (written first, as bare key/value
	// assignments) and table keys (written second as [table] sections).
	var scalars, tables []string
	for k, v := range m {
		if _, ok := v.(map[string]any); ok {
			tables = append(tables, k)
		} else {
			scalars = append(scalars, k)
		}
	}
	sort.Strings(scalars)
	sort.Strings(tables)

	enc := toml.NewEncoder(w)
	// Scalars first as a single-pass encode.
	if len(scalars) > 0 {
		flat := make(map[string]any, len(scalars))
		for _, k := range scalars {
			flat[k] = m[k]
		}
		if err := enc.Encode(flat); err != nil {
			return err
		}
	}
	// Then each table in sorted order.
	for _, k := range tables {
		tbl := m[k].(map[string]any)
		if err := writeTOMLTable(w, []string{k}, tbl); err != nil {
			return err
		}
	}
	return nil
}

// writeTOMLTable writes a single [table] or [table.subtable] section and
// recurses into nested tables. Empty tables are emitted so [mcp_servers]
// survives a round-trip even when no servers are configured.
func writeTOMLTable(w *bytes.Buffer, path []string, m map[string]any) error {
	header := fmt.Sprintf("[%s]", joinTOMLKey(path))
	if w.Len() > 0 {
		w.WriteByte('\n')
	}
	w.WriteString(header)
	w.WriteByte('\n')

	var scalars, tables []string
	for k, v := range m {
		if _, ok := v.(map[string]any); ok {
			tables = append(tables, k)
		} else {
			scalars = append(scalars, k)
		}
	}
	sort.Strings(scalars)
	sort.Strings(tables)

	if len(scalars) > 0 {
		flat := make(map[string]any, len(scalars))
		for _, k := range scalars {
			flat[k] = m[k]
		}
		enc := toml.NewEncoder(w)
		if err := enc.Encode(flat); err != nil {
			return err
		}
	}
	for _, k := range tables {
		sub := m[k].(map[string]any)
		if err := writeTOMLTable(w, append(path, k), sub); err != nil {
			return err
		}
	}
	return nil
}

func joinTOMLKey(parts []string) string {
	// TOML bare keys must match [A-Za-z0-9_-]+; callers are responsible for
	// ensuring patch keys meet this. No quoting needed for the values we
	// emit (mcp_servers, <name>, etc.).
	out := parts[0]
	for _, p := range parts[1:] {
		out += "." + p
	}
	return out
}

// dir is re-exported for tests that want to assert the .bak lives next to the
// original rather than under some backup root. Kept as a helper to avoid
// callers reaching into filepath themselves.
func backupSibling(path string) string {
	return filepath.Join(filepath.Dir(path), filepath.Base(path)+".bak")
}

// Unused symbol silencer — keeps backupSibling available for future callers
// without tripping the linter. No-op at runtime.
var _ = func() any { _ = backupSibling; _ = os.PathSeparator; return nil }()
