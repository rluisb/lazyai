package cmd

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const latestReleaseURL = "https://api.github.com/repos/rluisb/lazyai/releases/latest"

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

var updateSelfCmd = &cobra.Command{
	Use:   "update-self",
	Short: "Update lazyai-cli to the latest GitHub Release",
	Long:  "Download and replace the current lazyai-cli binary with the latest version from GitHub Releases.",
	RunE:  runUpdateSelf,
}

func init() {
	updateSelfCmd.Flags().Bool("check", false, "Only check if update available")
	updateSelfCmd.Flags().Bool("force", false, "Skip version check")
	updateSelfCmd.Flags().Bool("dry-run", false, "Preview without downloading")
	updateSelfCmd.Flags().Bool("verbose", false, "Show detailed output")
	rootCmd.AddCommand(updateSelfCmd)
	updateSelfCmd.GroupID = "lifecycle"
}

func runUpdateSelf(cmd *cobra.Command, args []string) error {
	checkOnly, _ := cmd.Flags().GetBool("check")
	force, _ := cmd.Flags().GetBool("force")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	verbose, _ := cmd.Flags().GetBool("verbose")

	if Version == "0.0.0-dev" {
		fmt.Println("Development build — skipping self-update")
		return nil
	}

	currentVersion := normalizeVersion(Version)
	client := http.DefaultClient
	token := os.Getenv("GITHUB_TOKEN")

	if verbose {
		fmt.Printf("Current version: %s\n", displayVersion(currentVersion))
		fmt.Printf("Fetching latest release from %s\n", latestReleaseURL)
	}

	release, err := fetchLatestRelease(client, token)
	if err != nil {
		return err
	}
	if release.TagName == "" {
		return fmt.Errorf("latest release did not include a tag_name")
	}

	latestVersion := normalizeVersion(release.TagName)
	if verbose {
		fmt.Printf("Latest version: %s\n", displayVersion(latestVersion))
	}

	updateAvailable := currentVersion < latestVersion
	if checkOnly {
		if updateAvailable {
			fmt.Printf("%s available (current: %s)\n", displayVersion(latestVersion), Version)
			os.Exit(1)
		}
		fmt.Printf("Already up to date (%s)\n", displayVersion(currentVersion))
		return nil
	}

	if !force && !updateAvailable {
		fmt.Printf("Already up to date (%s)\n", displayVersion(currentVersion))
		return nil
	}

	assetName := binaryAssetName(runtime.GOOS, runtime.GOARCH)
	binaryAsset, ok := findReleaseAsset(release.Assets, assetName)
	if !ok {
		return fmt.Errorf("release %s does not include asset %q", displayVersion(latestVersion), assetName)
	}

	checksumsAsset, ok := findReleaseAsset(release.Assets, "checksums.txt")
	if !ok {
		return fmt.Errorf("release %s does not include checksums.txt", displayVersion(latestVersion))
	}

	if dryRun {
		fmt.Printf("Would download lazyai-cli %s for %s/%s\n", displayVersion(latestVersion), runtime.GOOS, runtime.GOARCH)
		if verbose {
			fmt.Printf("Binary asset: %s\n", binaryAsset.BrowserDownloadURL)
			fmt.Printf("Checksums asset: %s\n", checksumsAsset.BrowserDownloadURL)
		}
		return nil
	}

	fmt.Printf("Downloading lazyai-cli %s for %s/%s...\n", displayVersion(latestVersion), runtime.GOOS, runtime.GOARCH)

	tempDir, err := os.MkdirTemp("", "lazyai-cli-update-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, assetName)
	if err := downloadFile(client, token, binaryAsset.BrowserDownloadURL, binaryPath); err != nil {
		return fmt.Errorf("downloading binary: %w", err)
	}

	checksumsPath := filepath.Join(tempDir, "checksums.txt")
	if err := downloadFile(client, token, checksumsAsset.BrowserDownloadURL, checksumsPath); err != nil {
		return fmt.Errorf("downloading checksums: %w", err)
	}

	if err := verifyChecksum(binaryPath, checksumsPath, assetName); err != nil {
		_ = os.Remove(binaryPath)
		return err
	}
	if verbose {
		fmt.Println("Checksum verified")
	}

	if err := replaceRunningBinary(binaryPath, verbose); err != nil {
		return err
	}

	fmt.Printf("✓ Updated to %s\n", displayVersion(latestVersion))
	return nil
}

func fetchLatestRelease(client *http.Client, token string) (*githubRelease, error) {
	req, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating release request: %w", err)
	}
	addGitHubHeaders(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("fetching latest release: GitHub returned %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("parsing latest release: %w", err)
	}
	return &release, nil
}

func addGitHubHeaders(req *http.Request, token string) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func normalizeVersion(version string) string {
	return strings.TrimPrefix(version, "v")
}

func displayVersion(version string) string {
	return "v" + normalizeVersion(version)
}

func binaryAssetName(goos, goarch string) string {
	name := fmt.Sprintf("lazyai-cli-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

func findReleaseAsset(assets []githubAsset, name string) (githubAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return githubAsset{}, false
}

func downloadFile(client *http.Client, token, url, destination string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	addGitHubHeaders(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	out, err := os.Create(destination)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func verifyChecksum(binaryPath, checksumsPath, assetName string) error {
	expectedChecksum, err := checksumForAsset(checksumsPath, assetName)
	if err != nil {
		return err
	}

	actualChecksum, err := sha256ForFile(binaryPath)
	if err != nil {
		return fmt.Errorf("computing checksum: %w", err)
	}

	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("checksum mismatch for %s: expected %s, got %s", assetName, expectedChecksum, actualChecksum)
	}
	return nil
}

func checksumForAsset(checksumsPath, assetName string) (string, error) {
	contents, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("reading checksums: %w", err)
	}

	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}

		filename := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filepath.Base(filename) == assetName {
			return fields[0], nil
		}
	}

	return "", fmt.Errorf("checksums.txt does not include %s", assetName)
}

func sha256ForFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func replaceRunningBinary(newBinaryPath string, verbose bool) error {
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating current executable: %w", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(executablePath)
	if err == nil {
		executablePath = resolvedPath
	} else if verbose {
		fmt.Printf("Could not resolve executable symlink %s: %v\n", executablePath, err)
	}

	if runtime.GOOS == "windows" {
		newPath := executablePath + ".new"
		if err := copyFile(newBinaryPath, newPath, 0755); err != nil {
			return fmt.Errorf("writing replacement binary: %w", err)
		}
		fmt.Printf("Windows cannot replace a running executable. New binary written to %s\n", newPath)
		fmt.Printf("After exiting, replace %s with %s\n", executablePath, newPath)
		return nil
	}

	dir := filepath.Dir(executablePath)
	base := filepath.Base(executablePath)
	stagedFile, err := os.CreateTemp(dir, "."+base+".update-*")
	if err != nil {
		return fmt.Errorf("creating staged binary: %w", err)
	}
	stagedPath := stagedFile.Name()
	if err := stagedFile.Close(); err != nil {
		_ = os.Remove(stagedPath)
		return fmt.Errorf("closing staged binary: %w", err)
	}
	defer os.Remove(stagedPath)

	if err := copyFile(newBinaryPath, stagedPath, 0755); err != nil {
		return fmt.Errorf("staging replacement binary: %w", err)
	}
	if verbose {
		fmt.Printf("Replacing %s\n", executablePath)
	}

	if err := os.Rename(stagedPath, executablePath); err != nil {
		return fmt.Errorf("replacing current binary: %w", err)
	}
	return nil
}

func copyFile(source, destination string, mode os.FileMode) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}

	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Chmod(destination, mode)
}
