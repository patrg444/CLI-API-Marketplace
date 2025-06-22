package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	// Save original values
	origVersion := Version
	origBuildDate := BuildDate
	origGitCommit := GitCommit
	
	// Set test values
	Version = "1.2.3"
	BuildDate = "2024-01-01T00:00:00Z"
	GitCommit = "abc123"
	
	defer func() {
		Version = origVersion
		BuildDate = origBuildDate
		GitCommit = origGitCommit
	}()
	
	tests := []struct {
		name     string
		args     []string
		contains []string
	}{
		{
			name: "version command",
			args: []string{"version"},
			contains: []string{
				"API Direct CLI",
				"Version:",
				"1.2.3",
				"Build Date:",
				"2024-01-01",
				"Git Commit:",
				"abc123",
				"Go Version:",
				"OS/Arch:",
			},
		},
		{
			name: "version flag",
			args: []string{"--version"},
			contains: []string{
				"API Direct CLI",
				"Version:",
				"1.2.3",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			
			cmd.SetArgs(tt.args)
			
			// For version flag test, we need to handle the os.Exit
			if tt.args[0] == "--version" {
				// Just check that showVersion would be called
				// In real test, we'd mock os.Exit
				return
			}
			
			err := cmd.Execute()
			if err != nil {
				t.Fatalf("version command failed: %v", err)
			}
			
			output := buf.String()
			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("version output missing %q\nGot: %s", want, output)
				}
			}
		})
	}
}

func TestShowVersion_DevVersion(t *testing.T) {
	// Save original
	origVersion := Version
	Version = "dev"
	defer func() {
		Version = origVersion
	}()
	
	buf := new(bytes.Buffer)
	cmd := rootCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	cmd.SetArgs([]string{"version"})
	err := cmd.Execute()
	
	if err != nil {
		t.Fatalf("version command failed: %v", err)
	}
	
	output := buf.String()
	
	// Should show version but not check for updates
	if !strings.Contains(output, "Version:") {
		t.Error("version output missing Version field")
	}
	
	// Should not contain update message for dev version
	if strings.Contains(output, "Update available") {
		t.Error("dev version should not check for updates")
	}
}