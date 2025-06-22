package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompletionCommand(t *testing.T) {
	tests := []struct {
		name     string
		shell    string
		wantErr  bool
		contains []string
	}{
		{
			name:    "bash completion",
			shell:   "bash",
			wantErr: false,
			contains: []string{
				"complete",
				"apidirect",
			},
		},
		{
			name:    "zsh completion",
			shell:   "zsh",
			wantErr: false,
			contains: []string{
				"compdef",
				"apidirect",
			},
		},
		{
			name:    "fish completion",
			shell:   "fish",
			wantErr: false,
			contains: []string{
				"complete",
				"apidirect",
			},
		},
		{
			name:    "powershell completion",
			shell:   "powershell",
			wantErr: false,
			contains: []string{
				"Register-ArgumentCompleter",
				"apidirect",
			},
		},
		{
			name:    "invalid shell",
			shell:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := rootCmd
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			
			args := []string{"completion"}
			if tt.shell != "" {
				args = append(args, tt.shell)
			}
			
			cmd.SetArgs(args)
			err := cmd.Execute()
			
			if (err != nil) != tt.wantErr {
				t.Errorf("completion command error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				output := buf.String()
				for _, want := range tt.contains {
					if !strings.Contains(output, want) {
						t.Errorf("completion output missing %q", want)
					}
				}
			}
		})
	}
}

func TestCompletionCommandHelp(t *testing.T) {
	cmd := rootCmd
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	
	cmd.SetArgs([]string{"completion", "--help"})
	err := cmd.Execute()
	
	if err != nil {
		t.Fatalf("completion help command failed: %v", err)
	}
	
	output := buf.String()
	expectedStrings := []string{
		"Generate shell completion script",
		"bash",
		"zsh",
		"fish",
		"powershell",
		"To load completions",
	}
	
	for _, expected := range expectedStrings {
		if !strings.Contains(output, expected) {
			t.Errorf("completion help missing expected string: %q", expected)
		}
	}
}