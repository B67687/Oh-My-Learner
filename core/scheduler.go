// Package core — SM-2 spaced repetition scheduler.
package core

import (
	"math"
	"time"
)

// Scheduler defines the interface for spaced repetition scheduling.
// Implementations include SM-2 (current) and FSRS (future).
type Scheduler interface {
	// ReviewCard applies the scheduling algorithm and returns the updated CardState.
	ReviewCard(card CardState, quality ReviewQuality) CardState
}

// SM2Scheduler implements the SM-2 spaced repetition algorithm.
type SM2Scheduler struct{}

// ReviewCard applies the SM-2 algorithm and returns the updated CardState.
//
// The algorithm:
//   - Updates easiness factor from the quality rating.
//   - For correct recall (quality >= 3): grows the repetition count and interval.
//   - For incorrect recall (quality < 3): resets repetition and interval to 1 day.
//   - Sets NextReviewAt to time.Now() + interval days.
func (s *SM2Scheduler) ReviewCard(card CardState, quality ReviewQuality) CardState {
	out := card

	// ── Update easiness factor ──────────────────────────────────────────
	q := float64(quality)
	delta := 0.1 - (5.0-q)*(0.08+(5.0-q)*0.02)
	out.EasinessFactor = card.EasinessFactor + delta
	if out.EasinessFactor < 1.3 {
		out.EasinessFactor = 1.3
	}

	// ── Update repetition and interval ──────────────────────────────────
	if quality.IsPassing() {
		out.Repetition = card.Repetition + 1

		switch out.Repetition {
		case 1:
			out.IntervalDays = 1
		case 2:
			out.IntervalDays = 6
		default:
			// interval = round(previous_interval * updated_EF)
			next := math.Round(float64(card.IntervalDays) * out.EasinessFactor)
			out.IntervalDays = int(next)
		}
	} else {
		out.Repetition = 0
		out.IntervalDays = 1
	}

	// ── Schedule next review ────────────────────────────────────────────
	out.NextReviewAt = time.Now().AddDate(0, 0, out.IntervalDays)

	return out
}

// DefaultCardState returns a new CardState with SM-2 defaults
// (easiness factor 2.5, zero repetitions, due immediately).
func DefaultCardState(id, subjectID string) CardState {
	now := time.Now()
	return CardState{
		ID:             id,
		SubjectID:      subjectID,
		EasinessFactor: 2.5,
		IntervalDays:   0,
		Repetition:     0,
		NextReviewAt:   now,
		CreatedAt:      now,
	}
}

// ReviewCard is a convenience wrapper calling the default SM-2 scheduler.
var defaultScheduler Scheduler = &SM2Scheduler{}

func ReviewCard(card CardState, quality ReviewQuality) CardState {
	return defaultScheduler.ReviewCard(card, quality)
}
