## Autocompletion

Reward fully supports autocompletion for the following shells:
* bash
* zsh
* fish
* powershell


### Bash

```
$ source <(reward completion bash)

# To load completions for each session, execute once:
Linux:
$ reward completion bash > /etc/bash_completion.d/reward
MacOS:
$ reward completion bash > /usr/local/etc/bash_completion.d/reward
```

### Zsh

If shell completion is not already enabled in your environment, you will need to enable it. You can execute the
following once:

```
$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ reward completion zsh > "${fpath[1]}/_reward"

# You will need to start a new shell for this setup to take effect.
```

### Fish

```
$ reward completion fish | source

# To load completions for each session, execute once:
$ reward completion fish > ~/.config/fish/completions/reward.fish
```

### Powershell

```
PS> reward completion powershell | Out-String | Invoke-Expression

# To load completions for every new session, run:
PS> reward completion powershell > reward.ps1
# and source this file from your powershell profile.
```