package cmd

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

// Version info (set during build)
var (
	Version   = "dev"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display version information about the API Direct CLI including build details.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersionTo(cmd.OutOrStdout())
	},
}

// Alternative --version flag
var versionFlag bool

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
}

func showVersion() {
	showVersionTo(os.Stdout)
}

func showVersionTo(w io.Writer) {
	fmt.Fprintln(w, color.CyanString("API Direct CLI"))
	fmt.Fprintf(w, "Version:      %s\n", color.GreenString(Version))
	fmt.Fprintf(w, "Build Date:   %s\n", BuildDate)
	fmt.Fprintf(w, "Git Commit:   %s\n", GitCommit)
	fmt.Fprintf(w, "Go Version:   %s\n", runtime.Version())
	fmt.Fprintf(w, "OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	// Check for updates
	if Version != "dev" {
		checkForUpdatesTo(w)
	}
}

func checkForUpdates() {
	checkForUpdatesTo(os.Stdout)
}

func checkForUpdatesTo(w io.Writer) {
	latestVersion, _, err := getLatestRelease()
	if err == nil && latestVersion != Version {
		fmt.Fprintf(w, "\n%s Update available: %s → %s\n", 
			color.YellowString("🆕"),
			Version, 
			color.GreenString(latestVersion))
		fmt.Fprintln(w, "Run 'apidirect self-update' to update.")
	}
}