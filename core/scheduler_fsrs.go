// Package core — FSRS-5 spaced repetition scheduler.
package core

import (
	"math"
	"time"
)

// ── FSRS-5 Parameters (Jarrett Ye) ────────────────────────────────────────────

// fsrsW is the 22-parameter set for the FSRS-5 algorithm.
//
//	w[0]  = initial stability (S₀)
//	w[1]  = unused / reserved
//	w[2]  = initial difficulty (D₀)
//	w[3]  = unused / reserved
//	w[4]  = difficulty delta scale (forget)
//	w[5]  = mean reversion weight for difficulty
//	w[6]  = difficulty delta scale (recall)
//	w[7]  = tiny adjustment
//	w[8]  = stability increase coefficient (used as exp(w[8]))
//	w[9]  = stability exponent in recall formula
//	w[10] = retrievability exponent factor
//	w[11] = forget stability multiplier
//	w[12] = forget stability exponent
//	w[13] = forget retrievability factor
//	w[14] = unused / reserved
//	w[15] = hard penalty multiplier (< 1 reduces stability)
//	w[16] = easy bonus multiplier (> 1 boosts stability)
//	w[17]–w[21] = unused / reserved for short-term scheduling
var fsrsW = []float64{
	0.40255, 1.18385, 3.173, 15.69105, 7.1949,
	0.5345, 1.4604, 0.0046, 1.54575, 0.1192,
	1.01925, 1.9395, 0.11, 0.29605, 1.2695,
	0.231, 2.9892, 0.8, 0.3105, 1.85,
	1.436, 1.5,
}

const (
	// fsrsRequestedRetention is the target retrieval probability for scheduling.
	fsrsRequestedRetention = 0.9

	// fsrsMaxInterval caps the scheduling interval in days.
	fsrsMaxInterval = 365
)

// ── FSRS Grade ────────────────────────────────────────────────────────────────

// fsrsGrade represents the four rating outcomes in FSRS-5.
type fsrsGrade int

const (
	fsrsAgain fsrsGrade = 1 // forgot (lapse)
	fsrsHard  fsrsGrade = 2 // recalled with difficulty
	fsrsGood  fsrsGrade = 3 // recalled normally
	fsrsEasy  fsrsGrade = 4 // recalled effortlessly
)

// qualityToFSRSGrade converts an SM-2 ReviewQuality (0–5) to an FSRS grade.
//
// Mapping:
//   - quality 5   → again (forgotten)
//   - quality 3–4 → hard
//   - quality 2   → good
//   - quality 0–1 → easy
func qualityToFSRSGrade(q ReviewQuality) fsrsGrade {
	switch {
	case q == 5:
		return fsrsAgain
	case q == 3 || q == 4:
		return fsrsHard
	case q == 2:
		return fsrsGood
	default: // 0, 1
		return fsrsEasy
	}
}

// ── Core FSRS-5 Formulas ──────────────────────────────────────────────────────

// retrievability computes the probability of recall after elapsed days
// given the current stability S:
//
//	R(t, S) = (1 + t / (9·S))⁻¹
func retrievability(elapsed, stability float64) float64 {
	if stability <= 0 {
		return 0
	}
	return math.Pow(1+elapsed/(9*stability), -1)
}

// stabilityAfterRecall computes the new stability after a successful recall
// (hard, good, or easy).  The base formula is:
//
//	S' = S · (1 + exp(w₈)·(11 − D)·S^(−w₉)·(exp((1−R)·w₁₀) − 1))
//
// Hard/easy bonuses are applied by the caller (see [FSRSScheduler.ReviewCard]).
func stabilityAfterRecall(s, d, r float64, w []float64) float64 {
	factor := 1 + math.Exp(w[8])*(11-d)*math.Pow(s, -w[9])*(math.Exp((1-r)*w[10])-1)
	ns := s * factor
	return math.Max(ns, w[0]) // floor at initial stability
}

// stabilityAfterForget computes the new stability after a lapse (again):
//
//	S' = w₁₁ · S^(w₁₂) · exp(w₁₃ · (1 − R))
//
// Stability after forgetting is capped to not exceed the previous value.
func stabilityAfterForget(s, d, r float64, w []float64) float64 {
	ns := w[11] * math.Pow(s, w[12]) * math.Exp(w[13]*(1-r))
	if ns > s {
		ns = s // cannot exceed previous stability
	}
	return math.Max(ns, w[0]) // floor at initial stability
}

