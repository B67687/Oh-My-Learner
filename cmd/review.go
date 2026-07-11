package cmd

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"oh-my-learner/core"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review due cards",
	Long: `Practice cards that are due for review.

Each card shows a question, waits for you to reveal the answer, then
asks for a quality rating (0-5) to update the spaced repetition schedule.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		now := time.Now()
		cards, err := store.DueCards(now)
		if err != nil {
			return fmt.Errorf("failed to get due cards: %w", err)
		}

		if len(cards) == 0 {
			fmt.Println("No cards due for review. Great work!")
			return nil
		}

		// Count unique subjects for interleaving display.
		subjects := make(map[string]int)
		for _, c := range cards {
			subjects[c.SubjectID]++
		}

		fmt.Printf("\n  Session: %d cards across %d subjects\n", len(cards), len(subjects))
		if len(cards) < 10 {
			fmt.Println("  Tip: fewer than 10 cards due — interleaving works best with more cards.")
		} else if len(subjects) < 2 && len(cards) >= 10 {
			fmt.Println("  Tip: all cards are from one subject. Add another subject for interleaving benefits.")
		}
		fmt.Println()

		for i, card := range cards {
			cwt, err := store.GetCardWithTemplate(card.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading template for card %s: %v\n", card.ID, err)
				continue
			}

			tmpl := core.Template{
				ID:               cwt.State.ID,
				SubjectID:        cwt.State.SubjectID,
				QuestionTemplate: cwt.QuestionTemplate,
				AnswerTemplate:   cwt.AnswerTemplate,
				Variables:        cwt.Variables,
			}

			rendered, err := core.RenderTemplate(tmpl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error rendering card %s: %v\n", card.ID, err)
				continue
			}

			fmt.Printf("\n  %d/%d -- %s --\n", i+1, len(cards), card.SubjectID)
			fmt.Printf("  Q: %s\n", rendered.Question)

			fmt.Print("  [press Enter to reveal answer]")
			if _, err := readLine(); err != nil {
				return fmt.Errorf("read error: %w", err)
			}

			fmt.Printf("  A: %s\n", rendered.Answer)

			var quality core.ReviewQuality
			for {
				fmt.Print("  Quality (0-5): ")
				input, err := readLine()
				if err != nil {
					return fmt.Errorf("read error: %w", err)
				}
				input = strings.TrimSpace(input)
				n, err := strconv.Atoi(input)
				if err != nil || n < 0 || n > 5 {
					fmt.Println("  Please enter a number between 0 and 5.")
					continue
				}
				quality = core.ReviewQuality(n)
				break
			}

			updated := core.ReviewCard(card, quality)
			if err := store.UpdateCardState(updated); err != nil {
				return fmt.Errorf("failed to update card: %w", err)
			}
			if err := store.InsertReview(card.ID, uint8(quality)); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to record review: %v\n", err)
			}

			daysUntil := time.Until(updated.NextReviewAt).Hours() / 24
			switch {
			case daysUntil < 1:
				fmt.Println("  Next review: tomorrow")
			case daysUntil < 2:
				fmt.Println("  Next review: in 1 day")
			default:
				fmt.Printf("  Next review: in %d days\n", int(math.Ceil(daysUntil)))
			}
		}

		return nil
	},
}
