package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var hookShell string
var hookTmux bool

var hookCmd = &cobra.Command{
	Use:   "hook",
	Short: "Install shell/tmux hooks for review reminders",
	Long: `Install implementation-intention hooks that show due counts.

--shell bash|zsh:  Install a shell prompt hook showing due card count.
--tmux:            Install a tmux status integration.

Hooks are written to ~/.config/oh-my-learner/hooks/ and instructions
for activating them are printed to stdout.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		configDir := getConfigDir()
		hooksDir := filepath.Join(configDir, "hooks")
		if err := os.MkdirAll(hooksDir, 0755); err != nil {
			return fmt.Errorf("failed to create hooks directory: %w", err)
		}

		installed := 0

		if hookShell != "" {
			if err := installShellHook(hooksDir, hookShell); err != nil {
				return fmt.Errorf("install shell hook: %w", err)
			}
			installed++
		}

		if hookTmux {
			if err := installTmuxHook(hooksDir); err != nil {
				return fmt.Errorf("install tmux hook: %w", err)
			}
			installed++
		}

		if installed == 0 {
			fmt.Println("No hook type specified. Use --shell bash|zsh or --tmux.")
			fmt.Println("Example: learn hook --shell zsh")
			return nil
		}

		return nil
	},
}

func installShellHook(hooksDir, shell string) error {
	hookPath := filepath.Join(hooksDir, "oh-my-learner-prompt."+shell)

	content := `# oh-my-learner prompt hook — shows due card count
# Source this file from your ~/.` + shell + `rc:
#   source "` + hookPath + `"

__oh_my_learner_due() {
    local due
    due=$(learn status --count 2>/dev/null)
    if [ -n "$due" ] && [ "$due" -gt 0 ]; then
        echo " 📚${due}"
    fi
}

# Append to existing PROMPT or PS1
if [ -n "$ZSH_VERSION" ]; then
    # zsh: add to RPROMPT
    if [[ ! "$RPROMPT" == *"__oh_my_learner_due"* ]]; then
        RPROMPT='$(__oh_my_learner_due)'" $RPROMPT"
    fi
elif [ -n "$BASH_VERSION" ]; then
    # bash: add to PS1
    if [[ ! "$PS1" == *"__oh_my_learner_due"* ]]; then
        PS1='$(__oh_my_learner_due)'" $PS1"
    fi
fi
`
	if err := os.WriteFile(hookPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write shell hook: %w", err)
	}

	fmt.Printf("✓ Shell hook installed: %s\n", hookPath)
	fmt.Printf("  Add to your ~/.%src:\n", shell)
	fmt.Printf("    source %s\n", hookPath)

	return nil
}

func installTmuxHook(hooksDir string) error {
	hookPath := filepath.Join(hooksDir, "oh-my-learner-tmux.conf")

	content := `# oh-my-learner tmux status integration
# Add to ~/.tmux.conf:
#   run-shell "source-file ~/.config/oh-my-learner/hooks/oh-my-learner-tmux.conf"

set -g status-right "#(learn status --count 2>/dev/null | xargs -I{} echo {} cards due) | #[fg=white]%H:%M"
`
	if err := os.WriteFile(hookPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("write tmux hook: %w", err)
	}

	fmt.Printf("✓ Tmux hook installed: %s\n", hookPath)
	fmt.Println("  Add to ~/.tmux.conf:")
	fmt.Printf("    run-shell \"source-file %s\"\n", hookPath)

	return nil
}

func init() {
	hookCmd.Flags().StringVarP(&hookShell, "shell", "s", "",
		"Install shell prompt hook (bash or zsh)")
	hookCmd.Flags().BoolVarP(&hookTmux, "tmux", "t", false,
		"Install tmux status integration")
}
