package core

import "time"

// TemplateType represents the kind of practice problem a template generates.
type TemplateType string

const (
	TemplateStandard   TemplateType = "standard"    // Q&A (existing behavior)
	TemplateCodeTrace  TemplateType = "code-trace"  // What does this code output?
	TemplateDebugFind  TemplateType = "debug-find"  // What is the bug in this code?
	TemplateExplainWhy TemplateType = "explain-why" // Explain why this concept works this way
)

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

// ReviewQuality is an SM-2 quality rating (0-5).
type ReviewQuality uint8

func (q ReviewQuality) IsPassing() bool { return q >= 3 }

// Template generates practice problems from parameterized variables.
type Template struct {
	ID               string              `toml:"id"`
	SubjectID        string              `toml:"-"`
	Type             TemplateType        `toml:"type"`
	QuestionTemplate string              `toml:"question"`
	AnswerTemplate   string              `toml:"answer"`
	Variables        map[string][]string `toml:"variables"`
}

// SubjectMeta is the metadata for an installed subject (name + prerequisites).
type SubjectMeta struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Prerequisites []string `json:"prerequisites,omitempty"`
}

// RenderedProblem is a concrete problem instance with bound variables.
type RenderedProblem struct {
	Question string
	Answer   string
	Bindings map[string]string
}

// SubjectDueCount is a subject with its due-card and total-card counts.
type SubjectDueCount struct {
	ID       string
	Name     string
	DueCount int
}

// CardWithTemplate is a card loaded with template data for rendering.
type CardWithTemplate struct {
	State            CardState
	Type             TemplateType
	QuestionTemplate string
	AnswerTemplate   string
	Variables        map[string][]string
}
