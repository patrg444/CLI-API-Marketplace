package cmd

import (
	"fmt"
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
		showVersion()
	},
}

// Alternative --version flag
var versionFlag bool

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Show version information")
}

func showVersion() {
	fmt.Println(color.CyanString("API Direct CLI"))
	fmt.Printf("Version:      %s\n", color.GreenString(Version))
	fmt.Printf("Build Date:   %s\n", BuildDate)
	fmt.Printf("Git Commit:   %s\n", GitCommit)
	fmt.Printf("Go Version:   %s\n", runtime.Version())
	fmt.Printf("OS/Arch:      %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	// Check for updates
	if Version != "dev" {
		checkForUpdates()
	}
}

func checkForUpdates() {
	latestVersion, _, err := getLatestRelease()
	if err == nil && latestVersion != Version {
		fmt.Printf("\n%s Update available: %s → %s\n", 
			color.YellowString("🆕"),
			Version, 
			color.GreenString(latestVersion))
		fmt.Println("Run 'apidirect self-update' to update.")
	}
}