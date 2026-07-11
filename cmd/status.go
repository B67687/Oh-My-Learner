package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"oh-my-learner/core"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show review status",
	Long: `Display a summary of all installed subjects and their review status.

Shows how many cards are due today and the total card count per subject.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		now := time.Now()
		dueCounts, err := store.SubjectDueCounts(now)
		if err != nil {
			return fmt.Errorf("failed to get subject due counts: %w", err)
		}

		if len(dueCounts) == 0 {
			fmt.Println("No subjects installed.")
			return nil
		}

		totalCounts, err := store.SubjectCardCounts()
		if err != nil {
			return fmt.Errorf("failed to get subject card counts: %w", err)
		}

		// Build a lookup: subject ID → total card count.
		totalBySubject := make(map[string]int)
		for _, tc := range totalCounts {
			totalBySubject[tc.SubjectID] = tc.Count
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "  Subject\tDue Today\tTotal")

		grandDue := 0
		grandTotal := 0

		for _, dc := range dueCounts {
			total := totalBySubject[dc.ID]
			fmt.Fprintf(w, "  %s\t%d\t%d\n", dc.Name, dc.DueCount, total)
			grandDue += dc.DueCount
			grandTotal += total
		}

		fmt.Fprintln(w, "  ──────────────────────────────")
		fmt.Fprintf(w, "  Total\t%d\t%d\n", grandDue, grandTotal)

		return w.Flush()
	},
}
