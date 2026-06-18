package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// sha256Hash computes a SHA-256 hex digest.
func sha256Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// LedgerIntegration provides helper functions for appending to the ledger
// from other commands (session, dispatch, workflow, etc.)

type LedgerEvent struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	Data      string `json:"data"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prev_hash"`
}

// appendToLedger appends an event to the ledger
func appendToLedger(eventType string, data map[string]string) error {
	ledgerPath := getLedgerPath()

	// Check if ledger exists
	if _, err := os.Stat(ledgerPath); os.IsNotExist(err) {
		// Initialize ledger if it doesn't exist
		if err := initLedger(ledgerPath); err != nil {
			return fmt.Errorf("failed to initialize ledger: %w", err)
		}
	}

	// Read last entry to get prev_hash
	entries, err := readLedgerEntries(ledgerPath)
	if err != nil {
		return err
	}

	var prevHash string
	if len(entries) > 0 {
		prevHash = entries[len(entries)-1].Hash
	}

	// Build data string from map
	dataParts := []string{}
	for k, v := range data {
		dataParts = append(dataParts, fmt.Sprintf("%s=%s", k, v))
	}
	dataStr := strings.Join(dataParts, " ")

	entry := LedgerEvent{
		ID:        fmt.Sprintf("entry_%d", time.Now().UnixNano()),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: eventType,
		Data:      dataStr,
		PrevHash:  prevHash,
	}

	// Calculate hash
	entryData, _ := json.Marshal(map[string]string{
		"id":         entry.ID,
		"timestamp":  entry.Timestamp,
		"event_type": entry.EventType,
		"data":       entry.Data,
		"prev_hash":  entry.PrevHash,
	})
	entry.Hash = sha256Hash(entryData)

	// Append to file
	file, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}
	defer file.Close()

	entryJSON, _ := json.Marshal(entry)
	_, err = file.WriteString(string(entryJSON) + "\n")
	if err != nil {
		return fmt.Errorf("failed to write to ledger: %w", err)
	}

	return nil
}

// initLedger creates a genesis entry
func initLedger(ledgerPath string) error {
	dir := filepath.Dir(ledgerPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	genesis := LedgerEvent{
		ID:        "genesis",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		EventType: "genesis",
		Data:      "Ledger initialized",
		Hash:      "",
		PrevHash:  "",
	}

	genesisData, _ := json.Marshal(map[string]string{
		"id":         genesis.ID,
		"timestamp":  genesis.Timestamp,
		"event_type": genesis.EventType,
		"data":       genesis.Data,
		"prev_hash":  genesis.PrevHash,
	})
	genesis.Hash = sha256Hash(genesisData)

	file, err := os.Create(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to create ledger: %w", err)
	}
	defer file.Close()

	entryJSON, _ := json.Marshal(genesis)
	_, err = file.WriteString(string(entryJSON) + "\n")
	return err
}

// readLedgerEntries reads all entries from the ledger
func readLedgerEntries(path string) ([]LedgerEvent, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger: %w", err)
	}

	var entries []LedgerEvent
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry LedgerEvent
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return entries, nil
}
