package core

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

// SubjectCardCount counts all cards for a subject.
type SubjectCardCount struct {
	SubjectID string
	Count     int
}

// StreakInfo holds the user's current and longest review streak.
type StreakInfo struct {
	CurrentStreak  int
	LongestStreak  int
	LastReviewDate string
}

// DailyActivity records a single day's review summary.
type DailyActivity struct {
	Date          string
	CardsReviewed int
	CardsRecalled int
}

// WeeklyRetention holds the retention rate for the current week.
type WeeklyRetention struct {
	TotalReviewed int
	TotalRecalled int
	Rate          float64 // 0.0 to 1.0
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
		knowledge_type   TEXT NOT NULL DEFAULT 'declarative',
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
	CREATE TABLE IF NOT EXISTS subject_prerequisites (
		subject_id  TEXT NOT NULL REFERENCES subjects(id),
		prerequisite_id TEXT NOT NULL REFERENCES subjects(id),
		PRIMARY KEY (subject_id, prerequisite_id)
	);
	CREATE INDEX IF NOT EXISTS idx_prereq_subject ON subject_prerequisites(subject_id);
	CREATE TABLE IF NOT EXISTS streak (
		id              INTEGER PRIMARY KEY CHECK (id = 1),
		current_streak  INTEGER NOT NULL DEFAULT 0,
		longest_streak  INTEGER NOT NULL DEFAULT 0,
		last_review_date TEXT
	);
	CREATE TABLE IF NOT EXISTS daily_log (
		date           TEXT PRIMARY KEY,
		cards_reviewed INTEGER NOT NULL DEFAULT 0,
		cards_recalled INTEGER NOT NULL DEFAULT 0
	);
	`
	if _, err := db.Exec(schema); err != nil {
		return err
	}

	// Migration: add self_explain_response column to reviews if missing.
	if _, err := db.Exec(`ALTER TABLE reviews ADD COLUMN self_explain_response TEXT`); err != nil {
		fmt.Fprintf(os.Stderr, "migration warn: %v\n", err)
	}

	// Ensure streak singleton row exists.
	if _, err := db.Exec(`INSERT OR IGNORE INTO streak (id, current_streak, longest_streak) VALUES (1, 0, 0)`); err != nil {
		fmt.Fprintf(os.Stderr, "migration warn: %v\n", err)
	}

	// Backfill template_type if column exists but is empty (schema upgrade).
	if _, err := db.Exec(`UPDATE cards SET template_type = 'standard' WHERE template_type IS NULL OR template_type = ''`); err != nil {
		fmt.Fprintf(os.Stderr, "migration warn: %v\n", err)
	}

	// Backfill knowledge_type for existing cards that may have been created before the column existed.
	if _, err := db.Exec(`UPDATE cards SET knowledge_type = 'declarative' WHERE knowledge_type IS NULL OR knowledge_type = ''`); err != nil {
		fmt.Fprintf(os.Stderr, "migration warn: %v\n", err)
	}
	return nil
}

func (s *Storage) Close() error {
	return s.db.Close()
}

// GetDB returns the underlying database handle for custom queries.
func (s *Storage) GetDB() *sql.DB { return s.db }
