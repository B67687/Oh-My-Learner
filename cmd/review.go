package cmd

import (
	"fmt"
	"math/rand/v2"
	"os"
	"time"

	"github.com/B67687/Oh-My-Learner/core"
	"github.com/spf13/cobra"
)

var reviewMode string

type cardWithType struct {
	card     core.CardState
	template *core.CardWithTemplate
}

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review due cards",
	Long: `Practice cards that are due for review.

Each card shows a question, waits for you to reveal the answer, then
asks for a quality rating (0-5) to update the spaced repetition schedule.

Supports multiple template types:
- standard: Q&A (existing)
- code-trace: what does this code output?
- debug-find: what is the bug in this code?
- explain-why: explain why this concept works this way

With --mode speed, the self-explanation step is skipped for faster reviews.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		now := time.Now()

		// ── Load all due cards ──────────────────────────────────────────────
		allCards, err := store.DueCards(now)
		if err != nil {
			return fmt.Errorf("failed to get due cards: %w", err)
		}

		if len(allCards) == 0 {
			fmt.Println("No cards due for review. Great work!")
			return nil
		}

		// ── Backlog forgiveness: cap to daily limit ─────────────────────────
		limit := getDailyReviewLimit()
		if len(allCards) > limit {
			rand.Shuffle(len(allCards), func(i, j int) {
				allCards[i], allCards[j] = allCards[j], allCards[i]
			})
			allCards = allCards[:limit]
		}

		loaded := make([]cardWithType, 0, len(allCards))
		for _, c := range allCards {
			cwt, err := store.GetCardWithTemplate(c.ID)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error loading template for card %s: %v\n", c.ID, err)
				continue
			}
			loaded = append(loaded, cardWithType{
				card:     c,
				template: cwt,
			})
		}
		if len(loaded) == 0 {
			fmt.Println("No reviewable cards found.")
			return nil
		}

		// Split into procedural (interleaved) and declarative (blocked by subject).
		var procedural []cardWithType
		declarativeBySubject := make(map[string][]cardWithType)

		for _, cwt := range loaded {
			if cwt.template.KnowledgeType == core.KnowledgeProcedural {
				procedural = append(procedural, cwt)
			} else {
				declarativeBySubject[cwt.template.State.SubjectID] = append(declarativeBySubject[cwt.template.State.SubjectID], cwt)
			}
		}

		// Build session order: procedural first (interleaved), then declarative blocked by subject.
		rand.Shuffle(len(procedural), func(i, j int) {
			procedural[i], procedural[j] = procedural[j], procedural[i]
		})

		var sessionCards []cardWithType
		sessionCards = append(sessionCards, procedural...)

		// Add declarative subjects in random order, but cards within each subject kept together.
		subjectOrder := make([]string, 0, len(declarativeBySubject))
		for s := range declarativeBySubject {
			subjectOrder = append(subjectOrder, s)
		}
		rand.Shuffle(len(subjectOrder), func(i, j int) {
			subjectOrder[i], subjectOrder[j] = subjectOrder[j], subjectOrder[i]
		})
		for _, s := range subjectOrder {
			sessionCards = append(sessionCards, declarativeBySubject[s]...)
		}

		subjects := make(map[string]int)
		for _, cwt := range loaded {
			subjects[cwt.template.State.SubjectID]++
		}

		fmt.Printf("\n  Session: %d cards across %d subjects (procedural interleaved, declarative blocked)\n",
			len(sessionCards), len(subjects))
		fmt.Printf("  Mode: %s\n", reviewMode)
		fmt.Println()

		totalRecalled, totalReviewed, err := runReviewSession(sessionCards, store)
		if err != nil {
			return err
		}

		// ── Session summary ────────────────────────────────────────────────
		today := time.Now().Format("2006-01-02")
		if err := store.UpdateStreak(today); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update streak: %v\n", err)
		}
		if err := store.LogDailyActivity(today, totalReviewed, totalRecalled); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to log daily activity: %v\n", err)
		}

		streak, err := store.GetStreak()
		if err == nil {
			fmt.Printf("\n  Session complete! %d/%d recalled\n", totalRecalled, totalReviewed)
			fmt.Printf("  Current streak: %d days (longest: %d)\n", streak.CurrentStreak, streak.LongestStreak)
		}

		return nil
	},
}

func init() {
	reviewCmd.Flags().StringVarP(&reviewMode, "mode", "m", "normal",
		"Review mode: 'normal' (with self-explanation) or 'speed' (skip self-explanation)")
}
