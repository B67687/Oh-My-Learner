package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/B67687/Oh-My-Learner/core"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Show weekly retention, streak, and activity log",
	Long: `Display your review statistics including:
- Current and longest streak
- Weekly retention rate
- Daily activity log (last 7 days)
- Recognition log`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := core.NewStorage(getDBPath())
		if err != nil {
			return fmt.Errorf("failed to open storage: %w", err)
		}
		defer store.Close()

		// ── Streak info ─────────────────────────────────────────────────────
		streak, err := store.GetStreak()
		if err != nil {
			return fmt.Errorf("failed to get streak: %w", err)
		}

		fmt.Printf("\n  Streak\n")
		fmt.Printf("    Current:  %d days\n", streak.CurrentStreak)
		fmt.Printf("    Longest:  %d days\n", streak.LongestStreak)
		if streak.LastReviewDate != "" {
			fmt.Printf("    Last review: %s\n", streak.LastReviewDate)
		} else {
			fmt.Printf("    Last review: never\n")
		}

		// ── Weekly retention ────────────────────────────────────────────────
		retention, err := store.GetWeeklyRetention()
		if err != nil {
			return fmt.Errorf("failed to get weekly retention: %w", err)
		}

		fmt.Printf("\n  Weekly Retention (last 7 days)\n")
		if retention.TotalReviewed > 0 {
			fmt.Printf("    Reviews:  %d\n", retention.TotalReviewed)
			fmt.Printf("    Recalled: %d\n", retention.TotalRecalled)
			fmt.Printf("    Rate:     %.1f%%\n", retention.Rate*100)
		} else {
			fmt.Println("    No reviews in the past 7 days.")
		}

		// ── Daily activity log (last 7 days) ────────────────────────────────
		fmt.Printf("\n  Daily Activity (last 7 days)\n")

		type dailyEntry struct {
			date          string
			cardsReviewed int
			cardsRecalled int
		}

		now := time.Now()
		weekAgo := now.AddDate(0, 0, -6)
		startOfWeek := weekAgo.Format("2006-01-02") + "T00:00:00Z"
		endOfWeek := now.Format("2006-01-02") + "T23:59:59Z"

		// Single query for all 7 days instead of 7 separate queries.
		rows, err := store.GetDB().Query(
			`SELECT DATE(reviewed_at) as day,
			        COUNT(*) as reviewed,
			        COALESCE(SUM(CASE WHEN quality >= 3 THEN 1 ELSE 0 END), 0) as recalled
			   FROM reviews
			  WHERE reviewed_at >= ? AND reviewed_at <= ?
			  GROUP BY DATE(reviewed_at)
			  ORDER BY day ASC`,
			startOfWeek, endOfWeek,
		)
		if err != nil {
			return fmt.Errorf("failed to query daily activity: %w", err)
		}
		defer rows.Close()

		dayMap := make(map[string]dailyEntry)
		for rows.Next() {
			var e dailyEntry
			if err := rows.Scan(&e.date, &e.cardsReviewed, &e.cardsRecalled); err != nil {
				return fmt.Errorf("failed to scan daily activity: %w", err)
			}
			dayMap[e.date] = e
		}

		var dailyLog []dailyEntry
		for i := 6; i >= 0; i-- {
			day := now.AddDate(0, 0, -i)
			dateStr := day.Format("2006-01-02")
			if entry, ok := dayMap[dateStr]; ok {
				dailyLog = append(dailyLog, entry)
			} else {
				dailyLog = append(dailyLog, dailyEntry{date: dateStr})
			}
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "  Date\tCards\tRecalled\tStreak")
		for _, entry := range dailyLog {
			streakMarker := ""
			if entry.cardsReviewed > 0 {
				streakMarker = "✓"
			}
			fmt.Fprintf(w, "  %s\t%d\t%d\t%s\n", entry.date, entry.cardsReviewed, entry.cardsRecalled, streakMarker)
		}
		w.Flush()

		// ── Self-explain responses (verbose only) ───────────────────────────
		if reportVerbose {
			fmt.Printf("\n  Recent Self-Explanations\n")

			seRows, err := store.GetDB().Query(
				`SELECT c.id, r.reviewed_at, r.self_explain_response
				   FROM reviews r
				   JOIN cards c ON c.id = r.card_id
				  WHERE r.self_explain_response IS NOT NULL AND r.self_explain_response != ''
				  ORDER BY r.reviewed_at DESC
				  LIMIT 20`,
			)
			if err != nil {
				return fmt.Errorf("failed to query self-explain responses: %w", err)
			}
			defer seRows.Close()

			count := 0
			for seRows.Next() {
				var cardID, reviewedAt, response string
				if err := seRows.Scan(&cardID, &reviewedAt, &response); err != nil {
					return fmt.Errorf("failed to scan self-explain: %w", err)
				}
				t, _ := time.Parse(time.RFC3339, reviewedAt)
				fmt.Printf("    %s  card=%s\n", t.Format("2006-01-02 15:04"), cardID[:8])
				fmt.Printf("      %s\n", response)
				count++
			}
			if count == 0 {
				fmt.Println("    No self-explanations recorded yet.")
			}
		}

		return nil
	},
}

var reportVerbose bool

func init() {
	reportCmd.Flags().BoolVarP(&reportVerbose, "verbose", "v", false, "show recent self-explain responses")
}
