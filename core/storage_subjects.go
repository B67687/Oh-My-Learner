package core

import (
	"fmt"
	"strings"
	"time"
)

func (s *Storage) UpsertSubject(id, name string) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO subjects (id, name) VALUES (?, ?)`,
		id, name,
	)
	return err
}

func (s *Storage) RemoveSubject(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete reviews for cards belonging to this subject.
	if _, err := tx.Exec(`DELETE FROM reviews WHERE card_id IN (SELECT id FROM cards WHERE subject_id = ?)`, id); err != nil {
		return fmt.Errorf("delete reviews: %w", err)
	}

	// Delete cards belonging to this subject.
	if _, err := tx.Exec(`DELETE FROM cards WHERE subject_id = ?`, id); err != nil {
		return fmt.Errorf("delete cards: %w", err)
	}

	// Delete prerequisite entries referencing this subject.
	if _, err := tx.Exec(`DELETE FROM subject_prerequisites WHERE subject_id = ? OR prerequisite_id = ?`, id, id); err != nil {
		return fmt.Errorf("delete prerequisites: %w", err)
	}

	// Delete the subject itself.
	if _, err := tx.Exec(`DELETE FROM subjects WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete subject: %w", err)
	}

	return tx.Commit()
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

// SubjectMetas returns all subjects with their names and prerequisites.
func (s *Storage) SubjectMetas() ([]SubjectMeta, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.name, GROUP_CONCAT(sp.prerequisite_id ORDER BY sp.prerequisite_id)
		FROM subjects s
		LEFT JOIN subject_prerequisites sp ON s.id = sp.subject_id
		GROUP BY s.id, s.name
		ORDER BY s.id
	`)
	if err != nil {
		return nil, fmt.Errorf("query subjects: %w", err)
	}
	defer rows.Close()

	var metas []SubjectMeta
	for rows.Next() {
		var m SubjectMeta
		var prereqsStr *string
		if err := rows.Scan(&m.ID, &m.Name, &prereqsStr); err != nil {
			return nil, fmt.Errorf("scan subject meta: %w", err)
		}
		if prereqsStr != nil && *prereqsStr != "" {
			m.Prerequisites = strings.Split(*prereqsStr, ",")
		}
		metas = append(metas, m)
	}
	return metas, rows.Err()
}

// SetPrerequisites replaces the prerequisite list for a subject.
func (s *Storage) SetPrerequisites(subjectID string, prereqs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM subject_prerequisites WHERE subject_id = ?`, subjectID); err != nil {
		return fmt.Errorf("clear prerequisites: %w", err)
	}
	for _, pid := range prereqs {
		if _, err := tx.Exec(`INSERT INTO subject_prerequisites (subject_id, prerequisite_id) VALUES (?, ?)`, subjectID, pid); err != nil {
			return fmt.Errorf("insert prerequisite %s: %w", pid, err)
		}
	}
	return tx.Commit()
}
