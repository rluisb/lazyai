package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var restoreRuntimeDBCmd = &cobra.Command{
	Use:   "restore-runtime-db [backup-path]",
	Short: "Restore runtime database from backup",
	Long: `Restore .specify/session.db from a backup file created before schema migration.

This is the rollback path for Phase 3 V2 schema migration. Copies the backup
SQLite database over the current runtime database. The current database is
renamed to .specify/session.db.pre-restore before the copy.

WARNING: This overwrites the current runtime database. All data written since
the backup was created will be lost.`,
	Args: cobra.ExactArgs(1),
	RunE: runRestoreRuntimeDB,
}

func init() {
	restoreRuntimeDBCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	rootCmd.AddCommand(restoreRuntimeDBCmd)
	restoreRuntimeDBCmd.GroupID = "lifecycle"
}

func runRestoreRuntimeDB(cmd *cobra.Command, args []string) error {
	backupPath := args[0]
	force, _ := cmd.Flags().GetBool("force")

	// Resolve backup path
	absBackup, err := filepath.Abs(backupPath)
	if err != nil {
		return fmt.Errorf("resolve backup path: %w", err)
	}

	if _, err := os.Stat(absBackup); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", absBackup)
	}

	dbPath := getRuntimeDBPath()
	absDB, err := filepath.Abs(dbPath)
	if err != nil {
		return fmt.Errorf("resolve DB path: %w", err)
	}

	// Safety confirmation
	if !ConfirmAction(fmt.Sprintf(
		"⚠️  This will replace %s with %s. All data since backup will be lost. Continue?",
		absDB, absBackup), force) {
		fmt.Println("Restore cancelled.")
		return nil
	}

	// If current DB exists, rename it as pre-restore backup
	if _, err := os.Stat(absDB); err == nil {
		preRestorePath := absDB + ".pre-restore"
		if err := os.Rename(absDB, preRestorePath); err != nil {
			return fmt.Errorf("rename current DB: %w", err)
		}
		fmt.Printf("  Current DB saved to %s\n", preRestorePath)
	}

	// Ensure .specify directory exists
	if err := os.MkdirAll(filepath.Dir(absDB), 0755); err != nil {
		return fmt.Errorf("create .specify directory: %w", err)
	}

	// Copy backup to session.db
	src, err := os.Open(absBackup)
	if err != nil {
		return fmt.Errorf("open backup: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(absDB)
	if err != nil {
		return fmt.Errorf("create DB file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		// Clean up partial copy
		os.Remove(absDB)
		return fmt.Errorf("copy backup: %w", err)
	}

	fmt.Printf("✓ Runtime database restored from %s\n", absBackup)

	// Record to ledger
	appendToLedger("runtime_db_restored", map[string]string{
		"backup": absBackup,
	})

	return nil
}
