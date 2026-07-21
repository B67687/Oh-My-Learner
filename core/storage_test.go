package core

import (
	"strings"
	"testing"
	"time"
)

func newTestStorage(t *testing.T) *Storage {
	t.Helper()
	s, err := NewStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatalf("NewStorage failed: %v", err)
	}
	return s
}

func insertTestCard(t *testing.T, s *Storage, subjectID string, kt KnowledgeType) {
	t.Helper()
	tmpl := Template{
		ID:               subjectID + "-test-card",
		SubjectID:        subjectID,
		KnowledgeType:    kt,
		Type:             TemplateStandard,
		QuestionTemplate: "Q: test?",
		AnswerTemplate:   "A: test answer.",
		Variables:        nil,
	}
	if err := s.InsertCard(DefaultCardState("card-"+subjectID+"-"+string(kt), subjectID), tmpl); err != nil {
		t.Fatalf("InsertCard failed: %v", err)
	}
}

func TestStreakTracking(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// Initially streak should be 0.
	info, err := s.GetStreak()
	if err != nil {
		t.Fatalf("GetStreak failed: %v", err)
	}
	if info.CurrentStreak != 0 {
		t.Errorf("initial current streak = %d, want 0", info.CurrentStreak)
	}

	// First review: streak becomes 1.
	if err := s.UpdateStreak("2026-01-01"); err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}
	info, _ = s.GetStreak()
	if info.CurrentStreak != 1 {
		t.Errorf("after first review, current streak = %d, want 1", info.CurrentStreak)
	}
	if info.LongestStreak != 1 {
		t.Errorf("after first review, longest streak = %d, want 1", info.LongestStreak)
	}

	// Consecutive day: streak becomes 2.
	if err := s.UpdateStreak("2026-01-02"); err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}
	info, _ = s.GetStreak()
	if info.CurrentStreak != 2 {
		t.Errorf("after consecutive day, current streak = %d, want 2", info.CurrentStreak)
	}
	if info.LongestStreak != 2 {
		t.Errorf("after consecutive day, longest streak = %d, want 2", info.LongestStreak)
	}

	// Same day: no change.
	if err := s.UpdateStreak("2026-01-02"); err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}
	info, _ = s.GetStreak()
	if info.CurrentStreak != 2 {
		t.Errorf("after same day update, current streak = %d, want 2", info.CurrentStreak)
	}

	// Miss 1 day (gap of 2 days): forgiven, streak preserved.
	if err := s.UpdateStreak("2026-01-04"); err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}
	info, _ = s.GetStreak()
	if info.CurrentStreak != 2 {
		t.Errorf("after 1-day miss (forgiven), current streak = %d, want 2", info.CurrentStreak)
	}
}

func TestStreakTracking_MissThreeDays_Resets(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// Build a 3-day streak.
	s.UpdateStreak("2026-01-01")
	s.UpdateStreak("2026-01-02")
	s.UpdateStreak("2026-01-03")

	info, _ := s.GetStreak()
	if info.CurrentStreak != 3 {
		t.Fatalf("expected streak 3, got %d", info.CurrentStreak)
	}

	// Miss 3+ days: streak resets to 1.
	if err := s.UpdateStreak("2026-01-07"); err != nil {
		t.Fatalf("UpdateStreak failed: %v", err)
	}
	info, _ = s.GetStreak()
	if info.CurrentStreak != 1 {
		t.Errorf("after 4-day gap, current streak = %d, want 1", info.CurrentStreak)
	}
	// Longest streak should still be 3.
	if info.LongestStreak != 3 {
		t.Errorf("longest streak should be 3, got %d", info.LongestStreak)
	}
}

