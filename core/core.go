// Package core defines the shared types and API for oh-my-learner.
// CLI and future TUI consume this package. No external dependencies.
package core

import "time"

// ─── Card State (SM-2) ──────────────────────────────────────────────────────

// CardState represents the SM-2 spaced repetition state for a single card.
type CardState struct {
	ID             string    `json:"id"`
	SubjectID      string    `json:"subject_id"`
	EasinessFactor float64   `json:"easiness_factor"`
	IntervalDays   int       `json:"interval_days"`
	Repetition     int       `json:"repetition"`
	NextReviewAt   time.Time `json:"next_review_at"`
	CreatedAt      time.Time `json:"created_at"`
}

// ReviewQuality is an SM-2 quality rating (0–5).
// 0=blackout, 1-2=wrong, 3=hard, 4=good, 5=perfect.
type ReviewQuality uint8

// IsPassing returns true if the quality indicates correct recall (≥3).
func (q ReviewQuality) IsPassing() bool { return q >= 3 }

// ─── Templates ──────────────────────────────────────────────────────────────

// Template generates practice problems from parameterized variables.
type Template struct {
	ID               string              `toml:"id"`
	SubjectID        string              `toml:"-"`
	QuestionTemplate string              `toml:"question"`
	AnswerTemplate   string              `toml:"answer"`
	Variables        map[string][]string `toml:"variables"`
}

// RenderedProblem is a concrete problem instance with bound variables.
type RenderedProblem struct {
	Question string
	Answer   string
	Bindings map[string]string
}

// ─── Storage ────────────────────────────────────────────────────────────────

// SubjectDueCount is a subject with its due-card and total-card counts.
type SubjectDueCount struct {
	ID       string
	Name     string
	DueCount int
}

// CardWithTemplate is a card loaded with template data for rendering.
type CardWithTemplate struct {
	State            CardState
	QuestionTemplate string
	AnswerTemplate   string
	Variables        map[string][]string
}
