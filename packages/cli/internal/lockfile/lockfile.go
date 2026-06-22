package lockfile

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Generated tracks a source file and its corresponding compiled output hash.
type Generated struct {
	Path       string `json:"path"`
	Target     string `json:"target"`
	SourceHash string `json:"sourceHash"`
	OutputHash string `json:"outputHash"`
	Managed    bool   `json:"managed"`
}

// AdapterLock describes the lock metadata for a single adapter.
type AdapterLock struct {
	Version      string `json:"version"`
	DocsSnapshot string `json:"docsSnapshot,omitempty"`
}

// SchemaVersion is the current .ai/lock.json schema version.
const SchemaVersion = "1.0"

// Lock is the persisted compilation lockfile metadata.
type Lock struct {
	Version       string                 `json:"version"`
	LazyaiVersion string                 `json:"lazyaiVersion"`
	CompiledAt    string                 `json:"compiledAt"`
	Adapters      map[string]AdapterLock `json:"adapters"`
	Generated     []Generated            `json:"generated"`
}

// Load reads the lockfile at <aiDir>/lock.json.
// If the file does not exist, it returns a lock with default values.
func Load(aiDir string) (*Lock, error) {
	path := filepath.Join(aiDir, "lock.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Lock{
				Version:   SchemaVersion,
				Adapters:  map[string]AdapterLock{},
				Generated: []Generated{},
			}, nil
		}
		return nil, fmt.Errorf("read lockfile %q: %w", path, err)
	}

	var lock Lock
	if err := json.Unmarshal(data, &lock); err != nil {
		return nil, fmt.Errorf("unmarshal lockfile %q: %w", path, err)
	}

	return &lock, nil
}

// Save writes the lockfile to <aiDir>/lock.json.
// The JSON is indented with two spaces, always ends with a trailing newline,
// and is written with mode 0644 in a 0755 directory.
func (l *Lock) Save(aiDir string) error {
	if l.Adapters == nil {
		l.Adapters = map[string]AdapterLock{}
	}
	if l.Generated == nil {
		l.Generated = []Generated{}
	}

	sort.Slice(l.Generated, func(i, j int) bool {
		return l.Generated[i].Path < l.Generated[j].Path
	})

	data, err := json.MarshalIndent(l, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal lock: %w", err)
	}

	if len(data) > 0 && data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		return fmt.Errorf("create lock directory %q: %w", aiDir, err)
	}

	path := filepath.Join(aiDir, "lock.json")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write lockfile %q: %w", path, err)
	}

	return nil
}

// HashBytes returns the SHA-256 hash for data, prefixed with "sha256:".
func HashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(sum[:])
}

// Find returns the generated entry by path.
func (l *Lock) Find(path string) (Generated, bool) {
	for _, generated := range l.Generated {
		if generated.Path == path {
			return generated, true
		}
	}
	return Generated{}, false
}

// Upsert inserts or replaces a generated entry and keeps entries sorted by path.
func (l *Lock) Upsert(g Generated) {
	for i, generated := range l.Generated {
		if generated.Path == g.Path {
			l.Generated[i] = g
			sort.Slice(l.Generated, func(i, j int) bool {
				return l.Generated[i].Path < l.Generated[j].Path
			})
			return
		}
	}

	l.Generated = append(l.Generated, g)
	sort.Slice(l.Generated, func(i, j int) bool {
		return l.Generated[i].Path < l.Generated[j].Path
	})
}