func TestDueCardsWithKnowledgeTypeFilter(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// Insert subject.
	if err := s.UpsertSubject("math", "Mathematics"); err != nil {
		t.Fatalf("UpsertSubject failed: %v", err)
	}
	if err := s.UpsertSubject("cs", "Computer Science"); err != nil {
		t.Fatalf("UpsertSubject failed: %v", err)
	}

	// Insert cards of both knowledge types.
	insertTestCard(t, s, "math", KnowledgeDeclarative)
	insertTestCard(t, s, "cs", KnowledgeProcedural)
	insertTestCard(t, s, "math", KnowledgeProcedural)

	now := time.Now()

	// All cards due.
	all, err := s.DueCards(now)
	if err != nil {
		t.Fatalf("DueCards failed: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("DueCards() returned %d cards, want 3", len(all))
	}

	// Filter declarative.
	decl, err := s.DueCards(now, KnowledgeDeclarative)
	if err != nil {
		t.Fatalf("DueCards with declarative filter failed: %v", err)
	}
	if len(decl) != 1 {
		t.Errorf("DueCards(declarative) returned %d cards, want 1", len(decl))
	}

	// Filter procedural.
	proc, err := s.DueCards(now, KnowledgeProcedural)
	if err != nil {
		t.Fatalf("DueCards with procedural filter failed: %v", err)
	}
	if len(proc) != 2 {
		t.Errorf("DueCards(procedural) returned %d cards, want 2", len(proc))
	}
}

func TestDailyActivityLogging(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	if err := s.LogDailyActivity("2026-01-01", 5, 4); err != nil {
		t.Fatalf("LogDailyActivity failed: %v", err)
	}

	// Log same date again (should accumulate).
	if err := s.LogDailyActivity("2026-01-01", 3, 2); err != nil {
		t.Fatalf("LogDailyActivity (second call) failed: %v", err)
	}

	// Weekly retention.
	wr, err := s.GetWeeklyRetention()
	if err != nil {
		t.Fatalf("GetWeeklyRetention failed: %v", err)
	}
	if wr.TotalReviewed != 0 {
		t.Errorf("weekly retention total reviewed = %d, want 0 (no reviews in reviews table)", wr.TotalReviewed)
	}
}

func TestWeeklyRetentionFromReviews(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	if err := s.UpsertSubject("math", "Mathematics"); err != nil {
		t.Fatalf("UpsertSubject failed: %v", err)
	}
	insertTestCard(t, s, "math", KnowledgeDeclarative)

	// Insert some test reviews.
	if err := s.InsertReview("card-math-declarative", 5, ""); err != nil {
		t.Fatalf("InsertReview failed: %v", err)
	}
	if err := s.InsertReview("card-math-declarative", 4, ""); err != nil {
		t.Fatalf("InsertReview failed: %v", err)
	}
	if err := s.InsertReview("card-math-declarative", 2, ""); err != nil {
		t.Fatalf("InsertReview failed: %v", err)
	}

	// Weekly retention: 2 passing out of 3.
	wr, err := s.GetWeeklyRetention()
	if err != nil {
		t.Fatalf("GetWeeklyRetention failed: %v", err)
	}
	if wr.TotalReviewed != 3 {
		t.Errorf("total reviewed = %d, want 3", wr.TotalReviewed)
	}
	if wr.TotalRecalled != 2 {
		t.Errorf("total recalled = %d, want 2", wr.TotalRecalled)
	}
	if wr.Rate != 2.0/3.0 {
		t.Errorf("rate = %f, want %f", wr.Rate, 2.0/3.0)
	}
}

func TestUpdateNonExistentCard(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// Create a card state with a non-existent ID.
	state := DefaultCardState("nonexistent-card-id", "math")

	// UpdateCardState should return an error because no card has this ID.
	err := s.UpdateCardState(state)
	if err == nil {
		t.Fatal("UpdateCardState: expected error for non-existent card, got nil")
	}

	// Verify the error message mentions the card ID.
	if !strings.Contains(err.Error(), "nonexistent-card-id") {
		t.Errorf("error = %q, want it to mention card ID", err.Error())
	}
}

func TestSubjectDueCountsEmpty(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// With no subjects or cards, SubjectDueCounts should return a nil/empty list.
	now := time.Now()
	counts, err := s.SubjectDueCounts(now)
	if err != nil {
		t.Fatalf("SubjectDueCounts failed: %v", err)
	}
	if len(counts) != 0 {
		t.Errorf("SubjectDueCounts = %d entries, want 0", len(counts))
	}
}

func TestLogDailyActivityNoReviews(t *testing.T) {
	s := newTestStorage(t)
	defer s.Close()

	// Log a day with zero reviews — should succeed without error.
	if err := s.LogDailyActivity("2026-01-01", 0, 0); err != nil {
		t.Fatalf("LogDailyActivity(0,0) failed: %v", err)
	}

	// Logging the same day with zero again should also work (upsert with no change).
	if err := s.LogDailyActivity("2026-01-01", 0, 0); err != nil {
		t.Fatalf("LogDailyActivity(0,0) second call failed: %v", err)
	}
}
