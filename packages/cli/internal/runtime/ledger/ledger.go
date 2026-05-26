// Package ledger provides an immutable append-only hash-chained ledger.
//
// This replaces the bash ledger.sh script with a Go-native implementation
// that maintains the same hash chain semantics but adds proper locking,
// secret redaction, and structured logging.
package ledger

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Ledger is an append-only hash-chained audit trail.
type Ledger struct {
	path     string
	lockPath string
	mu       sync.Mutex
}

// Entry represents a single ledger record.
type Entry struct {
	Seq         int             `json:"seq"`
	Timestamp   string          `json:"ts"`
	Type        string          `json:"type"`
	SessionID   string          `json:"session_id,omitempty"`
	WorkflowID  string          `json:"workflow_run_id,omitempty"`
	Agent       string          `json:"agent,omitempty"`
	Data        json.RawMessage `json:"data"`
	PrevHash    string          `json:"prev_hash"`
	Hash        string          `json:"hash"`
	Risk        string          `json:"risk,omitempty"`
	InputHash   string          `json:"input_hash,omitempty"`
	OutputHash  string          `json:"output_hash,omitempty"`
	Redactions  []string        `json:"redactions,omitempty"`
}

// secretKeys are fields that must be redacted before hashing.
var secretKeys = map[string]bool{
	"token": true, "access_token": true, "refresh_token": true,
	"auth_token": true, "bearer_token": true, "api_token": true,
	"jwt": true, "secret": true, "password": true,
	"api_key": true, "apikey": true, "key": true,
	"private_key": true, "client_secret": true, "client_id": true,
	"auth_code": true, "credential": true, "credentials": true,
}

// Open creates or opens a ledger at the given path.
func Open(path string) (*Ledger, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create ledger directory: %w", err)
	}

	l := &Ledger{
		path:     path,
		lockPath: path + ".lock",
	}

	// Initialize if empty
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := l.initGenesis(); err != nil {
			return nil, fmt.Errorf("init genesis: %w", err)
		}
	}

	return l, nil
}

// initGenesis creates the first ledger entry.
func (l *Ledger) initGenesis() error {
	genesis := &Entry{
		Seq:       1,
		Timestamp: now(),
		Type:      "genesis",
		Data:      json.RawMessage(`{}`),
		PrevHash:  strings.Repeat("0", 64),
	}
	genesis.Hash = l.computeHash(genesis)

	return l.writeEntry(genesis)
}

// Append adds a new entry to the ledger.
func (l *Ledger) Append(entry *Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Get last entry for prev_hash and seq
	last, err := l.readLastEntry()
	if err != nil {
		return fmt.Errorf("read last entry: %w", err)
	}

	entry.Seq = last.Seq + 1
	entry.Timestamp = now()
	entry.PrevHash = last.Hash

	// Redact secrets before hashing
	redacted, redactions := redactSecrets(entry.Data)
	entry.Data = redacted
	entry.Redactions = redactions

	entry.Hash = l.computeHash(entry)

	return l.writeEntry(entry)
}

// computeHash computes the SHA-256 hash of an entry.
// Uses pipe-delimited canonical form matching bash ledger.sh:
//   seq|ts|type|session_id|JSON(data)|prev_hash
func (l *Ledger) computeHash(entry *Entry) string {
	dataStr := string(entry.Data)
	canonical := fmt.Sprintf("%d|%s|%s|%s|%s|%s",
		entry.Seq,
		entry.Timestamp,
		entry.Type,
		entry.SessionID,
		dataStr,
		entry.PrevHash,
	)
	hash := sha256.Sum256([]byte(canonical))
	return hex.EncodeToString(hash[:])
}

// Verify checks the integrity of the entire ledger.
func (l *Ledger) Verify() error {
	entries, err := l.ReadAll()
	if err != nil {
		return fmt.Errorf("read ledger: %w", err)
	}

	if len(entries) == 0 {
		return fmt.Errorf("ledger is empty")
	}

	for i, entry := range entries {
		// Check genesis
		if i == 0 {
			if entry.PrevHash != strings.Repeat("0", 64) {
				return fmt.Errorf("entry %d: genesis prev_hash invalid", entry.Seq)
			}
			continue
		}

		// Check chain continuity
		prev := entries[i-1]
		if entry.PrevHash != prev.Hash {
			return fmt.Errorf("entry %d: hash chain broken (prev_hash=%s, expected=%s)",
				entry.Seq, entry.PrevHash[:16], prev.Hash[:16])
		}

		// Verify hash
		expected := l.computeHash(entry)
		if entry.Hash != expected {
			return fmt.Errorf("entry %d: hash mismatch (got=%s, expected=%s)",
				entry.Seq, entry.Hash[:16], expected[:16])
		}
	}

	return nil
}

// ReadAll returns all entries in the ledger.
func (l *Ledger) ReadAll() ([]*Entry, error) {
	data, err := os.ReadFile(l.path)
	if err != nil {
		return nil, fmt.Errorf("read ledger file: %w", err)
	}

	var entries []*Entry
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("parse entry: %w", err)
		}
		entries = append(entries, &entry)
	}

	return entries, nil
}

// readLastEntry returns the most recent entry.
func (l *Ledger) readLastEntry() (*Entry, error) {
	entries, err := l.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("ledger is empty")
	}
	return entries[len(entries)-1], nil
}

// writeEntry appends a single entry to the ledger file.
func (l *Ledger) writeEntry(entry *Entry) error {
	file, err := os.OpenFile(l.path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return fmt.Errorf("open ledger: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal entry: %w", err)
	}

	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("write entry: %w", err)
	}

	return nil
}

// redactSecrets recursively redacts sensitive fields from JSON data.
func redactSecrets(data json.RawMessage) (json.RawMessage, []string) {
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		// Not an object, return as-is
		return data, nil
	}

	var redactions []string
	redactValue(obj, "", &redactions)

	out, _ := json.Marshal(obj)
	return out, redactions
}

func redactValue(obj map[string]interface{}, prefix string, redactions *[]string) {
	for key, val := range obj {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		if secretKeys[strings.ToLower(key)] {
			switch val.(type) {
			case string:
				obj[key] = "[REDACTED]"
				*redactions = append(*redactions, fullKey)
			case []interface{}:
				obj[key] = []interface{}{"[REDACTED]"}
				*redactions = append(*redactions, fullKey)
			}
		} else if nested, ok := val.(map[string]interface{}); ok {
			redactValue(nested, fullKey, redactions)
		}
	}
}

// now returns the current time in ISO-8601 UTC format.
func now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z")
}
