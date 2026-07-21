package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/B67687/Oh-My-Learner/agent"
	"github.com/B67687/Oh-My-Learner/core"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var aiMode bool

var addCmd = &cobra.Command{
	Use:   "add <subject-or-topic>",
	Short: "Install a subject pack or generate cards via AI",
	Long: `Install a subject pack from a TOML template file, or generate cards
via AI for a given topic.

Without --ai:
  The subject pack is searched for in order:
    1. subjects/{id}.toml in the current directory
    2. subjects/{id}.toml next to the binary
  Existing cards are preserved; only new cards are inserted.

With --ai:
  Uses DeepSeek V4 Flash to generate 5-10 practice cards about the
  given topic. Requires OML_DEEPSEEK_KEY environment variable.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subjectID := strings.ToLower(args[0])

		if aiMode {
			return addWithAI(subjectID)
		}
		return addFromPack(subjectID)
	},
}

func init() {
	addCmd.Flags().BoolVarP(&aiMode, "ai", "a", false,
		"Generate cards via AI instead of loading from TOML pack")
}

func addFromPack(subjectID string) error {
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
}

func addWithAI(topic string) error {
	aiAgent, err := agent.NewDeepSeekAgent()
	if err != nil {
		return fmt.Errorf("failed to create AI agent: %w", err)
	}

	fmt.Printf("🤖 Generating cards for '%s'...\n", topic)

	cards, err := aiAgent.GenerateCards(topic)
	if err != nil {
		return fmt.Errorf("AI generation failed: %w", err)
	}

	store, err := core.NewStorage(getDBPath())
	if err != nil {
		return fmt.Errorf("failed to open storage: %w", err)
	}
	defer store.Close()

	// Use topic as subject ID.
	subjectID := strings.ReplaceAll(strings.ToLower(topic), " ", "-")

	// Upsert subject.
	if err := store.UpsertSubject(subjectID, topic); err != nil {
		return fmt.Errorf("failed to upsert subject: %w", err)
	}

	// Insert each AI-generated card.
	newCardCount := 0
	for _, c := range cards {
		tmpl := core.Template{
			ID:               fmt.Sprintf("ai-%s-%s", subjectID, uuid.NewString()[:8]),
			SubjectID:        subjectID,
			Type:             core.TemplateType(c.Type),
			KnowledgeType:    core.KnowledgeType(c.KnowledgeType),
			QuestionTemplate: c.Question,
			AnswerTemplate:   c.Answer,
			Variables:        nil,
		}
		card := core.DefaultCardState(uuid.NewString(), subjectID)
		if err := store.InsertCard(card, tmpl); err != nil {
			return fmt.Errorf("failed to insert card: %w", err)
		}
		newCardCount++
	}

	fmt.Printf("✓ Generated %d cards for '%s' from AI\n", newCardCount, topic)
	return nil
}

// DefaultCardState is defined in core but aliased here for clarity.
// It creates a new card with SM-2 defaults (interval=0, ease=2.5, due=now).
