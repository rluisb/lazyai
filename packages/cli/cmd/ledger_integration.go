package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime/ledger"
)

// LedgerIntegration provides helper functions for appending to the ledger
// from other commands (session, dispatch, workflow, etc.)

// LedgerEvent is the legacy display format for ledger entries.
// It is kept for callers that read the ledger for display purposes.
type LedgerEvent struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	EventType string `json:"event_type"`
	Data      string `json:"data"`
	Hash      string `json:"hash"`
	PrevHash  string `json:"prev_hash"`
}

// appendToLedger appends an event to the runtime-backed ledger.
func appendToLedger(eventType string, data map[string]string) error {
	ledgerPath := getLedgerPath()

	l, err := ledger.Open(ledgerPath)
	if err != nil {
		return fmt.Errorf("failed to open ledger: %w", err)
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshal ledger data: %w", err)
	}

	entry := &ledger.Entry{
		Type: eventType,
		Data: dataJSON,
	}

	if err := l.Append(entry); err != nil {
		return fmt.Errorf("failed to append to ledger: %w", err)
	}

	return nil
}

// initLedger creates a genesis entry using the runtime ledger.
func initLedger(ledgerPath string) error {
	_, err := ledger.Open(ledgerPath)
	return err
}

// readLedgerEntries reads all entries from the ledger and converts them to
// the legacy display format.
func readLedgerEntries(path string) ([]LedgerEvent, error) {
	l, err := ledger.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open ledger: %w", err)
	}

	entries, err := l.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read ledger: %w", err)
	}

	var result []LedgerEvent
	for _, e := range entries {
		dataStr := string(e.Data)
		var dataMap map[string]string
		if err := json.Unmarshal(e.Data, &dataMap); err == nil {
			if d, ok := dataMap["data"]; ok {
				dataStr = d
			}
		}
		result = append(result, LedgerEvent{
			ID:        fmt.Sprintf("entry_%d", e.Seq),
			Timestamp: e.Timestamp,
			EventType: e.Type,
			Data:      dataStr,
			Hash:      e.Hash,
			PrevHash:  e.PrevHash,
		})
	}

	return result, nil
}

// ledgerExists returns true if the ledger file exists at the expected path.
func ledgerExists() bool {
	_, err := os.Stat(getLedgerPath())
	return err == nil
}
