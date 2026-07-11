package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "learn",
	Short: "Spaced-repetition practice problem generator",
	Long: `A CLI tool that generates practice problems from templates
and schedules them with spaced repetition + interleaving.

Subject-agnostic. No hand-written cards.`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.AddCommand(reviewCmd)
	RootCmd.AddCommand(addCmd)
	RootCmd.AddCommand(statusCmd)
	RootCmd.AddCommand(configCmd)
}
