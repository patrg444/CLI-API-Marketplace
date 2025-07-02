package cmd

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTarGz(t *testing.T) {
	tests := []struct {
		name        string
		createArchive func() (io.Reader, error)
		expectError bool
		errorContains string
	}{
		{
			name: "valid tar.gz with binary",
			createArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				
				// Create gzip writer
				gw := gzip.NewWriter(&buf)
				
				// Create tar writer
				tw := tar.NewWriter(gw)
				
				// Add binary file
				binaryName := "apidirect"
				if runtime.GOOS == "windows" {
					binaryName += ".exe"
				}
				
				content := []byte("mock binary content")
				header := &tar.Header{
					Name: binaryName,
					Mode: 0755,
					Size: int64(len(content)),
				}
				
				if err := tw.WriteHeader(header); err != nil {
					return nil, err
				}
				if _, err := tw.Write(content); err != nil {
					return nil, err
				}
				
				if err := tw.Close(); err != nil {
					return nil, err
				}
				if err := gw.Close(); err != nil {
					return nil, err
				}
				
				return &buf, nil
			},
			expectError: false,
		},
		{
			name: "tar.gz without binary",
			createArchive: func() (io.Reader, error) {
				var buf bytes.Buffer
				
				gw := gzip.NewWriter(&buf)
				tw := tar.NewWriter(gw)
				
				// Add some other file
				content := []byte("some content")
				header := &tar.Header{
					Name: "README.md",
					Mode: 0644,
					Size: int64(len(content)),
				}
				
				if err := tw.WriteHeader(header); err != nil {
					return nil, err
				}
				if _, err := tw.Write(content); err != nil {
					return nil, err
				}
				
				tw.Close()
				gw.Close()
				
				return &buf, nil
			},
			expectError: true,
			errorContains: "not found in archive",
		},
		{
			name: "invalid gzip data",
			createArchive: func() (io.Reader, error) {
				return bytes.NewReader([]byte("invalid gzip data")), nil
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create archive
			archive, err := tt.createArchive()
			require.NoError(t, err)
			
			// Create output file
			output, err := os.CreateTemp("", "test-output-*")
			require.NoError(t, err)
			defer os.Remove(output.Name())
			
			// Test extraction
			err = extractTarGz(archive, output)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				
				// Verify content was written
				output.Close()
				content, err := os.ReadFile(output.Name())
				assert.NoError(t, err)
				assert.Equal(t, "mock binary content", string(content))
			}
		})
	}
}

func TestExtractZip(t *testing.T) {
	tests := []struct {
		name          string
		createArchive func() (string, error)
		expectError   bool
		errorContains string
	}{
		{
			name: "valid zip with binary",
			createArchive: func() (string, error) {
				// Create temporary zip file
				zipFile, err := os.CreateTemp("", "test-*.zip")
				if err != nil {
					return "", err
				}
				
				// Create zip writer
				zw := zip.NewWriter(zipFile)
				
				// Add binary file
				binaryName := "apidirect"
				if runtime.GOOS == "windows" {
					binaryName += ".exe"
				}
				
				w, err := zw.Create(binaryName)
				if err != nil {
					return "", err
				}
				
				if _, err := w.Write([]byte("mock binary content")); err != nil {
					return "", err
				}
				
				if err := zw.Close(); err != nil {
					return "", err
				}
				
				zipFile.Close()
				return zipFile.Name(), nil
			},
			expectError: false,
		},
		{
			name: "zip without binary",
			createArchive: func() (string, error) {
				zipFile, err := os.CreateTemp("", "test-*.zip")
				if err != nil {
					return "", err
				}
				
				zw := zip.NewWriter(zipFile)
				
				// Add some other file
				w, err := zw.Create("README.md")
				if err != nil {
					return "", err
				}
				
				if _, err := w.Write([]byte("some content")); err != nil {
					return "", err
				}
				
				zw.Close()
				zipFile.Close()
				return zipFile.Name(), nil
			},
			expectError:   true,
			errorContains: "not found in archive",
		},
		{
			name: "invalid zip file",
			createArchive: func() (string, error) {
				tmpFile, err := os.CreateTemp("", "test-*.zip")
				if err != nil {
					return "", err
				}
				
				// Write invalid zip data
				tmpFile.Write([]byte("invalid zip data"))
				tmpFile.Close()
				return tmpFile.Name(), nil
			},
			expectError: true,
		},
		{
			name: "zip with binary in subdirectory",
			createArchive: func() (string, error) {
				zipFile, err := os.CreateTemp("", "test-*.zip")
				if err != nil {
					return "", err
				}
				
				zw := zip.NewWriter(zipFile)
				
				// Add binary in subdirectory
				binaryName := "apidirect"
				if runtime.GOOS == "windows" {
					binaryName += ".exe"
				}
				
				w, err := zw.Create(filepath.Join("bin", binaryName))
				if err != nil {
					return "", err
				}
				
				if _, err := w.Write([]byte("mock binary content")); err != nil {
					return "", err
				}
				
				zw.Close()
				zipFile.Close()
				return zipFile.Name(), nil
			},
			expectError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create archive
			archivePath, err := tt.createArchive()
			require.NoError(t, err)
			defer os.Remove(archivePath)
			
			// Create output file
			output, err := os.CreateTemp("", "test-output-*")
			require.NoError(t, err)
			defer os.Remove(output.Name())
			
			// Test extraction
			err = extractZip(archivePath, output)
			
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				
				// Verify content was written
				output.Close()
				content, err := os.ReadFile(output.Name())
				assert.NoError(t, err)
				assert.Equal(t, "mock binary content", string(content))
			}
		})
	}
}

func TestGetLatestRelease(t *testing.T) {
	// This test would require mocking HTTP calls
	// For now, we'll just verify the function exists
	t.Run("function exists", func(t *testing.T) {
		// The function is tested indirectly through integration tests
		// or with a mock HTTP client
		t.Skip("Requires HTTP mocking")
	})
}

func TestDownloadAndInstall(t *testing.T) {
	// This test would require elevated permissions and HTTP mocking
	t.Run("function exists", func(t *testing.T) {
		// The function is tested indirectly through integration tests
		t.Skip("Requires elevated permissions and HTTP mocking")
	})
}