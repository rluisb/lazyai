package cmd

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/spf13/cobra"
)

var backupCmd = &cobra.Command{
	Use:     "backup",
	Short:   "Backup LazyAI data",
	Long:    `Create a backup of database, ledger, and configuration.`,
	GroupID: "safety",
}

var backupCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a backup",
	Long:  `Create a tarball backup of all LazyAI data.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		if output == "" {
			timestamp := time.Now().Format("20060102_150405")
			output = fmt.Sprintf("lazyai-backup-%s.tar.gz", timestamp)
		}

		fmt.Printf("🗄️  Creating backup: %s\n", output)

		// Create tarball
		file, err := os.Create(output)
		if err != nil {
			return fmt.Errorf("error creating backup file: %w", err)
		}
		defer file.Close()

		gzipWriter := gzip.NewWriter(file)
		defer gzipWriter.Close()

		tarWriter := tar.NewWriter(gzipWriter)
		defer tarWriter.Close()

		// Files to backup
		filesToBackup := []string{
			".ai-setup.db",
			".specify/ledger.jsonl",
			".opencode/config.yaml",
		}

		backedUp := 0
		for _, filePath := range filesToBackup {
			if err := addFileToTar(tarWriter, filePath); err != nil {
				fmt.Printf("  ⚠️  Skipping %s: %v\n", filePath, err)
				continue
			}
			fmt.Printf("  ✅ %s\n", filePath)
			backedUp++
		}

		// Backup memory files if they exist
		memoryDir := ".specify/memory"
		if entries, err := os.ReadDir(memoryDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					memoryPath := filepath.Join(memoryDir, entry.Name())
					if err := addFileToTar(tarWriter, memoryPath); err == nil {
						fmt.Printf("  ✅ %s\n", memoryPath)
						backedUp++
					}
				}
			}
		}

		fmt.Printf("\n✅ Backup complete: %s (%d files)\n", output, backedUp)

		// Record to ledger
		appendToLedger("backup_created", map[string]string{
			"file":  output,
			"files": fmt.Sprintf("%d", backedUp),
		})

		return nil
	},
}

var backupRestoreCmd = &cobra.Command{
	Use:   "restore [backup-file]",
	Short: "Restore from backup",
	Long:  `Restore LazyAI data from a backup tarball. WARNING: This will overwrite current data.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		backupFile := args[0]
		force, _ := cmd.Flags().GetBool("force")

		// Safety confirmation
		if !ConfirmAction(fmt.Sprintf("⚠️  This will overwrite current data with backup from %s. Continue?", backupFile), force) {
			fmt.Println("❌ Restore cancelled.")
			return nil
		}

		fmt.Printf("🗄️  Restoring from: %s\n", backupFile)

		file, err := os.Open(backupFile)
		if err != nil {
			return fmt.Errorf("error opening backup file: %w", err)
		}
		defer file.Close()

		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("error reading gzip: %w", err)
		}
		defer gzipReader.Close()

		tarReader := tar.NewReader(gzipReader)

		restoreRoot, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("error resolving restore directory: %w", err)
		}

		restored := 0
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("error reading tar: %w", err)
			}

			// Reject archive entries that would escape the restore root
			// (path traversal via "..", absolute paths, etc.) before any write.
			destPath, err := safeRestorePath(restoreRoot, header.Name)
			if err != nil {
				fmt.Printf("  ⚠️  Skipping unsafe entry %s: %v\n", header.Name, err)
				continue
			}

			// Create parent directory if needed
			dir := filepath.Dir(destPath)
			if err := os.MkdirAll(dir, 0o755); err != nil {
				fmt.Printf("  ⚠️  Error creating directory %s: %v\n", dir, err)
				continue
			}

			// Write file
			outFile, err := os.Create(destPath)
			if err != nil {
				fmt.Printf("  ⚠️  Error creating file %s: %v\n", header.Name, err)
				continue
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				fmt.Printf("  ⚠️  Error writing file %s: %v\n", header.Name, err)
				continue
			}
			outFile.Close()

			fmt.Printf("  ✅ %s\n", header.Name)
			restored++
		}

		fmt.Printf("\n✅ Restore complete: %d files restored\n", restored)

		// Record to ledger
		appendToLedger("backup_restored", map[string]string{
			"file":  backupFile,
			"files": fmt.Sprintf("%d", restored),
		})

		return nil
	},
}

// addFileToTar adds a file to a tar archive
func addFileToTar(tw *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    filePath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	if _, err := io.Copy(tw, file); err != nil {
		return err
	}

	return nil
}

// safeRestorePath joins a tar entry name under restoreRoot and rejects any
// entry that would escape that root via absolute paths or ".." traversal.
func safeRestorePath(restoreRoot, name string) (string, error) {
	return files.SafeJoin(restoreRoot, name)
}

func init() {
	backupCreateCmd.Flags().StringP("output", "o", "", "Output file path (default: lazyai-backup-YYYYMMDD_HHMMSS.tar.gz)")
	backupRestoreCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	backupCmd.AddCommand(backupCreateCmd)
	backupCmd.AddCommand(backupRestoreCmd)
	rootCmd.AddCommand(backupCmd)
}
