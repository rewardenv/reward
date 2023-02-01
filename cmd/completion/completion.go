package completion

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	cmdpkg "github.com/rewardenv/reward/cmd"
	"github.com/rewardenv/reward/internal/config"
)

func NewCompletionCmd(c *config.Config) *cmdpkg.Command {
	return &cmdpkg.Command{
		Command: &cobra.Command{
			Use:   "completion [bash|zsh|fish|powershell]",
			Short: "Generate completion script",
			Long: fmt.Sprintf(
				`To load completions:

Bash:

$ source <(%[1]v completion bash)

# To load completions for each session, execute once:
Linux:
  $ %[1]v completion bash | sudo tee -a /etc/bash_completion.d/%[1]v
MacOS:
  $ %[1]v completion bash > /usr/local/etc/bash_completion.d/%[1]v

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ %[1]v completion zsh > "${fpath[1]}/_%[1]v"

# You will need to start a new shell for this setup to take effect.

Fish:

$ %[1]v completion fish | source

# To load completions for each session, execute once:
$ %[1]v completion fish > ~/.config/fish/completions/%[1]v.fish

Powershell:

PS> %[1]v completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> %[1]v completion powershell > %[1]v.ps1
# and source this file from your powershell profile.
`, c.AppName(),
			),
			DisableFlagsInUseLine: true,
			ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
			Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
			Run: func(cmd *cobra.Command, args []string) {
				switch args[0] {
				case "bash":
					_ = cmd.Root().GenBashCompletion(os.Stdout)
				case "zsh":
					_ = cmd.Root().GenZshCompletion(os.Stdout)
				case "fish":
					_ = cmd.Root().GenFishCompletion(os.Stdout, true)
				case "powershell":
					_ = cmd.Root().GenPowerShellCompletion(os.Stdout)
				}
			},
		},
		Config: c,
	}
}
