package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/B67687/Oh-My-Learner/core"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <subject>",
	Short: "Install a subject pack",
	Long: `Install a subject pack from a TOML template file.

The subject pack is searched for in order:
  1. subjects/{id}.toml in the current directory
  2. subjects/{id}.toml next to the binary

Existing cards are preserved; only new cards are inserted.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID := strings.ToLower(args[0])

		packPath, err := findSubjectPack(subjectID)
		if err != nil {
			return err
		}

		data, err := os.ReadFile(packPath)
		if err != nil {
			return fmt.Errorf("failed to read subject pack: %w", err)
		}

		templates, err := core.LoadSubjectPack(data, subjectID)
		if err != nil {
			return fmt.Errorf("failed to load subject pack: %w", err)
		}

		if len(templates) == 0 {
			return fmt.Errorf("subject pack '%s' contains no templates", subjectID)
		}

		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		// Get subject metadata (name, prerequisites).
		subjectName, subjectPrereqs, err := core.SubjectPackMeta(data)
		if err != nil {
			return fmt.Errorf("failed to parse subject pack meta: %w", err)
		}
		if subjectName == "" {
			subjectName = subjectID
		}

		// Upsert subject with real name.
		if err := store.UpsertSubject(subjectID, subjectName); err != nil {
			return fmt.Errorf("failed to upsert subject: %w", err)
		}

		// Store prerequisites.
		if len(subjectPrereqs) > 0 {
			if err := store.SetPrerequisites(subjectID, subjectPrereqs); err != nil {
				return fmt.Errorf("failed to set prerequisites: %w", err)
			}
		}

		// Insert one card per template, skipping existing on conflict.
		newCardCount := 0
		for _, t := range templates {
			t.SubjectID = subjectID

			card := core.DefaultCardState(uuid.NewString(), subjectID)
			if err := store.InsertCard(card, t); err != nil {
				// Card likely already exists — skip silently.
				continue
			}
			newCardCount++
		}

		fmt.Printf("✓ Installed subject '%s' (%d templates, %d new)\n",
			subjectID, len(templates), newCardCount)

		return nil
	},
}
