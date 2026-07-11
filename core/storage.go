// Package core provides SQLite-backed persistence for oh-my-learner.
// Uses modernc.org/sqlite (pure Go, no CGO) via database/sql.
package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// ─── Additional types (used by Storage, not in core.go) ────────────────

// SubjectCardCount counts all cards for a subject.
type SubjectCardCount struct {
	SubjectID string
	Count     int
}

// ─── Storage ────────────────────────────────────────────────────────────

// Storage provides SQLite-backed persistence for subjects, cards, and reviews.
type Storage struct {
	db *sql.DB
}

// NewStorage opens (or creates) the SQLite database at path, runs auto-migration,
// and returns a ready-to-use Storage.
func NewStorage(path string) (*Storage, error) {
	dsn := path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := migrate(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &Storage{db: db}, nil
}

// migrate creates the schema if it does not exist.
func migrate(db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS subjects (
		id   TEXT PRIMARY KEY,
		name TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS cards (
		id               TEXT PRIMARY KEY,
		subject_id       TEXT NOT NULL REFERENCES subjects(id),
		template_question TEXT NOT NULL,
		template_answer   TEXT NOT NULL,
		variables        TEXT NOT NULL,
		easiness_factor  REAL NOT NULL DEFAULT 2.5,
		interval_days    INTEGER NOT NULL DEFAULT 0,
		repetition       INTEGER NOT NULL DEFAULT 0,
		next_review_at   TEXT NOT NULL,
		created_at       TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS reviews (
		id          TEXT PRIMARY KEY,
		card_id     TEXT NOT NULL REFERENCES cards(id),
		quality     INTEGER NOT NULL,
		reviewed_at TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_cards_subject    ON cards(subject_id);
	CREATE INDEX IF NOT EXISTS idx_cards_next_review ON cards(next_review_at);
	CREATE INDEX IF NOT EXISTS idx_reviews_card     ON reviews(card_id);
	`
	_, err := db.Exec(schema)
	return err
}

// Close shuts down the database connection.
func (s *Storage) Close() error {
	return s.db.Close()
}

// ─── Subjects ───────────────────────────────────────────────────────────

// UpsertSubject inserts or replaces a subject row.
func (s *Storage) UpsertSubject(id, name string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO subjects (id, name) VALUES (?, ?)`,
		id, name,
	)
	return err
}

// RemoveSubject deletes a subject by id.
func (s *Storage) RemoveSubject(id string) error {
	_, err := s.db.Exec(`DELETE FROM subjects WHERE id = ?`, id)
	return err
}

// ─── Cards ──────────────────────────────────────────────────────────────

// DueCards returns all cards whose next_review_at is at or before now.
func (s *Storage) DueCards(now time.Time) ([]CardState, error) {
	nowStr := now.Format(time.RFC3339)
	rows, err := s.db.Query(
		`SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at
		   FROM cards
		  WHERE next_review_at <= ?
	   ORDER BY RANDOM()`, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("query due cards: %w", err)
	}
	defer rows.Close()

	var cards []CardState
	for rows.Next() {
		var (
			c             CardState
			nextReviewStr string
			createdAtStr  string
		)
		if err := rows.Scan(
			&c.ID, &c.SubjectID, &c.EasinessFactor, &c.IntervalDays,
			&c.Repetition, &nextReviewStr, &createdAtStr,
		); err != nil {
			return nil, fmt.Errorf("scan card: %w", err)
		}
		c.NextReviewAt, _ = time.Parse(time.RFC3339, nextReviewStr)
		c.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// SubjectDueCounts returns per-subject due-card counts at the given time.
func (s *Storage) SubjectDueCounts(now time.Time) ([]SubjectDueCount, error) {
	nowStr := now.Format(time.RFC3339)
	rows, err := s.db.Query(
		`SELECT s.id, s.name, COUNT(c.id)
		   FROM subjects s
		   LEFT JOIN cards c ON c.subject_id = s.id AND c.next_review_at <= ?
		  GROUP BY s.id, s.name
		  ORDER BY s.id`, nowStr,
	)
	if err != nil {
		return nil, fmt.Errorf("query subject due counts: %w", err)
	}
	defer rows.Close()

	var results []SubjectDueCount
	for rows.Next() {
		var r SubjectDueCount
		if err := rows.Scan(&r.ID, &r.Name, &r.DueCount); err != nil {
			return nil, fmt.Errorf("scan subject due count: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SubjectCardCounts returns the total card count for every subject.
func (s *Storage) SubjectCardCounts() ([]SubjectCardCount, error) {
	rows, err := s.db.Query(
		`SELECT s.id, COUNT(c.id)
		   FROM subjects s
		   LEFT JOIN cards c ON c.subject_id = s.id
		  GROUP BY s.id
		  ORDER BY s.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("query subject card counts: %w", err)
	}
	defer rows.Close()

	var results []SubjectCardCount
	for rows.Next() {
		var r SubjectCardCount
		if err := rows.Scan(&r.SubjectID, &r.Count); err != nil {
			return nil, fmt.Errorf("scan subject card count: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetCardWithTemplate loads a card together with its template data.
func (s *Storage) GetCardWithTemplate(cardID string) (*CardWithTemplate, error) {
	var (
		cwt           CardWithTemplate
		nextReviewStr string
		createdAtStr  string
		variablesStr  string
	)

	err := s.db.QueryRow(
		`SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at,
		        template_question, template_answer, variables
		   FROM cards WHERE id = ?`, cardID,
	).Scan(
		&cwt.State.ID, &cwt.State.SubjectID,
		&cwt.State.EasinessFactor, &cwt.State.IntervalDays, &cwt.State.Repetition,
		&nextReviewStr, &createdAtStr,
		&cwt.QuestionTemplate, &cwt.AnswerTemplate,
		&variablesStr,
	)
	if err != nil {
		return nil, fmt.Errorf("get card with template: %w", err)
	}

	cwt.State.NextReviewAt, _ = time.Parse(time.RFC3339, nextReviewStr)
	cwt.State.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

	if err := json.Unmarshal([]byte(variablesStr), &cwt.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &cwt, nil
}

// InsertCard creates a new card from its state and template data.
func (s *Storage) InsertCard(state CardState, tmpl Template) error {
	variablesJSON, err := json.Marshal(tmpl.Variables)
	if err != nil {
		return fmt.Errorf("marshal variables: %w", err)
	}

	_, err = s.db.Exec(
		`INSERT INTO cards
		   (id, subject_id, template_question, template_answer, variables,
		    easiness_factor, interval_days, repetition, next_review_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		state.ID, state.SubjectID,
		tmpl.QuestionTemplate, tmpl.AnswerTemplate,
		string(variablesJSON),
		state.EasinessFactor, state.IntervalDays, state.Repetition,
		state.NextReviewAt.Format(time.RFC3339),
		state.CreatedAt.Format(time.RFC3339),
	)
	return err
}

// UpdateCardState persists SM-2 scheduling fields for an existing card.
func (s *Storage) UpdateCardState(state CardState) error {
	_, err := s.db.Exec(
		`UPDATE cards
		    SET easiness_factor = ?, interval_days = ?, repetition = ?, next_review_at = ?
		  WHERE id = ?`,
		state.EasinessFactor, state.IntervalDays, state.Repetition,
		state.NextReviewAt.Format(time.RFC3339),
		state.ID,
	)
	return err
}

// ─── Reviews ────────────────────────────────────────────────────────────

// InsertReview records a review attempt for the given card.
func (s *Storage) InsertReview(cardID string, quality uint8) error {
	id := uuid.New().String()
	_, err := s.db.Exec(
		`INSERT INTO reviews (id, card_id, quality, reviewed_at) VALUES (?, ?, ?, ?)`,
		id, cardID, quality, time.Now().Format(time.RFC3339),
	)
	return err
}

// TodayReviewCount returns how many reviews were recorded today.
func (s *Storage) TodayReviewCount() (int, error) {
	today := time.Now().Format("2006-01-02")
	var count int
	err := s.db.QueryRow(
		`SELECT COUNT(*) FROM reviews WHERE reviewed_at LIKE ? || '%'`, today,
	).Scan(&count)
	return count, err
}
