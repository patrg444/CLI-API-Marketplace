package cmd

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Ensure HOME is set for tests
	originalHome := os.Getenv("HOME")
	if originalHome == "" {
		os.Setenv("HOME", "/tmp")
	}
	
	// Mock browser opener to prevent opening browsers during tests
	originalBrowserOpener := browserOpener
	browserOpener = func(url string) error {
		// Do nothing during tests
		return nil
	}
	
	// Run tests
	code := m.Run()
	
	// Restore original values after tests
	browserOpener = originalBrowserOpener
	if originalHome == "" {
		os.Unsetenv("HOME")
	} else {
		os.Setenv("HOME", originalHome)
	}
	
	// Exit with the test result code
	os.Exit(code)
}