// difficultyAfterRecall updates difficulty after a successful recall.
// Difficulty trends toward D₀ via mean reversion:
//
//	D' = clamp(w₅·D₀ + (1−w₅)·(D + Δ), 1, 10)
//
// where Δ = w₆ (a small positive bump so difficulty drifts slowly upward
// between lapses).
func difficultyAfterRecall(d float64, w []float64) float64 {
	d += w[6]                  // small increase over time
	d = w[5]*w[2] + (1-w[5])*d // mean reversion toward D₀
	return clamp(d, 1, 10)
}

// difficultyAfterForget updates difficulty after a lapse.
// Difficulty receives a larger increase:
//
//	D' = clamp(w₅·D₀ + (1−w₅)·(D + w₄·0.6), 1, 10)
func difficultyAfterForget(d float64, w []float64) float64 {
	d += w[4] * 0.6            // significant increase on forgetting
	d = w[5]*w[2] + (1-w[5])*d // mean reversion
	return clamp(d, 1, 10)
}

// clamp bounds v to [lo, hi].
func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// ── FSRSScheduler ─────────────────────────────────────────────────────────────

// FSRSScheduler implements the FSRS-5 spaced repetition algorithm.
type FSRSScheduler struct{}

// ReviewCard applies the FSRS-5 algorithm and returns the updated CardState.
//
// For cards not yet initialised (Stability == 0 or Difficulty == 0) the
// initial values S₀ = w₀ and D₀ = w₂ are used.
//
// After a lapse (again) the card is scheduled for immediate re-review
// (interval 0).  After a successful recall (hard / good / easy) the interval
// is computed from the new stability so that retrievability drops to the
// requested retention level (0.9):
//
//	I = 9·S · (1/r − 1)  =  S   (when r = 0.9)
func (s *FSRSScheduler) ReviewCard(card CardState, quality ReviewQuality) CardState {
	out := card
	grade := qualityToFSRSGrade(quality)
	w := fsrsW

	// ── Initialise stability / difficulty for new-to-FSRS cards ─────────────
	stability := card.Stability
	difficulty := card.Difficulty
	isNew := stability == 0
	if isNew {
		stability = w[0] // S₀
	}
	if difficulty == 0 {
		difficulty = w[2] // D₀
	}

	// ── Elapsed days since last review ──────────────────────────────────────
	elapsedDays := float64(card.IntervalDays)
	if overdue := time.Since(card.NextReviewAt).Hours() / 24; overdue > 0 {
		elapsedDays += overdue
	}

	// New cards have no retrieval history — treat as zero retrievability.
	var r float64
	if isNew {
		r = 0
	} else {
		r = retrievability(elapsedDays, stability)
	}

	// ── Update difficulty and stability per grade ───────────────────────────
	switch grade {
	case fsrsAgain:
		difficulty = difficultyAfterForget(difficulty, w)
		stability = stabilityAfterForget(stability, difficulty, r, w)

	case fsrsHard:
		difficulty = difficultyAfterRecall(difficulty, w)
		stability = stabilityAfterRecall(stability, difficulty, r, w)
		stability *= w[15] // hard penalty

	case fsrsGood:
		difficulty = difficultyAfterRecall(difficulty, w)
		stability = stabilityAfterRecall(stability, difficulty, r, w)

	case fsrsEasy:
		difficulty = difficultyAfterRecall(difficulty, w)
		stability = stabilityAfterRecall(stability, difficulty, r, w)
		stability *= w[16] // easy bonus
	}

	// ── Compute next review interval ────────────────────────────────────────
	// From the user's retrievability formula:
	//   R(I, S) = (1 + I/(9·S))⁻¹ = r_target
	// ⇒ I = 9·S·(1/r_target − 1)
	// With r_target = 0.9: I = 9·S·(1/0.9 − 1) = S
	interval := int(math.Round(stability))
	if interval > fsrsMaxInterval {
		interval = fsrsMaxInterval
	}
	if interval < 0 {
		interval = 0
	}

	out.Stability = stability
	out.Difficulty = difficulty
	out.IntervalDays = interval
	out.NextReviewAt = time.Now().AddDate(0, 0, interval)

	return out
}
