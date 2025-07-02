package cmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	updateForce bool
	updateCheck bool
)

// selfUpdateCmd represents the self-update command
var selfUpdateCmd = &cobra.Command{
	Use:   "self-update",
	Short: "Update API Direct CLI to the latest version",
	Long: `Check for updates and automatically update the API Direct CLI to the latest version.

This command will:
1. Check the current version
2. Query GitHub for the latest release
3. Download and install the update if available
4. Verify the installation

Examples:
  apidirect self-update
  apidirect self-update --check
  apidirect self-update --force`,
	RunE: runSelfUpdate,
}

func init() {
	rootCmd.AddCommand(selfUpdateCmd)
	
	selfUpdateCmd.Flags().BoolVar(&updateForce, "force", false, "Force update even if already on latest version")
	selfUpdateCmd.Flags().BoolVar(&updateCheck, "check", false, "Only check for updates without installing")
}

func runSelfUpdate(cmd *cobra.Command, args []string) error {
	currentVersion := Version
	if currentVersion == "" || currentVersion == "dev" {
		currentVersion = "0.0.0"
	}
	
	// Get latest release info
	latestVersion, downloadURL, err := getLatestRelease()
	if err != nil {
		return fmt.Errorf("Failed to check for updates: %w", err)
	}
	
	fmt.Printf("Current version: %s\n", color.YellowString(currentVersion))
	fmt.Printf("Latest version: %s\n", color.GreenString(latestVersion))
	
	if updateCheck {
		if currentVersion == latestVersion {
			fmt.Println("\n✅ You are on the latest version!")
		} else {
			fmt.Printf("\n🆕 Update available! Run 'apidirect self-update' to install.\n")
		}
		return nil
	}
	
	// Check if update needed
	if currentVersion == latestVersion && !updateForce {
		fmt.Println("\n✅ You are already on the latest version!")
		return nil
	}
	
	if !updateForce {
		fmt.Printf("\n🔄 Update available: %s → %s\n", currentVersion, latestVersion)
		if !confirmAction("Do you want to update?") {
			fmt.Println("Update cancelled.")
			return nil
		}
	}
	
	// Download and install update
	fmt.Println("\n📥 Downloading update...")
	if err := downloadAndInstall(downloadURL); err != nil {
		return fmt.Errorf("Failed to install update: %w", err)
	}
	
	fmt.Println(color.GreenString("\n✅ Update completed successfully!"))
	fmt.Println("Please restart your terminal or run 'apidirect --version' to verify.")
	
	return nil
}

func getLatestRelease() (version, downloadURL string, err error) {
	resp, err := http.Get("https://api.github.com/repos/api-direct/cli/releases/latest")
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}
	
	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", "", err
	}
	
	version = strings.TrimPrefix(release.TagName, "v")
	
	// Find appropriate asset for current platform
	osName := runtime.GOOS
	archName := runtime.GOARCH
	if archName == "amd64" {
		archName = "x86_64"
	}
	
	expectedName := fmt.Sprintf("apidirect_%s_%s_%s", version, osName, archName)
	
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, expectedName) {
			downloadURL = asset.BrowserDownloadURL
			break
		}
	}
	
	if downloadURL == "" {
		return "", "", fmt.Errorf("No compatible binary found for %s/%s", osName, archName)
	}
	
	return version, downloadURL, nil
}

func downloadAndInstall(url string) error {
	// Create temporary file
	tmpFile, err := os.CreateTemp("", "apidirect-update-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	
	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Download failed with status %d", resp.StatusCode)
	}
	
	// Handle compressed files
	if strings.HasSuffix(url, ".tar.gz") {
		if err := extractTarGz(resp.Body, tmpFile); err != nil {
			return fmt.Errorf("Failed to extract tar.gz: %w", err)
		}
	} else if strings.HasSuffix(url, ".zip") {
		// For zip files, we need to download to a temporary file first
		zipFile, err := os.CreateTemp("", "apidirect-*.zip")
		if err != nil {
			return err
		}
		defer os.Remove(zipFile.Name())
		
		if _, err := io.Copy(zipFile, resp.Body); err != nil {
			return err
		}
		zipFile.Close()
		
		if err := extractZip(zipFile.Name(), tmpFile); err != nil {
			return fmt.Errorf("Failed to extract zip: %w", err)
		}
	} else {
		// Direct binary download
		if _, err := io.Copy(tmpFile, resp.Body); err != nil {
			return err
		}
	}
	tmpFile.Close()
	
	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		return err
	}
	
	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	
	// On Windows, we need to rename the old executable first
	if runtime.GOOS == "windows" {
		backupPath := execPath + ".backup"
		if err := os.Rename(execPath, backupPath); err != nil {
			return fmt.Errorf("Failed to backup current executable: %w", err)
		}
		defer os.Remove(backupPath)
	}
	
	// Replace current executable
	if err := os.Rename(tmpFile.Name(), execPath); err != nil {
		// Try with elevated permissions
		fmt.Println("⚠️  Update requires elevated permissions. You may be prompted for your password.")
		
		// Use sudo on Unix-like systems
		if runtime.GOOS != "windows" {
			if err := runElevated("mv", tmpFile.Name(), execPath); err != nil {
				return fmt.Errorf("Failed to install update with elevated permissions: %w", err)
			}
		} else {
			return fmt.Errorf("Please run this command as Administrator")
		}
	}
	
	return nil
}

func runElevated(command string, args ...string) error {
	// This is a simplified version - in production, use proper elevation
	fmt.Printf("sudo %s %s\n", command, strings.Join(args, " "))
	return fmt.Errorf("Automatic elevation not implemented - please run manually with sudo")
}

// extractTarGz extracts a tar.gz archive and finds the binary
func extractTarGz(r io.Reader, output *os.File) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()
	
	tr := tar.NewReader(gzr)
	
	// Look for the binary in the archive
	binaryName := "apidirect"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		
		// Check if this is our binary
		if filepath.Base(header.Name) == binaryName {
			// Copy the binary to our output file
			if _, err := io.Copy(output, tr); err != nil {
				return err
			}
			return nil
		}
	}
	
	return fmt.Errorf("Binary %s not found in archive", binaryName)
}

// extractZip extracts a zip archive and finds the binary
func extractZip(zipPath string, output *os.File) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	
	// Look for the binary in the archive
	binaryName := "apidirect"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	
	for _, f := range r.File {
		if filepath.Base(f.Name) == binaryName {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()
			
			// Copy the binary to our output file
			if _, err := io.Copy(output, rc); err != nil {
				return err
			}
			return nil
		}
	}
	
	return fmt.Errorf("Binary %s not found in archive", binaryName)
}