package core

import (
	"math"
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
	result := (&SM2Scheduler{}).ReviewCard(card, 5)

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
	result := (&SM2Scheduler{}).ReviewCard(card, 5)

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
	result := (&SM2Scheduler{}).ReviewCard(card, 5)

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
	result := (&SM2Scheduler{}).ReviewCard(card, 0)

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
	result := (&SM2Scheduler{}).ReviewCard(card, 0)

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

// ── FSRS-5 Tests ──────────────────────────────────────────────────────────

func TestQualityToFSRSGrade(t *testing.T) {
	tests := []struct {
		quality  ReviewQuality
		expected fsrsGrade
	}{
		{0, fsrsEasy},
		{1, fsrsEasy},
		{2, fsrsGood},
		{3, fsrsHard},
		{4, fsrsHard},
		{5, fsrsAgain},
	}

	for _, tt := range tests {
		got := qualityToFSRSGrade(tt.quality)
		if got != tt.expected {
			t.Errorf("qualityToFSRSGrade(%d) = %d, want %d", tt.quality, got, tt.expected)
		}
	}
}

func TestRetrievability(t *testing.T) {
	// R(t, S) = (1 + t/(9·S))⁻¹
	tests := []struct {
		elapsed   float64
		stability float64
		expected  float64
	}{
		{0, 10, 1.0},                    // no elapsed time → perfect recall
		{10, 10, math.Pow(1+1/9.0, -1)}, // t=S → R ≈ 0.9
		{90, 10, math.Pow(2, -1)},       // t=9S → R = 0.5
		{0, 0, 0},                       // zero stability → R=0
	}

	for _, tt := range tests {
		got := retrievability(tt.elapsed, tt.stability)
		if math.Abs(got-tt.expected) > 1e-9 {
			t.Errorf("retrievability(%.1f, %.1f) = %.6f, want %.6f",
				tt.elapsed, tt.stability, got, tt.expected)
		}
	}
}

func TestStabilityAfterRecall(t *testing.T) {
	// Verify S' > S for a normal recall (retrievability ≈ requested retention).
	s := 10.0
	d := 3.0
	r := 0.9
	w := fsrsW

	ns := stabilityAfterRecall(s, d, r, w)
	if ns <= s {
		t.Errorf("stabilityAfterRecall(%.1f, %.1f, %.2f) = %.4f, want > %.1f", s, d, r, ns, s)
	}
	if ns < w[0] {
		t.Errorf("stabilityAfterRecall = %.4f, want >= initial stability %.4f", ns, w[0])
	}
}

func TestStabilityAfterForget(t *testing.T) {
	// Forgetting should reduce stability (or at least not increase it beyond s).
	s := 10.0
	d := 3.0
	r := 0.9
	w := fsrsW

	ns := stabilityAfterForget(s, d, r, w)
	if ns > s {
		t.Errorf("stabilityAfterForget(%.1f, %.1f, %.2f) = %.4f, want <= %.1f", s, d, r, ns, s)
	}
	if ns < w[0] {
		t.Errorf("stabilityAfterForget = %.4f, want >= initial stability %.4f", ns, w[0])
	}
}

func TestDifficultyClamp(t *testing.T) {
	w := fsrsW

	// After many forgets, difficulty should not exceed 10.
	d := 9.0
	for i := 0; i < 20; i++ {
		d = difficultyAfterForget(d, w)
		if d < 1 || d > 10 {
			t.Errorf("difficultyAfterForget = %.2f after %d iterations, want [1, 10]", d, i+1)
		}
	}

	// After many recalls, difficulty should not drop below 1.
	d = 2.0
	for i := 0; i < 20; i++ {
		d = difficultyAfterRecall(d, w)
		if d < 1 || d > 10 {
			t.Errorf("difficultyAfterRecall = %.2f after %d iterations, want [1, 10]", d, i+1)
		}
	}
}

func TestFSRSReviewCard_NewCardEasy(t *testing.T) {
	// quality 0 → easy → stability should increase, interval > 0.
	card := DefaultCardState("fsrs1", "math")
	sched := &FSRSScheduler{}

	result := sched.ReviewCard(card, 0) // easy

	// Stability must be set to a positive value.
	if result.Stability <= 0 {
		t.Errorf("Stability = %.4f, want > 0", result.Stability)
	}
	if result.Difficulty <= 0 {
		t.Errorf("Difficulty = %.4f, want > 0", result.Difficulty)
	}
	if result.IntervalDays <= 0 {
		t.Errorf("IntervalDays = %d, want > 0 for easy recall", result.IntervalDays)
	}
	// Easy bonus should be applied (w[16] = 2.9892 > 1).
	// The base stability (without easy bonus) would be lower.
	// We just verify the scheduler doesn't crash and produces sane output.
}

func TestFSRSReviewCard_NewCardAgain(t *testing.T) {
	// quality 5 → again → interval should be 0 (immediate re-review).
	card := DefaultCardState("fsrs2", "math")
	sched := &FSRSScheduler{}

	result := sched.ReviewCard(card, 5) // again

	if result.IntervalDays != 0 {
		t.Errorf("IntervalDays = %d, want 0 after again/lapse", result.IntervalDays)
	}
	if result.Stability <= 0 {
		t.Errorf("Stability = %.4f, want > 0", result.Stability)
	}
}

func TestFSRSReviewCard_HardPenalty(t *testing.T) {
	// quality 3 → hard; compare interval vs quality 2 → good on the same card.
	card := DefaultCardState("fsrs3", "math")
	sched := &FSRSScheduler{}

	goodResult := sched.ReviewCard(card, 2) // good
	hardResult := sched.ReviewCard(card, 3) // hard

	// Hard should produce a shorter interval than good (w[15] = 0.231).
	if hardResult.IntervalDays >= goodResult.IntervalDays {
		t.Errorf("hard interval %d >= good interval %d; hard penalty should reduce it",
			hardResult.IntervalDays, goodResult.IntervalDays)
	}
}

func TestFSRSReviewCard_MaxInterval(t *testing.T) {
	// Build a card with very high stability and verify interval is capped.
	card := CardState{
		ID:             "fsrs4",
		SubjectID:      "math",
		EasinessFactor: 2.5,
		IntervalDays:   100,
		Repetition:     20,
		NextReviewAt:   time.Now(),
		CreatedAt:      time.Now(),
		Stability:      500,
		Difficulty:     3,
	}
	sched := &FSRSScheduler{}

	result := sched.ReviewCard(card, 2) // good

	if result.IntervalDays > fsrsMaxInterval {
		t.Errorf("IntervalDays = %d, want <= %d (max interval)",
			result.IntervalDays, fsrsMaxInterval)
	}
}

func TestFSRSReviewCard_Sequence(t *testing.T) {
	// Run 10 "good" (quality 2) reviews and verify stability grows monotonically.
	card := DefaultCardState("fsrs-seq", "science")
	sched := &FSRSScheduler{}

	var prevStability float64 = -1
	for i := 0; i < 10; i++ {
		// Advance the clock so each review is on-schedule.
		card.NextReviewAt = time.Now()
		card = sched.ReviewCard(card, 2) // good

		if card.Stability < fsrsW[0] {
			t.Errorf("review %d: Stability = %.4f, want >= %.4f",
				i+1, card.Stability, fsrsW[0])
		}
		if card.Difficulty < 1 || card.Difficulty > 10 {
			t.Errorf("review %d: Difficulty = %.4f, want [1, 10]", i+1, card.Difficulty)
		}
		if i > 0 && card.Stability <= prevStability {
			t.Errorf("review %d: Stability = %.4f, want > prev %.4f (monotonic growth for good reviews)",
				i+1, card.Stability, prevStability)
		}
		prevStability = card.Stability
	}
}

func TestFSRSReviewCard_ImplementsInterface(t *testing.T) {
	// Compile-time interface satisfaction check.
	var _ Scheduler = (*FSRSScheduler)(nil)
}
