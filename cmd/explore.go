package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/B67687/Oh-My-Learner/core"
	"github.com/spf13/cobra"
)

var exploreCmd = &cobra.Command{
	Use:   "explore [subject]",
	Short: "Explore topic map with card counts and due status",
	Long: `Display the topic map showing subjects, card counts, and due status.

Without arguments, shows all subjects with card counts, due counts, and
completion percentages. With a subject argument, shows the detailed view
with card types and prerequisite chains.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		now := time.Now()

		if len(args) == 0 {
			return showSubjectOverview(store, now)
		}
		return showSubjectDetail(store, args[0], now)
	},
}

func showSubjectOverview(store *core.Storage, now time.Time) error {
	dueCounts, err := store.SubjectDueCounts(now)
	if err != nil {
		return fmt.Errorf("failed to get due counts: %w", err)
	}

	totalCounts, err := store.SubjectCardCounts()
	if err != nil {
		return fmt.Errorf("failed to get card counts: %w", err)
	}

	totalBySubject := make(map[string]int)
	for _, tc := range totalCounts {
		totalBySubject[tc.SubjectID] = tc.Count
	}

	metas, err := store.SubjectMetas()
	if err != nil {
		return fmt.Errorf("failed to get subject metas: %w", err)
	}

	nameByID := make(map[string]string)
	prereqsByID := make(map[string][]string)
	for _, m := range metas {
		nameByID[m.ID] = m.Name
		prereqsByID[m.ID] = m.Prerequisites
	}

	if len(dueCounts) == 0 {
		fmt.Println("No subjects installed. Use 'learn add <topic>' to add one.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "  Subject\tCards\tDue\tComplete\tPrerequisites")

	for _, dc := range dueCounts {
		total := totalBySubject[dc.ID]
		pct := 0.0
		if total > 0 {
			pct = float64(total-dc.DueCount) / float64(total) * 100
		}
		prereqs := prereqsByID[dc.ID]
		prereqStr := ""
		if len(prereqs) > 0 {
			names := make([]string, len(prereqs))
			for i, p := range prereqs {
				if n, ok := nameByID[p]; ok {
					names[i] = n
				} else {
					names[i] = p
				}
			}
			prereqStr = fmt.Sprintf("need: %s", strings.Join(names, ", "))
		}
		fmt.Fprintf(w, "  %s\t%d\t%d\t%.0f%%\t%s\n", dc.Name, total, dc.DueCount, pct, prereqStr)
	}

	return w.Flush()
}

func showSubjectDetail(store *core.Storage, subjectID string, now time.Time) error {
	metas, err := store.SubjectMetas()
	if err != nil {
		return fmt.Errorf("failed to get subject metas: %w", err)
	}

	// Find the subject.
	var target *core.SubjectMeta
	for _, m := range metas {
		if m.ID == subjectID || m.Name == subjectID {
			target = &m
			break
		}
	}
	if target == nil {
		return fmt.Errorf("subject %q not found", subjectID)
	}

	fmt.Printf("\n  Subject: %s (%s)\n", target.Name, target.ID)

	// Show prerequisites.
	if len(target.Prerequisites) > 0 {
		fmt.Println("  Prerequisites:")
		for _, p := range target.Prerequisites {
			fmt.Printf("    - %s\n", p)
		}
	} else {
		fmt.Println("  No prerequisites")
	}

	// Show card stats.
	counts, err := store.SubjectCardCounts()
	if err != nil {
		return fmt.Errorf("failed to get card counts: %w", err)
	}
	totalCards := 0
	for _, cc := range counts {
		if cc.SubjectID == target.ID {
			totalCards = cc.Count
			break
		}
	}

	dueCards, err := store.DueCards(now)
	if err != nil {
		return fmt.Errorf("failed to get due cards: %w", err)
	}
	totalDue := 0
	for _, c := range dueCards {
		if c.SubjectID == target.ID {
			totalDue++
		}
	}

	fmt.Printf("\n  Cards: %d total, %d due (%.0f%% complete)\n",
		totalCards, totalDue,
		percentComplete(totalCards, totalDue))

	fmt.Println()
	return nil
}

func percentComplete(total, due int) float64 {
	if total == 0 {
		return 100
	}
	return float64(total-due) / float64(total) * 100
}
