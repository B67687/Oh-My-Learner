package core

import (
	"database/sql"
	"fmt"
	"time"
)

// ── Streak Tracking ──────────────────────────────────────────────────────────

// UpdateStreak updates the review streak based on today's activity.
// Called after each review session completes. If the user reviewed cards
// today, the streak is extended. Missed days do NOT break the streak
// (backlog forgiveness: up to 2 missed days per week are forgiven).
func (s *Storage) UpdateStreak(today string) error {
	var (
		currentStreak  int
		longestStreak  int
		lastReviewDate sql.NullString
	)

	err := s.db.QueryRow(
		`SELECT current_streak, longest_streak, last_review_date FROM streak WHERE id = 1`,
	).Scan(&currentStreak, &longestStreak, &lastReviewDate)
	if err != nil {
		return fmt.Errorf("read streak: %w", err)
	}

	now, err := time.Parse("2006-01-02", today)
	if err != nil {
		return fmt.Errorf("parse today: %w", err)
	}

	if lastReviewDate.Valid && lastReviewDate.String == today {
		// Already reviewed today — no change.
		return nil
	}

	if !lastReviewDate.Valid || lastReviewDate.String == "" {
		// First review ever — start streak at 1.
		currentStreak = 1
	} else {
		lastDate, err := time.Parse("2006-01-02", lastReviewDate.String)
		if err != nil {
			return fmt.Errorf("parse last review date: %w", err)
		}
		daysDiff := now.Sub(lastDate).Hours() / 24

		if daysDiff <= 1 {
			// Consecutive day — increment streak.
			currentStreak++
		} else if daysDiff <= 3 {
			// Missed 1-2 days — forgive (streak preserved, not incremented).
			// Streak stays the same.
		} else {
			// Missed 3+ days — streak resets to 1.
			currentStreak = 1
		}
	}

	if currentStreak > longestStreak {
		longestStreak = currentStreak
	}

	_, err = s.db.Exec(
		`UPDATE streak SET current_streak = ?, longest_streak = ?, last_review_date = ? WHERE id = 1`,
		currentStreak, longestStreak, today,
	)
	return err
}

// GetStreak returns the current and longest streak info.
func (s *Storage) GetStreak() (*StreakInfo, error) {
	var info StreakInfo
	var lastDate sql.NullString

	err := s.db.QueryRow(
		`SELECT current_streak, longest_streak, last_review_date FROM streak WHERE id = 1`,
	).Scan(&info.CurrentStreak, &info.LongestStreak, &lastDate)
	if err != nil {
		return nil, fmt.Errorf("read streak: %w", err)
	}
	if lastDate.Valid {
		info.LastReviewDate = lastDate.String
	}
	return &info, nil
}

// LogDailyActivity records a day's review activity in the daily_log table.
func (s *Storage) LogDailyActivity(date string, reviewed int, recalled int) error {
	_, err := s.db.Exec(
		`INSERT INTO daily_log (date, cards_reviewed, cards_recalled)
		 VALUES (?, ?, ?)
		 ON CONFLICT(date) DO UPDATE SET
		   cards_reviewed = cards_reviewed + ?,
		   cards_recalled = cards_recalled + ?`,
		date, reviewed, recalled, reviewed, recalled,
	)
	return err
}

// GetWeeklyRetention returns the retention rate for the current week (last 7 days).
func (s *Storage) GetWeeklyRetention() (*WeeklyRetention, error) {
	weekAgo := time.Now().AddDate(0, 0, -7).Format(time.RFC3339)
	now := time.Now().Format(time.RFC3339)

	var wr WeeklyRetention

	// Count reviews in the past 7 days.
	row := s.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(CASE WHEN quality >= 3 THEN 1 ELSE 0 END), 0)
		   FROM reviews
		  WHERE reviewed_at >= ? AND reviewed_at <= ?`,
		weekAgo, now,
	)
	if err := row.Scan(&wr.TotalReviewed, &wr.TotalRecalled); err != nil {
		return nil, fmt.Errorf("query weekly retention: %w", err)
	}

	if wr.TotalReviewed > 0 {
		wr.Rate = float64(wr.TotalRecalled) / float64(wr.TotalReviewed)
	}

	return &wr, nil
}
