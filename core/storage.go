package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

// SubjectCardCount counts all cards for a subject.
type SubjectCardCount struct {
	SubjectID string
	Count     int
}

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
		template_type    TEXT NOT NULL DEFAULT 'standard',
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
	if err != nil {
		return err
	}
	// Backfill template_type if column exists but is empty (schema upgrade).
	_, _ = db.Exec(`UPDATE cards SET template_type = 'standard' WHERE template_type IS NULL OR template_type = ''`)
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) UpsertSubject(id, name string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO subjects (id, name) VALUES (?, ?)`,
		id, name,
	)
	return err
}

func (s *Storage) RemoveSubject(id string) error {
	_, err := s.db.Exec(`DELETE FROM subjects WHERE id = ?`, id)
	return err
}

func (s *Storage) Subjects() ([]SubjectDueCount, error) {
	rows, err := s.db.Query(
		`SELECT s.id, s.name, COUNT(c.id)
		   FROM subjects s
		   LEFT JOIN cards c ON c.subject_id = s.id
		  GROUP BY s.id, s.name
		  ORDER BY s.id`,
	)
	if err != nil {
		return nil, fmt.Errorf("query subjects: %w", err)
	}
	defer rows.Close()

	var results []SubjectDueCount
	for rows.Next() {
		var r SubjectDueCount
		if err := rows.Scan(&r.ID, &r.Name, &r.DueCount); err != nil {
			return nil, fmt.Errorf("scan subject: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

func (s *Storage) AddCard(tmpl Template) (CardState, error) {
	id := uuid.New().String()
	now := time.Now()
	state := CardState{
		ID:             id,
		SubjectID:      tmpl.SubjectID,
		EasinessFactor: 2.5,
		IntervalDays:   0,
		Repetition:     0,
		NextReviewAt:   now,
		CreatedAt:      now,
	}
	if err := s.InsertCard(state, tmpl); err != nil {
		return CardState{}, err
	}
	return state, nil
}

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

// GetCardWithTemplate loads a card with its rendering data.
func (s *Storage) GetCardWithTemplate(cardID string) (*CardWithTemplate, error) {
	var (
		cwt           CardWithTemplate
		nextReviewStr string
		createdAtStr  string
		variablesStr  string
		typeStr       string
	)

	err := s.db.QueryRow(
		`SELECT id, subject_id, easiness_factor, interval_days, repetition,
		        next_review_at, created_at,
		        template_type, template_question, template_answer, variables
		   FROM cards WHERE id = ?`, cardID,
	).Scan(
		&cwt.State.ID, &cwt.State.SubjectID,
		&cwt.State.EasinessFactor, &cwt.State.IntervalDays, &cwt.State.Repetition,
		&nextReviewStr, &createdAtStr,
		&typeStr, &cwt.QuestionTemplate, &cwt.AnswerTemplate,
		&variablesStr,
	)
	if err != nil {
		return nil, fmt.Errorf("get card with template: %w", err)
	}

	cwt.State.NextReviewAt, _ = time.Parse(time.RFC3339, nextReviewStr)
	cwt.State.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)
	cwt.Type = TemplateType(typeStr)

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
		   (id, subject_id, template_type, template_question, template_answer, variables,
		    easiness_factor, interval_days, repetition, next_review_at, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		state.ID, state.SubjectID,
		string(tmpl.Type),
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
		    SET easiness_factor = ?, interval_days = ?, repetition = ?,
		        next_review_at = ?
		  WHERE id = ?`,
		state.EasinessFactor, state.IntervalDays, state.Repetition,
		state.NextReviewAt.Format(time.RFC3339),
		state.ID,
	)
	return err
}

// InsertReview records a review event for audit/analytics.
func (s *Storage) InsertReview(cardID string, quality uint8) error {
	id := uuid.New().String()
	_, err := s.db.Exec(
		`INSERT INTO reviews (id, card_id, quality, reviewed_at)
		 VALUES (?, ?, ?, ?)`,
		id, cardID, quality, time.Now().Format(time.RFC3339),
	)
	return err
}

// SubjectCardCounts returns total card counts per subject.
func (s *Storage) SubjectCardCounts() ([]SubjectCardCount, error) {
	rows, err := s.db.Query(
		`SELECT s.id, COUNT(c.id)
		   FROM subjects s
		   LEFT JOIN cards c ON c.subject_id = s.id
		  GROUP BY s.id
		  ORDER BY s.id`)
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
