package cmd

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/B67687/Oh-My-Learner/core"
)

// runReviewSession iterates through sessionCards, rendering templates,
// displaying questions, collecting quality ratings, applying SM-2, and
// persisting card state. Returns the number of reviewed and recalled cards.
func runReviewSession(sessionCards []cardWithType, store *core.Storage) (totalRecalled, totalReviewed int, err error) {
	sch := &core.SM2Scheduler{}

	for i, cwt := range sessionCards {
		c := cwt.card
		tmpl := core.Template{
			ID:               c.ID,
			SubjectID:        c.SubjectID,
			KnowledgeType:    cwt.template.KnowledgeType,
			Type:             cwt.template.Type,
			QuestionTemplate: cwt.template.QuestionTemplate,
			AnswerTemplate:   cwt.template.AnswerTemplate,
			Variables:        cwt.template.Variables,
		}

		rendered, err := core.RenderTemplate(tmpl)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error rendering card %s: %v\n", c.ID, err)
			continue
		}

		// Determine labeling.
		typeLabel := string(tmpl.Type)
		ktLabel := string(tmpl.KnowledgeType)
		ktBadge := "[" + ktLabel + "]"

		fmt.Printf("\n  %d/%d -- %s %s [%s] --\n", i+1, len(sessionCards), c.SubjectID, ktBadge, typeLabel)

		// ── Display question ────────────────────────────────────────────
		switch tmpl.Type {
		case core.TemplateCodeTrace:
			fmt.Println("  What does this code output?")
			fmt.Println("  ```")
			for _, line := range strings.Split(rendered.Question, "\n") {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println("  ```")
		case core.TemplateDebugFind:
			fmt.Println("  What's the bug in this code?")
			fmt.Println("  ```")
			for _, line := range strings.Split(rendered.Question, "\n") {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println("  ```")
		case core.TemplateExplainWhy:
			fmt.Printf("  Explain why: %s\n", rendered.Question)
		default:
			fmt.Printf("  Q: %s\n", rendered.Question)
		}

		// ── Reveal answer ───────────────────────────────────────────────
		fmt.Print("  [press Enter to reveal answer]")
		if _, err := readLine(); err != nil {
			return totalRecalled, totalReviewed, fmt.Errorf("read error: %w", err)
		}

		switch tmpl.Type {
		case core.TemplateCodeTrace:
			fmt.Println("  Output:")
			fmt.Println("  ```")
			for _, line := range strings.Split(rendered.Answer, "\n") {
				fmt.Printf("  %s\n", line)
			}
			fmt.Println("  ```")
		default:
			fmt.Printf("  A: %s\n", rendered.Answer)
		}

		var selfExplain string
		if reviewMode != "speed" {
			fmt.Println()
			fmt.Print("  [Why is this correct? (press Enter to skip)] ")
			if line, err := readLine(); err == nil {
				selfExplain = strings.TrimSpace(line)
			}
		}

		// ── Quality rating ──────────────────────────────────────────────
		var quality core.ReviewQuality
		for {
			fmt.Print("  Quality (0-5): ")
			input, err := readLine()
			if err != nil {
				return totalRecalled, totalReviewed, fmt.Errorf("read error: %w", err)
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

		// ── Apply SM-2 and persist ──────────────────────────────────────
		updated := sch.ReviewCard(c, quality)
		if err := store.UpdateCardState(updated); err != nil {
			return totalRecalled, totalReviewed, fmt.Errorf("failed to update card: %w", err)
		}
		if err := store.InsertReview(c.ID, uint8(quality), selfExplain); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to record review: %v\n", err)
		}

		totalReviewed++
		if quality.IsPassing() {
			totalRecalled++
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

	return
}
