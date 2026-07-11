package core

import (
	"testing"
	"time"
)

func TestDefaultCardState(t *testing.T) {
	// Given: we create a default card
	before := time.Now()
	card := DefaultCardState("c1", "math")
	after := time.Now()

	// Then: fields are set to SM-2 defaults
	if card.ID != "c1" || card.SubjectID != "math" {
		t.Errorf("id / subject_id mismatch: %q / %q", card.ID, card.SubjectID)
	}
	if card.EasinessFactor != 2.5 {
		t.Errorf("initial EF = %.2f, want 2.5", card.EasinessFactor)
	}
	if card.IntervalDays != 0 {
		t.Errorf("initial interval = %d, want 0", card.IntervalDays)
	}
	if card.Repetition != 0 {
		t.Errorf("initial repetition = %d, want 0", card.Repetition)
	}

	// NextReviewAt should be set to roughly now
	if card.NextReviewAt.Before(before) || card.NextReviewAt.After(after) {
		t.Errorf("NextReviewAt %v not between %v and %v", card.NextReviewAt, before, after)
	}
}

func TestReviewCard_PerfectRecall_StartsAt1Day(t *testing.T) {
	// Given: a fresh card with defaults
	card := DefaultCardState("c2", "science")

	// When: perfect recall (quality 5)
	result := ReviewCard(card, 5)

	// Then: interval is 1 day, repetition is 1
	if result.IntervalDays != 1 {
		t.Errorf("interval = %d, want 1", result.IntervalDays)
	}
	if result.Repetition != 1 {
		t.Errorf("repetition = %d, want 1", result.Repetition)
	}
}

func TestReviewCard_SecondReview_Gives6Days(t *testing.T) {
	// Given: a card that has been reviewed once (rep=1, interval=1)
	card := CardState{
		ID:             "c3",
		SubjectID:      "science",
		EasinessFactor: 2.5,
		IntervalDays:   1,
		Repetition:     1,
		NextReviewAt:   time.Now(),
		CreatedAt:      time.Now(),
	}

	// When: second perfect review
	result := ReviewCard(card, 5)

	// Then: interval jumps to 6, repetition becomes 2
	if result.IntervalDays != 6 {
		t.Errorf("interval = %d, want 6", result.IntervalDays)
	}
	if result.Repetition != 2 {
		t.Errorf("repetition = %d, want 2", result.Repetition)
	}
}

// TestReviewCard_ThirdReview_AppliesEF checks that from the third review onward
// the interval is computed as round(prev_interval * EF).
func TestReviewCard_ThirdReview_AppliesEF(t *testing.T) {
	// Given: a card after two successful reviews
	card := CardState{
		ID:             "c4",
		SubjectID:      "lang",
		EasinessFactor: 2.5,
		IntervalDays:   6,
		Repetition:     2,
		NextReviewAt:   time.Now(),
		CreatedAt:      time.Now(),
	}

	// When: third perfect review
	result := ReviewCard(card, 5)

	// Then: interval = round(6 * 2.6) = round(15.6) = 16
	// (EF moves from 2.5 to 2.6 for q=5)
	if result.IntervalDays != 16 {
		t.Errorf("interval = %d, want 16 (round(6 × 2.6))", result.IntervalDays)
	}
	if result.Repetition != 3 {
		t.Errorf("repetition = %d, want 3", result.Repetition)
	}
	if result.EasinessFactor != 2.6 {
		t.Errorf("EF = %.2f, want 2.6", result.EasinessFactor)
	}
}

func TestReviewCard_FailedReview_Resets(t *testing.T) {
	// Given: a card that was well-established (rep=5, interval=100)
	card := CardState{
		ID:             "c5",
		SubjectID:      "history",
		EasinessFactor: 2.5,
		IntervalDays:   100,
		Repetition:     5,
		NextReviewAt:   time.Now(),
		CreatedAt:      time.Now(),
	}

	// When: blackout (quality 0)
	result := ReviewCard(card, 0)

	// Then: repetition resets to 0, interval drops to 1
	if result.Repetition != 0 {
		t.Errorf("repetition = %d, want 0", result.Repetition)
	}
	if result.IntervalDays != 1 {
		t.Errorf("interval = %d, want 1", result.IntervalDays)
	}
}

func TestReviewCard_EasinessFactorFloor(t *testing.T) {
	// Given: a card with low EF (2.0) — a failing review (quality 0) pushes EF
	// down by 0.8 to 1.2, which should be clamped to 1.3.
	card := CardState{
		ID:             "c6",
		SubjectID:      "physics",
		EasinessFactor: 2.0,
		IntervalDays:   1,
		Repetition:     0,
		NextReviewAt:   time.Now(),
		CreatedAt:      time.Now(),
	}

	// When: failing review (quality 0)
	result := ReviewCard(card, 0)

	// Then: EF is clamped to the 1.3 minimum
	if result.EasinessFactor < 1.3 {
		t.Errorf("EF = %.2f, want >= 1.3", result.EasinessFactor)
	}
	if result.EasinessFactor != 1.3 {
		t.Errorf("EF = %.2f, want exactly 1.3 (clamped)", result.EasinessFactor)
	}
	// And the failing-review reset still applies
	if result.Repetition != 0 {
		t.Errorf("repetition = %d, want 0", result.Repetition)
	}
	if result.IntervalDays != 1 {
		t.Errorf("interval = %d, want 1", result.IntervalDays)
	}
}
