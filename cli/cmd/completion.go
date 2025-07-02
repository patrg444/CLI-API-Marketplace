package cmd

import (
	"github.com/spf13/cobra"
)

// completionCmd represents the completion command
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for API Direct CLI.

To load completions:

Bash:
  $ source <(apidirect completion bash)
  # To load completions for each session, execute once:
  # Linux:
  $ apidirect completion bash > /etc/bash_completion.d/apidirect
  # macOS:
  $ apidirect completion bash > /usr/local/etc/bash_completion.d/apidirect

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ apidirect completion zsh > "${fpath[1]}/_apidirect"
  # You will need to start a new shell for this setup to take effect.

Fish:
  $ apidirect completion fish | source
  # To load completions for each session, execute once:
  $ apidirect completion fish > ~/.config/fish/completions/apidirect.fish

PowerShell:
  PS> apidirect completion powershell | Out-String | Invoke-Expression
  # To load completions for every new session, run:
  PS> apidirect completion powershell > apidirect.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return cmd.Root().GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}