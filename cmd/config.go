package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

const defaultConfigContent = `# oh-my-learner configuration
daily_review_limit = 50
`

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show or initialize configuration",
	Long: `Display the current configuration or create a default one if none exists.

Configuration is stored at ~/.config/oh-my-learner/config.toml.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("cannot find home directory: %w", err)
		}

		configDir := filepath.Join(homeDir, ".config", "oh-my-learner")
		configPath := filepath.Join(configDir, "config.toml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.WriteFile(configPath, []byte(defaultConfigContent), 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("Created default config at %s\n", configPath)
			return nil
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		fmt.Print(string(data))
		return nil
	},
}
